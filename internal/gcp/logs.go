package gcp

import (
	"context"
	"fmt"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/lpmourato/c9s/internal/model"
	"google.golang.org/api/iterator"
)

const (
	defaultFlushTimeout = 500 * time.Millisecond
	defaultBatchSize    = 100
	maxBufferSize       = 5000
	minPollInterval     = 10 * time.Second
	maxPollInterval     = 30 * time.Second
	pollBackoffFactor   = 1.5
)

// CloudRunLogStreamer streams logs from a Cloud Run service
type CloudRunLogStreamer struct {
	projectID   string
	serviceName string
	region      string
	client      *logging.Client

	// Buffering and batching
	buffer       []model.LogEntry
	batchTimer   *time.Timer
	flushTimeout time.Duration

	// Stream state
	lastTimestamp time.Time
	initialized   bool
	pollInterval  time.Duration
}

// NewCloudRunLogStreamer creates a new log streamer for a Cloud Run service
func NewCloudRunLogStreamer(projectID, serviceName, region string) (model.LogStreamer, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}

	return &CloudRunLogStreamer{
		projectID:    projectID,
		serviceName:  serviceName,
		region:       region,
		client:       client,
		buffer:       make([]model.LogEntry, 0, defaultBatchSize),
		flushTimeout: defaultFlushTimeout,
		batchTimer:   time.NewTimer(defaultFlushTimeout),
		pollInterval: minPollInterval,
	}, nil
}

// flushBuffer sends buffered logs to the channel
func (s *CloudRunLogStreamer) flushBuffer(ctx context.Context, ch chan<- model.LogEntry) bool {
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
func (s *CloudRunLogStreamer) addToBuffer(ctx context.Context, ch chan<- model.LogEntry, entry model.LogEntry) bool {
	s.buffer = append(s.buffer, entry)

	// Flush if buffer is full
	if len(s.buffer) >= defaultBatchSize {
		return s.flushBuffer(ctx, ch)
	}

	return true
}

// adjustPollInterval updates the polling interval based on whether new logs were found
func (s *CloudRunLogStreamer) adjustPollInterval(foundNewLogs bool) {
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
func (s *CloudRunLogStreamer) StreamLogs(ctx context.Context) chan model.LogEntry {
	ch := make(chan model.LogEntry, defaultBatchSize)
	flushTicker := time.NewTicker(s.flushTimeout)
	pollTicker := time.NewTicker(s.pollInterval)

	go func() {
		defer close(ch)
		defer s.client.Close()
		defer flushTicker.Stop()
		defer pollTicker.Stop()

		// For initial fetch only - try progressively larger windows
		timeWindows := []struct {
			duration time.Duration
			desc     string
		}{
			{10 * time.Minute, "last 10 minutes"},
			{1 * time.Hour, "last hour"},
			{24 * time.Hour, "last 24 hours"},
			{0, "all time"}, // 0 means no time filter
		}

		baseFilter := fmt.Sprintf(`resource.type="cloud_run_revision" resource.labels.service_name="%s"`,
			s.serviceName)

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
				var filter string
				var statusMsg model.LogEntry

				if !s.initialized {
					// Initial load - try different time windows until we find logs
					foundInitialLogs := false
					for _, window := range timeWindows {
						filter = baseFilter
						if window.duration > 0 {
							timestamp := time.Now().Add(-window.duration).Format(time.RFC3339)
							filter = fmt.Sprintf(`%s AND timestamp >= "%s"`, baseFilter, timestamp)
						}

						statusMsg = model.LogEntry{
							Timestamp: time.Now(),
							Severity:  "INFO",
							Message:   fmt.Sprintf("Initial load: searching for logs from %s...", window.desc),
						}
						if !s.addToBuffer(ctx, ch, statusMsg) {
							return
						}

						// Get initial logs
						if s.fetchLogs(ctx, ch, filter, true) {
							foundInitialLogs = true
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
						time.Sleep(time.Second * 2)
						continue
					}

					s.initialized = true
					// Set initial timestamp if not set
					if s.lastTimestamp.IsZero() {
						s.lastTimestamp = time.Now().Add(-time.Second) // Start from 1 second ago
					}
				}

				// Incremental updates - fetch only new logs
				// Using >= ensures we don't miss any logs with the exact same timestamp
				filter = fmt.Sprintf(`%s AND timestamp >= "%s"`, baseFilter, s.lastTimestamp.Add(time.Nanosecond).Format(time.RFC3339Nano))
				foundNewLogs := s.fetchLogs(ctx, ch, filter, false)

				// Adjust polling interval based on whether we found new logs
				s.adjustPollInterval(foundNewLogs)
				pollTicker.Reset(s.pollInterval)
			}
		}
	}()

	return ch
}

// fetchLogs retrieves logs using the given filter and updates lastTimestamp
func (s *CloudRunLogStreamer) fetchLogs(ctx context.Context, ch chan<- model.LogEntry, filter string, isInitialLoad bool) bool {
	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", s.projectID)},
		Filter:        filter,
		OrderBy:       "timestamp asc", // Always get logs in chronological order
		PageSize:      defaultBatchSize,
	}

	it := s.client.ListLogEntries(ctx, req)
	foundLogs := false
	var latestTimestamp time.Time
	var logs []model.LogEntry

	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			errMsg := model.LogEntry{
				Timestamp: time.Now(),
				Severity:  "ERROR",
				Message:   fmt.Sprintf("Error reading logs: %v", err),
			}
			s.addToBuffer(ctx, ch, errMsg)
			break
		}

		foundLogs = true
		timestamp := entry.GetTimestamp().AsTime()

		// Track the latest timestamp seen
		if latestTimestamp.IsZero() || timestamp.After(latestTimestamp) {
			latestTimestamp = timestamp
		}

		// Skip if we've already seen this log (can happen due to timestamp granularity)
		if !isInitialLoad && !timestamp.After(s.lastTimestamp) {
			continue
		}

		logEntry := model.LogEntry{
			Timestamp: timestamp,
			Severity:  entry.GetSeverity().String(),
			Message:   entry.GetTextPayload(),
		}
		logs = append(logs, logEntry)
	}

	// Send logs in chronological order
	for _, logEntry := range logs {
		if !s.addToBuffer(ctx, ch, logEntry) {
			return false
		}
	}

	// Force a flush after all logs
	if !s.flushBuffer(ctx, ch) {
		return false
	}

	// Update the last timestamp if we found logs
	if foundLogs && !latestTimestamp.IsZero() {
		s.lastTimestamp = latestTimestamp
	}

	return foundLogs
}
