package model

import (
	"context"
	"time"
)

// LogEntry represents a single log message with its metadata
type LogEntry struct {
	// Timestamp when the log was created
	Timestamp time.Time
	// Severity level of the log (e.g., INFO, WARNING, ERROR)
	Severity string
	// Actual log message content
	Message string
}

// LogStreamer provides an interface for streaming logs
type LogStreamer interface {
	// StreamLogs starts streaming logs and returns a channel that will receive log entries.
	// The streaming continues until the context is cancelled.
	StreamLogs(ctx context.Context) chan LogEntry
}

// CloudProviderOptions contains configuration for cloud provider log streaming
type CloudProviderOptions struct {
	// GCP project identifier
	ProjectID string
	// Name of the service whose logs are being streamed
	ServiceName string
	// Region where the service is deployed
	Region string
}
