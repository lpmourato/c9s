package mock

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/lpmourato/c9s/internal/model"
)

// LogStreamer implements model.LogStreamer for testing
type LogStreamer struct {
	serviceName string
}

// NewLogStreamer creates a new mock log streamer for testing
func NewLogStreamer(serviceName string) *LogStreamer {
	return &LogStreamer{
		serviceName: serviceName,
	}
}

// StreamLogs implements model.LogStreamer interface for testing purposes
func (m *LogStreamer) StreamLogs(ctx context.Context) chan model.LogEntry {
	ch := make(chan model.LogEntry)
	// Test different log levels
	logTypes := []struct {
		severity string
		messages []string
	}{
		{
			severity: "ERROR",
			messages: []string{
				"Failed to connect to database",
				"Invalid configuration detected",
				"Service crashed unexpectedly",
			},
		},
		{
			severity: "WARNING",
			messages: []string{
				"High memory usage detected",
				"Retrying failed request",
				"ERROR: Operation completed with warnings", // WARNING log containing ERROR
			},
		},
		{
			severity: "WARN",
			messages: []string{
				"Database connection slow",
				"Low disk space detected",
				"ERROR: Task completed with warnings", // WARN log containing ERROR
			},
		},

		{
			severity: "INFO",
			messages: []string{
				"Service started successfully",
				"Request processed",
				"Cache refreshed",
			},
		},
		{
			severity: "DEBUG",
			messages: []string{
				"Connection pool stats: active=5",
				"Cache hit ratio: 85%",
				"Request headers received",
			},
		},
		{
			severity: "", // Empty severity in GCP is equivalent to INFO
			messages: []string{
				"System status check completed",
				"Routine maintenance running",
				"Backup completed successfully",
			},
		},
	}

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond * time.Duration(500+rand.Intn(1500))):
				// Select a random log type
				logType := logTypes[rand.Intn(len(logTypes))]

				// Select a random message for this severity
				msg := logType.messages[rand.Intn(len(logType.messages))]

				entry := model.LogEntry{
					Timestamp: time.Now(),
					Severity:  logType.severity,
					Message:   fmt.Sprintf("[%s] %s", m.serviceName, msg),
				}

				ch <- entry
			}
		}
	}()

	return ch
}
