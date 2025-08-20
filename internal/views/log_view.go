package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
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

func NewLogView(app *ui.App, serviceName, region string) *LogView {
	ctx, cancel := context.WithCancel(context.Background())

	v := &LogView{
		TextView:    tview.NewTextView().SetDynamicColors(true),
		app:         app,
		serviceName: serviceName,
		region:      region,
		ctx:         ctx,
		cancel:      cancel,
		streamer:    model.NewMockLogStreamer(serviceName),
	}

	v.SetBorder(true)
	v.SetTitle(fmt.Sprintf(" %s - %s ", serviceName, region))
	v.SetTitleAlign(tview.AlignLeft)

	// Set up key bindings
	v.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			v.cancel()         // Stop log streaming
			app.ReturnToMain() // Always return to main view
			return nil
		}
		return event
	})

	// Start streaming logs
	go v.streamLogs()

	return v
}

func (v *LogView) streamLogs() {
	logChan := v.streamer.StreamLogs(v.ctx)

	for entry := range logChan {
		v.app.QueueUpdateDraw(func() {
			color := "[white]"
			switch strings.ToUpper(entry.Severity) {
			case "ERROR":
				color = "[red]"
			case "WARNING":
				color = "[yellow]"
			case "INFO":
				color = "[white]"
			case "DEBUG":
				color = "[gray]"
			}

			timestamp := entry.Timestamp.Format("15:04:05.000")
			fmt.Fprintf(v, "%s %s%s %s[white]\n",
				timestamp,
				color,
				entry.Severity,
				entry.Message,
			)

			// Auto-scroll to bottom
			v.ScrollToEnd()
		})
	}
}
