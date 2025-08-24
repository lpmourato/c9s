package model

import (
	"context"
	"time"
)

// LogProvider defines the interface that cloud providers must implement
type LogProvider interface {
	// FetchLogs retrieves logs from the provider's API
	FetchLogs(ctx context.Context, filter string, pageSize int) ([]LogEntry, error)
	// BuildFilter combines the base filter with a timestamp for incremental fetching
	BuildFilter(baseFilter string, timestamp time.Time) string
	// GetBaseFilter returns the provider-specific base filter format
	GetBaseFilter(serviceName string) string
}

// TimeWindow represents a time window for fetching logs
type TimeWindow struct {
	Duration time.Duration
	Desc     string
}
