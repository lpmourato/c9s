package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/lpmourato/c9s/internal/model"
)

const (
	defaultFlushTimeout = 500 * time.Millisecond
	defaultBatchSize    = 100
	minPollInterval     = 2 * time.Second
	maxPollInterval     = 30 * time.Second
	pollBackoffFactor   = 1.5
)

// LogService provides a generic implementation for log streaming
type LogService struct {
	provider model.LogProvider
	opts     model.CloudProviderOptions

	// Buffering and batching
	buffer       []model.LogEntry
	batchTimer   *time.Timer
	flushTimeout time.Duration

	// Stream state
	lastTimestamp time.Time
	initialized   bool
	pollInterval  time.Duration
}

// GetInitialTimeWindows returns the standard time windows for initial log fetching
func (s *LogService) getInitialTimeWindows() []model.TimeWindow {
	return []model.TimeWindow{
		{Duration: 10 * time.Minute, Desc: "last 10 minutes"},
		{Duration: 1 * time.Hour, Desc: "last hour"},
		{Duration: 24 * time.Hour, Desc: "last 24 hours"},
		{Duration: 0, Desc: "all time"},
	}
}

// NewLogService creates a new log streaming service
func NewLogService(provider model.LogProvider, opts model.CloudProviderOptions) model.LogStreamer {
	return &LogService{
		provider:     provider,
		opts:         opts,
		buffer:       make([]model.LogEntry, 0, defaultBatchSize),
		flushTimeout: defaultFlushTimeout,
		batchTimer:   time.NewTimer(defaultFlushTimeout),
		pollInterval: minPollInterval,
	}
}

// flushBuffer sends buffered logs to the channel
func (s *LogService) flushBuffer(ctx context.Context, ch chan<- model.LogEntry) bool {
	if len(s.buffer) == 0 {
		return true
	}

	for _, entry := range s.buffer {
		select {
		case ch <- entry:
		case <-ctx.Done():
			return false
		}
	}

	s.buffer = s.buffer[:0]
	return true
}

// addToBuffer adds a log entry to the buffer and flushes if needed
func (s *LogService) addToBuffer(ctx context.Context, ch chan<- model.LogEntry, entry model.LogEntry) bool {
	s.buffer = append(s.buffer, entry)

	// Flush if buffer is full
	if len(s.buffer) >= defaultBatchSize {
		return s.flushBuffer(ctx, ch)
	}

	return true
}

// adjustPollInterval updates the polling interval based on whether new logs were found
func (s *LogService) adjustPollInterval(foundNewLogs bool) {
	if foundNewLogs {
		// If we found logs, reset to minimum interval
		s.pollInterval = minPollInterval
	} else {
		// If no logs found, back off gradually up to max interval
		newInterval := time.Duration(float64(s.pollInterval) * pollBackoffFactor)
		if newInterval > maxPollInterval {
			newInterval = maxPollInterval
		}
		s.pollInterval = newInterval
	}
}

// addStatusMessage adds a status message to the log buffer
func (s *LogService) addStatusMessage(ctx context.Context, ch chan<- model.LogEntry, severity, message string) bool {
	statusMsg := model.LogEntry{
		Timestamp: time.Now(),
		Severity:  severity,
		Message:   message,
	}
	return s.addToBuffer(ctx, ch, statusMsg)
}

// processLogs adds multiple log entries to the buffer and flushes if needed
func (s *LogService) processLogs(ctx context.Context, ch chan<- model.LogEntry, logs []model.LogEntry) bool {
	for _, entry := range logs {
		if !s.addToBuffer(ctx, ch, entry) {
			return false
		}
	}
	return s.flushBuffer(ctx, ch)
}

// handleInitialLoad handles the initial loading of logs, trying different time windows
func (s *LogService) handleInitialLoad(ctx context.Context, ch chan<- model.LogEntry, baseFilter string) (bool, error) {
	timeWindows := s.getInitialTimeWindows()

	for _, window := range timeWindows {
		if !s.addStatusMessage(ctx, ch, "INFO", fmt.Sprintf("Initial load: searching for logs from %s...", window.Desc)) {
			return false, nil
		}

		logs, err := s.fetchLogs(ctx, baseFilter, window.Duration, true)
		if err != nil {
			continue
		}
		if len(logs) > 0 {
			if !s.processLogs(ctx, ch, logs) {
				return false, nil
			}
			return true, nil
		}
	}

	if !s.addStatusMessage(ctx, ch, "WARNING", "No logs found for this service") {
		return false, nil
	}
	return false, nil
}

// handleIncrementalUpdate handles fetching and processing new logs
func (s *LogService) handleIncrementalUpdate(ctx context.Context, ch chan<- model.LogEntry, baseFilter string) (bool, error) {
	logs, err := s.fetchLogs(ctx, baseFilter, 0, false)
	if err != nil {
		if !s.addStatusMessage(ctx, ch, "ERROR", fmt.Sprintf("Error fetching logs: %v", err)) {
			return false, nil
		}
		return false, err
	}

	if len(logs) > 0 {
		if !s.processLogs(ctx, ch, logs) {
			return false, nil
		}
	}

	return len(logs) > 0, nil
}

// StreamLogs implements model.LogStreamer interface
func (s *LogService) StreamLogs(ctx context.Context) chan model.LogEntry {
	ch := make(chan model.LogEntry, defaultBatchSize)
	flushTicker := time.NewTicker(s.flushTimeout)
	s.pollInterval = minPollInterval
	pollTicker := time.NewTicker(s.pollInterval)

	go func() {
		defer close(ch)
		defer flushTicker.Stop()
		defer pollTicker.Stop()

		baseFilter := s.provider.GetBaseFilter(s.opts.ServiceName)

		for {
			select {
			case <-ctx.Done():
				s.flushBuffer(ctx, ch)
				return
			case <-flushTicker.C:
				if !s.flushBuffer(ctx, ch) {
					return
				}
			case <-pollTicker.C:
				if !s.initialized {
					foundLogs, err := s.handleInitialLoad(ctx, ch, baseFilter)
					if err != nil || !foundLogs {
						continue
					}

					s.initialized = true
					if s.lastTimestamp.IsZero() {
						s.lastTimestamp = time.Now().Add(-time.Second)
					}
					continue
				}

				foundNewLogs, _ := s.handleIncrementalUpdate(ctx, ch, baseFilter)
				s.adjustPollInterval(foundNewLogs)
				pollTicker.Reset(s.pollInterval)
			}
		}
	}()

	return ch
}

// fetchLogs retrieves logs using the given filter and updates lastTimestamp
func (s *LogService) fetchLogs(ctx context.Context, baseFilter string, duration time.Duration, isInitialLoad bool) ([]model.LogEntry, error) {
	var filter string
	if duration > 0 {
		timestamp := time.Now().Add(-duration)
		filter = s.provider.BuildFilter(baseFilter, timestamp)
	} else if !isInitialLoad {
		filter = s.provider.BuildFilter(baseFilter, s.lastTimestamp.Add(time.Nanosecond))
	} else {
		filter = baseFilter
	}

	logs, err := s.provider.FetchLogs(ctx, filter, defaultBatchSize)
	if err != nil {
		return nil, err
	}

	// Process logs and update timestamp
	var latestTimestamp time.Time
	var filteredLogs []model.LogEntry

	for _, log := range logs {
		// Track the latest timestamp seen
		if latestTimestamp.IsZero() || log.Timestamp.After(latestTimestamp) {
			latestTimestamp = log.Timestamp
		}

		// Skip if we've already seen this log (can happen due to timestamp granularity)
		if !isInitialLoad && !log.Timestamp.After(s.lastTimestamp) {
			continue
		}

		filteredLogs = append(filteredLogs, log)
	}

	// Update the last timestamp if we found logs
	if len(filteredLogs) > 0 && !latestTimestamp.IsZero() {
		s.lastTimestamp = latestTimestamp
	}

	return filteredLogs, nil
}
