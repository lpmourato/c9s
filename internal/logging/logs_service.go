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

// LogProvider defines the interface that cloud providers must implement
type LogProvider interface {
	FetchLogs(ctx context.Context, filter string, pageSize int) ([]model.LogEntry, error)
	BuildFilter(baseFilter string, timestamp time.Time) string
}

// TimeWindow represents a time window for fetching logs
type TimeWindow struct {
	Duration time.Duration
	Desc     string
}

// LogService provides a generic implementation for log streaming
type LogService struct {
	provider LogProvider
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
func (s *LogService) getInitialTimeWindows() []TimeWindow {
	return []TimeWindow{
		{10 * time.Minute, "last 10 minutes"},
		{1 * time.Hour, "last hour"},
		{24 * time.Hour, "last 24 hours"},
		{0, "all time"},
	}
}

// NewLogService creates a new log streaming service
func NewLogService(provider LogProvider, opts model.CloudProviderOptions) model.LogStreamer {
	return &LogService{
		provider:     provider,
		opts:         opts,
		buffer:       make([]model.LogEntry, 0, defaultBatchSize),
		flushTimeout: defaultFlushTimeout,
		batchTimer:   time.NewTimer(defaultFlushTimeout),
		pollInterval: minPollInterval,
	}
}

// GetBaseFilter returns the base filter for the log provider
func (s *LogService) GetBaseFilter() string {
	return fmt.Sprintf(`resource.type="cloud_run_revision" resource.labels.service_name="%s"`, s.opts.ServiceName)
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

// StreamLogs implements model.LogStreamer interface
func (s *LogService) StreamLogs(ctx context.Context) chan model.LogEntry {
	ch := make(chan model.LogEntry, defaultBatchSize)
	flushTicker := time.NewTicker(s.flushTimeout)
	// Ensure we use minPollInterval for initial polling
	s.pollInterval = minPollInterval
	pollTicker := time.NewTicker(s.pollInterval)

	go func() {
		defer close(ch)
		defer flushTicker.Stop()
		defer pollTicker.Stop()

		baseFilter := fmt.Sprintf(`resource.type="cloud_run_revision" resource.labels.service_name="%s"`, s.opts.ServiceName)

		for {
			select {
			case <-ctx.Done():
				// Final flush before exiting
				s.flushBuffer(ctx, ch)
				return
			case <-flushTicker.C:
				if !s.flushBuffer(ctx, ch) {
					return
				}
			case <-pollTicker.C:
				//var filter string
				var statusMsg model.LogEntry

				if !s.initialized {
					// Initial load - try different time windows until we find logs
					foundInitialLogs := false
					timeWindows := s.getInitialTimeWindows()

					for _, window := range timeWindows {
						statusMsg = model.LogEntry{
							Timestamp: time.Now(),
							Severity:  "INFO",
							Message:   fmt.Sprintf("Initial load: searching for logs from %s...", window.Desc),
						}
						if !s.addToBuffer(ctx, ch, statusMsg) {
							return
						}

						// Get initial logs
						logs, err := s.fetchLogs(ctx, baseFilter, window.Duration, true)
						if err != nil {
							continue
						}
						if len(logs) > 0 {
							foundInitialLogs = true
							// Add the logs to the buffer
							for _, entry := range logs {
								if !s.addToBuffer(ctx, ch, entry) {
									return
								}
							}
							if !s.flushBuffer(ctx, ch) {
								return
							}
							break
						}
					}

					if !foundInitialLogs {
						statusMsg = model.LogEntry{
							Timestamp: time.Now(),
							Severity:  "WARNING",
							Message:   "No logs found for this service",
						}
						if !s.addToBuffer(ctx, ch, statusMsg) {
							return
						}
						if !s.flushBuffer(ctx, ch) {
							return
						}
						continue
					}

					s.initialized = true
					// Set initial timestamp if not set
					if s.lastTimestamp.IsZero() {
						s.lastTimestamp = time.Now().Add(-time.Second) // Start from 1 second ago
					}
				}

				// Incremental updates - fetch only new logs
				logs, err := s.fetchLogs(ctx, baseFilter, 0, false)
				if err != nil {
					statusMsg = model.LogEntry{
						Timestamp: time.Now(),
						Severity:  "ERROR",
						Message:   fmt.Sprintf("Error fetching logs: %v", err),
					}
					if !s.addToBuffer(ctx, ch, statusMsg) {
						return
					}
					continue
				}

				// Process any new logs
				if len(logs) > 0 {
					for _, entry := range logs {
						if !s.addToBuffer(ctx, ch, entry) {
							return
						}
					}
					if !s.flushBuffer(ctx, ch) {
						return
					}
				}

				foundNewLogs := len(logs) > 0
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
