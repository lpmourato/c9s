package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/interfaces"
	"github.com/lpmourato/c9s/internal/logging"
	"github.com/lpmourato/c9s/internal/model"
)

// NewMockLogView creates a new log view with mock data for testing
func NewMockLogView(app interfaces.UIController, serviceName, region string) tview.Primitive {
	ctx, cancel := context.WithCancel(context.Background())

	// Create mock provider
	mockProvider := &model.MockLogProvider{ServiceName: serviceName}

	opts := model.CloudProviderOptions{
		ServiceName: serviceName,
		Region:      region,
	}

	streamer := logging.NewLogService(mockProvider, opts)

	tv := tview.NewTextView()
	tv.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle(fmt.Sprintf(" %s - %s (MOCK) ", serviceName, region)).
		SetTitleAlign(tview.AlignLeft)

	// Show loading message
	loadingMsg := fmt.Sprintf("[gray]Starting mock log stream for [yellow::b]%s[gray]...\n\n",
		serviceName)
	fmt.Fprintf(tv, "%s", loadingMsg)

	// Set up key bindings
	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancel()           // Stop log streaming
			app.ReturnToMain() // Always return to main view
			return nil
		}
		return event
	})

	// Start streaming logs in background
	go func() {
		logChan := streamer.StreamLogs(ctx)
		for entry := range logChan {
			entry := entry // capture for closure
			app.QueueUpdateDraw(func() {
				timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
				level := entry.Severity

				// K9s-style coloring
				var levelColor string

				// Only treat DEFAULT level logs with ERROR as errors
				if level == "DEFAULT" && strings.Contains(entry.Message, "ERROR") {
					level = "ERROR"
				}

				// Now apply coloring based on final level determination
				switch level {
				case "ERROR", "CRITICAL", "FATAL":
					levelColor = "[red::b]"
				case "WARNING", "WARN":
					levelColor = "[yellow::b]"
				case "INFO":
					levelColor = "[green::b]"
				case "DEBUG":
					levelColor = "[gray::b]"
				case "TRACE":
					levelColor = "[blue::b]"
				default:
					levelColor = "[green::b]" // Use same color as INFO for normal DEFAULT messages
				}

				// Escape any existing color codes in the message
				message := strings.ReplaceAll(entry.Message, "[", "[[")

				// Format: gray timestamp, bold colored level, white message
				logLine := fmt.Sprintf("[gray]%s[-:-:-] %s%-7s[-:-:-] [white::b]%s[-:-:-]\n",
					timestamp,
					levelColor,
					level,
					message,
				)
				fmt.Fprintf(tv, "%s", logLine)

				// Auto-scroll to bottom
				tv.ScrollToEnd()
			})
		}
	}()

	return tv
}

// StartMockLogView creates and returns a mock log view
func StartMockLogView() error {
	app := NewApp()

	// Open mock log view
	app.SwitchToView(NewMockLogView(app, "test-service", "test-region"))

	// Run the application
	return app.Run()
}
