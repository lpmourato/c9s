package model

import (
	"context"
	"time"
)

// MockLogProvider implements LogProvider interface for testing
type MockLogProvider struct {
	ServiceName string
}

func (p *MockLogProvider) GetBaseFilter(serviceName string) string {
	return "mock-filter"
}

func (p *MockLogProvider) BuildFilter(baseFilter string, timestamp time.Time) string {
	return baseFilter
}

func (p *MockLogProvider) FetchLogs(ctx context.Context, filter string, batchSize int) ([]LogEntry, error) {
	return nil, nil
}
