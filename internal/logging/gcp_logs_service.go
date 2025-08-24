package logging

import (
	"context"
	"fmt"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/lpmourato/c9s/internal/model"
	"google.golang.org/api/iterator"
)

// GCPLogProvider handles log streaming for GCP Cloud Run
type GCPLogProvider struct {
	client    *logging.Client
	projectID string
}

// NewGCPLogService creates a new log streaming service for GCP Cloud Run
func NewGCPLogService(projectID, serviceName, region string) (model.LogProvider, error) {
	// Initialize GCP client
	ctx := context.Background()
	client, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}

	return &GCPLogProvider{
		client:    client,
		projectID: projectID,
	}, nil
}

/*
FetchLogs implements LogProvider.FetchLogs

This implementation is specific to GCP because:
1. Uses GCP-specific client and request types (loggingpb.ListLogEntriesRequest)
2. Requires GCP-specific resource naming format ("projects/%s")
3. Uses GCP-specific query parameters (OrderBy, PageSize)
4. Handles GCP-specific log entry format conversion to our model.LogEntry
If we were to implement this for a different provider (AWS, Azure), they would
use their own SDK clients and request/response formats
*/
func (p *GCPLogProvider) FetchLogs(ctx context.Context, filter string, pageSize int) ([]model.LogEntry, error) {
	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", p.projectID)},
		Filter:        filter,
		OrderBy:       "timestamp asc",
		PageSize:      int32(pageSize),
	}

	it := p.client.ListLogEntries(ctx, req)
	var logs []model.LogEntry

	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		logs = append(logs, model.LogEntry{
			Timestamp: entry.GetTimestamp().AsTime(),
			Severity:  entry.GetSeverity().String(),
			Message:   entry.GetTextPayload(),
		})
	}

	return logs, nil
}

// BuildFilter implements LogProvider.BuildFilter
func (p *GCPLogProvider) BuildFilter(baseFilter string, timestamp time.Time) string {
	if timestamp.IsZero() {
		return baseFilter
	}
	return fmt.Sprintf(`%s AND timestamp >= "%s"`, baseFilter, timestamp.Format(time.RFC3339Nano))
}

// GetBaseFilter implements LogProvider.GetBaseFilter
// Returns GCP Cloud Run specific filter format
func (p *GCPLogProvider) GetBaseFilter(serviceName string) string {
	return fmt.Sprintf(`resource.type="cloud_run_revision" resource.labels.service_name="%s"`, serviceName)
}

// Close closes the GCP logging client
func (p *GCPLogProvider) Close() error {
	return p.client.Close()
}
