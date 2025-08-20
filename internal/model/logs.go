package model

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Severity  string
	Message   string
}

// LogStreamer provides an interface for streaming logs
type LogStreamer interface {
	StreamLogs(ctx context.Context) chan LogEntry
}

// MockLogStreamer implements LogStreamer for testing
type MockLogStreamer struct {
	serviceName string
}

func NewMockLogStreamer(serviceName string) *MockLogStreamer {
	return &MockLogStreamer{
		serviceName: serviceName,
	}
}

func (m *MockLogStreamer) StreamLogs(ctx context.Context) chan LogEntry {
	ch := make(chan LogEntry)
	severities := []string{"INFO", "WARNING", "ERROR", "DEBUG"}
	messages := []string{
		"Request processed successfully",
		"Connection attempt failed",
		"Cache hit ratio: 85%%",
		"Memory usage: %dMB",
		"Processing request from %s",
		"Response time: %dms",
		"Background task completed",
		"Starting scheduled job",
	}

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond * time.Duration(500+rand.Intn(1500))):
				severity := severities[rand.Intn(len(severities))]
				msgTemplate := messages[rand.Intn(len(messages))]
				var msg string

				switch msgTemplate {
				case "Memory usage: %dMB":
					msg = fmt.Sprintf(msgTemplate, 100+rand.Intn(900))
				case "Processing request from %s":
					regions := []string{"us-east1", "europe-west4", "asia-east1"}
					msg = fmt.Sprintf(msgTemplate, regions[rand.Intn(len(regions))])
				case "Response time: %dms":
					msg = fmt.Sprintf(msgTemplate, 10+rand.Intn(990))
				default:
					msg = msgTemplate
				}

				entry := LogEntry{
					Timestamp: time.Now(),
					Severity:  severity,
					Message:   fmt.Sprintf("[%s] %s", m.serviceName, msg),
				}

				ch <- entry
			}
		}
	}()

	return ch
}
