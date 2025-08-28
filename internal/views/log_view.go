package views

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

type LogView struct {
	*tview.TextView
	app         interfaces.UIController
	serviceName string
	region      string
	ctx         context.Context
	cancel      context.CancelFunc
	streamer    model.LogStreamer
	topMessage  string
}

func NewLogView(app interfaces.UIController, projectID, serviceName, region string) (*LogView, error) {
	provider, err := logging.NewGCPLogService(projectID, serviceName, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create log streamer: %v", err)
	}

	return NewLogViewWithProvider(app, provider, projectID, serviceName, region)
}

// NewMockLogView creates a new log view with mock data for testing
func NewMockLogView(app interfaces.UIController, serviceName, region string) *LogView {
	ctx, cancel := context.WithCancel(context.Background())

	// Create mock provider
	mockProvider := &model.MockLogProvider{ServiceName: serviceName}

	opts := model.CloudProviderOptions{
		ServiceName: serviceName,
		Region:      region,
	}

	streamer := logging.NewLogService(mockProvider, opts)

	v := &LogView{
		TextView:    tview.NewTextView().SetDynamicColors(true),
		app:         app,
		serviceName: serviceName,
		region:      region,
		ctx:         ctx,
		cancel:      cancel,
		streamer:    streamer,
	}

	v.SetBorder(true)
	v.SetTitle(fmt.Sprintf(" %s - %s (MOCK) ", serviceName, region))
	v.SetTitleAlign(tview.AlignLeft)

	// Show loading message (for mock view)
	loadingMsg := fmt.Sprintf("[gray::b]Starting mock log stream for [yellow::b]%s[gray::b]...\n\n",
		serviceName)
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

	return v
}

// NewLogViewWithProvider creates a new log view with the specified provider
func NewLogViewWithProvider(app interfaces.UIController, provider model.LogProvider, projectID, serviceName, region string) (*LogView, error) {
	ctx, cancel := context.WithCancel(context.Background())

	opts := model.CloudProviderOptions{
		ProjectID:   projectID,
		ServiceName: serviceName,
		Region:      region,
	}

	streamer := logging.NewLogService(provider, opts)

	v := &LogView{
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
	loadingMsg := fmt.Sprintf("[gray::b]Loading logs from [yellow::b]%s[gray::b] in region [yellow::b]%s[gray::b]...\n\n",
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
		entry := entry // capture for goroutine
		v.app.QueueUpdateDraw(func() {
			timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
			message := entry.Message

			// If this is an initial status message, set as topMessage
			if strings.HasPrefix(message, "Initial load: searching for logs from") {
				v.topMessage = fmt.Sprintf("[gray::b]%s[-:-:-]\n", message)
				v.SetText(v.topMessage)
				v.ScrollToBeginning()
				return
			}

			// Parse level from the message content when severity is DEFAULT
			level := "INFO" // default level
			message = strings.TrimSpace(message)

			// Check for common log level patterns in the message
			switch {
			case strings.Contains(message, "ERROR"):
				level = "ERROR"
			case strings.Contains(message, "WARN") || strings.Contains(message, "WARNING"):
				level = "WARN"
			case strings.Contains(message, "INFO"):
				level = "INFO"
			case strings.Contains(message, "DEBUG"):
				level = "DEBUG"
			case strings.Contains(message, "TRACE"):
				level = "TRACE"
			}

			message = strings.TrimSpace(message)

			// K9s-style coloring
			var levelColor string

			// Apply coloring based on level
			switch level {
			case "ERROR", "CRITICAL", "FATAL":
				levelColor = "[red::b]"
			case "WARN":
				levelColor = "[yellow::b]"
			case "INFO":
				levelColor = "[green::b]"
			case "DEBUG":
				levelColor = "[gray::b]"
			case "TRACE":
				levelColor = "[blue::b]"
			default:
				levelColor = "[green::b]" // Use INFO color for any unrecognized level
			}

			// Escape any existing color codes in the message
			message = strings.ReplaceAll(message, "[", "[[")

			// Format: gray timestamp, bold colored level, white message
			logLine := fmt.Sprintf("[gray::b]%s[-:-:-] %s%-7s[-:-:-] [white::b]%s[-:-:-]\n",
				timestamp,
				levelColor,
				level,
				message,
			)

			// Get current content and append new log line, always keeping topMessage at the top
			currentContent := v.GetText(false)
			if v.topMessage != "" {
				// Remove topMessage from currentContent if present
				currentContent = strings.TrimPrefix(currentContent, v.topMessage)
				newContent := v.topMessage + currentContent + logLine
				v.SetText(newContent)
			} else {
				v.SetText(currentContent + logLine)
			}

			// Keep the view scrolled to end to show latest logs
			v.ScrollToEnd()
		})
	}
}
