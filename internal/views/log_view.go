package views

import (
	"context"
	"fmt"
	"time"
	"github.com/derailed/tcell/v2"
	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
)

// LogView represents the log viewer for Cloud Run services
type LogView struct {
	*ui.TextView
	client *gcp.Client
	app    *ui.App
}

// NewLogView returns a new log viewer
func NewLogView(client *gcp.Client, app *ui.App) *LogView {
	view := &LogView{
		TextView: ui.NewTextView(),
		client:   client,
		app:      app,
	}

	// Setup key bindings directly on the primitive
	view.TextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages := app.GetPages()
			pages.RemovePage("logs")
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	return view
}

// ShowLogs displays logs for a given service
func (v *LogView) ShowLogs(ctx context.Context, serviceName string) error {
	v.Clear()
	v.SetTitle(fmt.Sprintf(" %s Logs (ESC to exit) ", serviceName))
	v.SetDynamicColors(true)
	v.SetScrollable(true)
	v.SetChangedFunc(func() {
		v.ScrollToEnd()
		v.app.Draw()
	})

	logStream, err := gcp.NewLogStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to create log stream: %v", err)
	}

	// Create a context that can be cancelled
	streamCtx, cancel := context.WithCancel(ctx)
	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cancel() // Cancel the log streaming
			pages := v.app.GetPages()
			pages.RemovePage("logs")
			pages.SwitchToPage("main")
			return nil
		}
		return event
	})

	// Set up log streaming with periodic refresh
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		// Initial fetch
		if err := logStream.StreamLogs(streamCtx, v.client.Project(), serviceName, v); err != nil && err != context.Canceled {
			v.Write([]byte(fmt.Sprintf("[red]Error loading logs: %v[white]\n", err)))
		}

		// Continuous refresh
		for {
			select {
			case <-streamCtx.Done():
				return
			case <-ticker.C:
				if err := logStream.StreamLogs(streamCtx, v.client.Project(), serviceName, v); err != nil && err != context.Canceled {
					v.Write([]byte(fmt.Sprintf("[red]Error refreshing logs: %v[white]\n", err)))
				}
			}
		}
	}()

	return nil
}

// Write implements io.Writer for the log view
func (v *LogView) Write(p []byte) (n int, err error) {
	v.TextView.Write(p)
	return len(p), nil
}
