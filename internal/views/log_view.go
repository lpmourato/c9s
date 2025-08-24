package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/logging"
	"github.com/lpmourato/c9s/internal/model"
	"github.com/lpmourato/c9s/internal/ui"
)

type LogView struct {
	*tview.TextView
	app         *ui.App
	serviceName string
	region      string
	ctx         context.Context
	cancel      context.CancelFunc
	streamer    model.LogStreamer
}

func NewLogView(app *ui.App, projectID, serviceName, region string) (*LogView, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Ensure cancel is called on error paths
	var v *LogView
	defer func() {
		if v == nil {
			cancel()
		}
	}()

	provider, err := logging.NewGCPLogService(projectID, serviceName, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create log streamer: %v", err)
	}

	opts := model.CloudProviderOptions{
		ProjectID:   projectID,
		ServiceName: serviceName,
		Region:      region,
	}

	streamer := logging.NewLogService(provider, opts)

	v = &LogView{
		TextView:    tview.NewTextView().SetDynamicColors(true),
		app:         app,
		serviceName: serviceName,
		region:      region,
		ctx:         ctx,
		cancel:      cancel,
		streamer:    streamer,
	}

	v.SetBorder(true)
	v.SetTitle(fmt.Sprintf(" %s - %s ", serviceName, region))
	v.SetTitleAlign(tview.AlignLeft)

	// Show loading message
	loadingMsg := fmt.Sprintf("Loading logs from %s in region %s...\n\n",
		serviceName, region)
	fmt.Fprint(v, loadingMsg)

	// Set up key bindings
	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.cancel()         // Stop log streaming
			app.ReturnToMain() // Always return to main view
			return nil
		}
		return event
	})

	return v, nil
}

// SetStreamer sets the log streamer
func (v *LogView) SetStreamer(streamer model.LogStreamer) {
	v.streamer = streamer
}

// StreamLogs starts streaming logs
func (v *LogView) StreamLogs() {
	v.streamLogs()
}

func (v *LogView) streamLogs() {
	logChan := v.streamer.StreamLogs(v.ctx)

	for entry := range logChan {
		v.app.QueueUpdateDraw(func() {
			timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
			level := strings.ToUpper(entry.Severity)
			fmt.Fprintf(v, "%s %-7s %s\n",
				timestamp,
				level,
				entry.Message,
			)

			// Auto-scroll to bottom
			v.ScrollToEnd()
		})
	}
}
