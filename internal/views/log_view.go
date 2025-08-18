package views

import (
	"context"
	"fmt"

	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
)

// LogView represents the log viewer for Cloud Run services
type LogView struct {
	*ui.TextView
	client *gcp.Client
}

// NewLogView returns a new log viewer
func NewLogView(client *gcp.Client) *LogView {
	view := &LogView{
		TextView: ui.NewTextView(),
		client:   client,
	}

	return view
}

// ShowLogs displays logs for a given service
func (v *LogView) ShowLogs(ctx context.Context, serviceName string) error {
	v.Clear()
	v.SetTitle(fmt.Sprintf(" %s Logs ", serviceName))

	logStream, err := gcp.NewLogStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create log stream: %v", err)
	}

	go func() {
		if err := logStream.StreamLogs(ctx, v.client.Project(), serviceName, v); err != nil {
			v.Write([]byte(fmt.Sprintf("Error streaming logs: %v\n", err)))
		}
	}()

	return nil
}

// Write implements io.Writer for the log view
func (v *LogView) Write(p []byte) (n int, err error) {
	v.TextView.Write(p)
	return len(p), nil
}
