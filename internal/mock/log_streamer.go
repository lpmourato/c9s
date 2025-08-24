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
					regions := []string{"us-east", "eu-west", "ap-east"}
					msg = fmt.Sprintf(msgTemplate, regions[rand.Intn(len(regions))])
				case "Response time: %dms":
					msg = fmt.Sprintf(msgTemplate, 10+rand.Intn(990))
				default:
					msg = msgTemplate
				}

				entry := model.LogEntry{
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
