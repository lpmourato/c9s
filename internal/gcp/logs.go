package gcp

import (
	"context"
	"fmt"
	"io"

	logging "cloud.google.com/go/logging/apiv2"
	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
)

// LogStream handles streaming logs from Cloud Run services
type LogStream struct {
	client *logging.Client
}

// NewLogStream creates a new log streaming client
func NewLogStream(ctx context.Context) (*LogStream, error) {
	client, err := logging.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging client: %v", err)
	}

	return &LogStream{
		client: client,
	}, nil
}

// StreamLogs streams logs for a given Cloud Run service
func (ls *LogStream) StreamLogs(ctx context.Context, project, service string, w io.Writer) error {
	filter := fmt.Sprintf(`resource.type="cloud_run_revision" AND resource.labels.service_name="%s"`, service)
	
	req := &loggingpb.TailLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", project)},
		Filter:       filter,
		BufferWindow: &durationpb.Duration{Seconds: 20}, // Buffer 20s of logs
	}

	stream, err := ls.client.TailLogEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to start log stream: %v", err)
	}

	if err := stream.Send(req); err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error receiving logs: %v", err)
		}

		for _, entry := range resp.Entries {
			var payload string
			switch {
			case entry.GetTextPayload() != "":
				payload = entry.GetTextPayload()
			case entry.GetJsonPayload() != nil:
				payload = entry.GetJsonPayload().String()
			case entry.GetProtoPayload() != nil:
				payload = entry.GetProtoPayload().String()
			default:
				continue
			}

			severity := entry.GetSeverity().String()
			fmt.Fprintf(w, "[%s] [%s] %s\n", 
				entry.GetTimestamp().AsTime().Format("2006-01-02 15:04:05"),
				severity,
				payload)
		}
		
		// Force a redraw of the view
		if flusher, ok := w.(interface{ Flush() }); ok {
			flusher.Flush()
		}
	}
}
