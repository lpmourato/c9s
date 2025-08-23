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

// CloudRunLogStreamer streams logs from a Cloud Run service
type CloudRunLogStreamer struct {
	projectID   string
	serviceName string
	region      string
	client      *logging.Client
}

// NewCloudRunLogStreamer creates a new log streamer for a Cloud Run service
func NewCloudRunLogStreamer(projectID, serviceName, region string) (model.LogStreamer, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}

	return &CloudRunLogStreamer{
		projectID:   projectID,
		serviceName: serviceName,
		region:      region,
		client:      client,
	}, nil
}

// StreamLogs implements model.LogStreamer interface
func (s *CloudRunLogStreamer) StreamLogs(ctx context.Context) chan model.LogEntry {
	ch := make(chan model.LogEntry)

	go func() {
		defer close(ch)
		defer s.client.Close()

		// Basic filter for Cloud Run logs
		filter := fmt.Sprintf(`resource.type="cloud_run_revision" resource.labels.service_name="%s"`,
			s.serviceName)

		// Create initial request
		req := &loggingpb.ListLogEntriesRequest{
			ResourceNames: []string{fmt.Sprintf("projects/%s", s.projectID)},
			Filter:        filter,
			OrderBy:       "timestamp desc",
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Get logs
				it := s.client.ListLogEntries(ctx, req)
				for {
					entry, err := it.Next()
					if err == iterator.Done {
						break
					}
					if err != nil {
						fmt.Printf("Error reading logs: %v\n", err)
						break
					}

					logEntry := model.LogEntry{
						Timestamp: entry.GetTimestamp().AsTime(),
						Severity:  entry.GetSeverity().String(),
						Message:   entry.GetTextPayload(),
					}

					select {
					case ch <- logEntry:
					case <-ctx.Done():
						return
					}
				}

				// Simple polling interval
				time.Sleep(time.Second * 2)
			}
		}
	}()

	return ch
}
