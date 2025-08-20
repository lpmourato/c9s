package views

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/gcp"
	"github.com/lpmourato/c9s/internal/ui"
)

// CloudRunView represents the Cloud Run services view
type CloudRunView struct {
	*ui.Table
	client  *gcp.Client
	app     *ui.App
	header  *tview.TextView
	project string
	region  string
}

// NewCloudRunView returns a new Cloud Run view
func NewCloudRunView(ctx context.Context, client *gcp.Client, app *ui.App) *CloudRunView {
	table := ui.NewTable()
	table.SetApp(app)
	table.SetSelectable(true, false)

	// Make sure table can receive keyboard focus
	table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick {
			app.SetFocus(table)
		}
		return action, event
	})

	// Create header for session info with k9s style
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true).
		SetTextAlign(tview.AlignLeft)

	// Set k9s-like styling with black background
	header.SetBackgroundColor(tcell.ColorBlack)

	// Create main flex layout
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(tcell.ColorDefault)

	// Add components with more height for header
	flex.AddItem(header, 4, 0, false) // Increased header height
	flex.AddItem(table, 0, 1, true)

	view := &CloudRunView{
		Table:   table,
		client:  client,
		app:     app,
		header:  header,
		project: client.Project(),
		region:  client.Region(),
	}

	// Set up the table
	view.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
	view.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorNavy))

	// Show loading message
	loadingCell := ui.NewTableCell("Loading Cloud Run services...").
		SetTextColor(tcell.ColorYellow).
		SetExpansion(1).
		SetAlign(tview.AlignCenter)
	view.SetCell(1, 0, loadingCell)
	view.Select(1, 0)

	// Set up input handlers at the table level
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlD:
			view.showServiceDescription()
			return nil
		case tcell.KeyEnter:
			view.showLogs()
			return nil
		case tcell.KeyCtrlR:
			_ = view.Refresh(ctx)
			return nil
		}
		// Let all other keys be handled by the table (including Up/Down)
		return event
	})

	// Update header immediately
	view.updateHeader()

	// Set the flex layout as the root primitive and give the table focus
	app.SetRoot(flex, true)
	app.SetFocus(view)

	return view
}

// Refresh updates the view with latest Cloud Run services
func (v *CloudRunView) Refresh(ctx context.Context) error {
	// Create a context with reasonable timeout
	refreshCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get all services in one call
	services, err := v.client.ListServices(refreshCtx)
	if err != nil {
		return fmt.Errorf("failed to list services: %v", err)
	}

	// Remember current selection
	currentRow, currentCol := v.GetSelection()

	// Prepare all updates before queueing them
	updates := func() {
		// Clear and set up table
		v.Clear()
		v.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
		v.SetTitle(fmt.Sprintf(" Cloud Run Services (%d) ", len(services)))

		// Convert all services to summaries and sort by name
		summaries := make([]*gcp.ServiceSummary, len(services))
		for i, svc := range services {
			summaries[i] = gcp.ToSummary(svc)
		}
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].Name < summaries[j].Name
		})

		// Update all rows with sorted data
		for row, summary := range summaries {
			v.updateServiceRow(row+1, summary)
		}

		// Restore selection
		if currentRow > 0 && currentRow <= len(services) {
			v.Select(currentRow, currentCol)
		} else {
			v.Select(1, 0)
		}
	}

	// Queue updates
	v.app.QueueUpdateDraw(func() {
		updates()
		v.updateHeader() // Refresh header too
	})

	return nil
}

// StartRefreshTimer starts the periodic refresh timer
func (v *CloudRunView) StartRefreshTimer(ctx context.Context) {
	// Prepare initial state update
	initialState := func() {
		v.Clear()
		v.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
		v.SetTitle(" Cloud Run Services (Loading...) ")
	}

	// Queue initial state update
	v.app.QueueUpdateDraw(initialState)

	// Start refresh loop in background
	go func() {
		// Perform initial refresh
		if err := v.Refresh(ctx); err != nil {
			v.app.QueueUpdateDraw(func() {
				v.SetTitle(fmt.Sprintf(" Cloud Run Services (Error: %v) ", err))
			})
		}

		// Set up regular refresh interval
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Don't start a new refresh if context is cancelled
				select {
				case <-ctx.Done():
					return
				default:
					if err := v.Refresh(ctx); err != nil {
						v.app.QueueUpdateDraw(func() {
							v.SetTitle(fmt.Sprintf(" Cloud Run Services (Error: %v) ", err))
						})
					}
				}
			}
		}
	}()
}

// updateServiceRow updates a single row in the table with service data
func (v *CloudRunView) updateServiceRow(row int, summary *gcp.ServiceSummary) {
	// Basic service info with left alignment
	v.SetCell(row, 0, ui.NewTableCell(summary.Name).
		SetExpansion(1).
		SetAlign(tview.AlignLeft))

	v.SetCell(row, 1, ui.NewTableCell(summary.Region).
		SetExpansion(1).
		SetAlign(tview.AlignLeft))

	v.SetCell(row, 2, ui.NewTableCell(summary.URL).
		SetExpansion(2). // URLs need more space
		SetAlign(tview.AlignLeft))

	// Update status with detailed state information
	var statusComponents []string
	statusComponents = append(statusComponents, summary.Status)

	// Add instance and scaling info
	if summary.IsServing {
		if summary.ActualInstance > 0 {
			statusComponents = append(statusComponents,
				fmt.Sprintf("Running:%d", summary.ActualInstance))
			if summary.MinInstances > 0 {
				statusComponents = append(statusComponents,
					fmt.Sprintf("Min:%d", summary.MinInstances))
			}
		} else if summary.IsScaledToZero {
			statusComponents = append(statusComponents, "Scaled:0")
		}
	} else {
		reason := "Not Serving"
		if summary.ServingMessage != "" {
			reason = summary.ServingMessage
			if len(reason) > 30 {
				reason = reason[:27] + "..."
			}
		}
		statusComponents = append(statusComponents, reason)
	}

	// Add condition states
	statusComponents = append(statusComponents,
		fmt.Sprintf("[R:%s S:%s]",
			summary.RawReadyState,
			summary.RawServingState))

	// Add revision if available
	if rev := summary.Revision; rev != "" {
		shortRev := rev
		if parts := strings.Split(rev, "/"); len(parts) > 0 {
			shortRev = parts[len(parts)-1]
		}
		if strings.HasPrefix(shortRev, summary.Name+"-") {
			shortRev = strings.TrimPrefix(shortRev, summary.Name+"-")
		}
		statusComponents = append(statusComponents, fmt.Sprintf("rev:%s", shortRev))
	}

	statusText := strings.Join(statusComponents, " ")
	statusCell := ui.NewTableCell(statusText).
		SetExpansion(1).
		SetAlign(tview.AlignLeft)

	switch {
	case strings.HasPrefix(summary.Status, "Ready"):
		statusCell.SetTextColor(tcell.ColorGreen)
	case strings.HasPrefix(summary.Status, "Failed"):
		statusCell.SetTextColor(tcell.ColorRed)
	case summary.Status == "Stopped":
		statusCell.SetTextColor(tcell.ColorYellow)
	case strings.HasPrefix(summary.Status, "Updating"):
		statusCell.SetTextColor(tcell.ColorYellow)
	case strings.HasPrefix(summary.Status, "Creating"):
		statusCell.SetTextColor(tcell.ColorYellow)
	case strings.HasPrefix(summary.Status, "Pending"):
		statusCell.SetTextColor(tcell.ColorYellow)
	default:
		statusCell.SetTextColor(tcell.ColorGray)
	}
	v.SetCell(row, 3, statusCell)

	// Format and update last deploy time with right alignment
	lastDeploy := summary.LastDeployTime
	if t, err := time.Parse(time.RFC3339, lastDeploy); err == nil {
		lastDeploy = t.Local().Format("2006-01-02 15:04:05")
	}
	v.SetCell(row, 4, ui.NewTableCell(lastDeploy).
		SetExpansion(1).
		SetAlign(tview.AlignRight))

	// Update traffic information with detailed formatting
	var traffic string
	var trafficColor tcell.Color
	switch {
	case summary.Status == "Stopped":
		traffic = "No traffic (stopped)"
		trafficColor = tcell.ColorYellow
	case strings.HasPrefix(summary.Status, "Failed"):
		traffic = "No traffic (failed)"
		trafficColor = tcell.ColorRed
	case len(summary.Traffic) == 0:
		traffic = "No traffic"
		trafficColor = tcell.ColorGray
	default:
		trafficParts := make([]string, 0, len(summary.Traffic))
		for _, t := range summary.Traffic {
			if t.Percent > 0 {
				trafficParts = append(trafficParts, fmt.Sprintf("%s (%d%%)", t.RevisionName, t.Percent))
			}
		}
		if len(trafficParts) > 0 {
			traffic = strings.Join(trafficParts, ", ")
			trafficColor = tcell.ColorGreen
		} else {
			traffic = "No traffic"
			trafficColor = tcell.ColorGray
		}
	}

	v.SetCell(row, 5, ui.NewTableCell(traffic).
		SetExpansion(2). // Give more space for traffic splits
		SetAlign(tview.AlignLeft).
		SetTextColor(trafficColor))
}

// updateHeader updates the header content with session info
func (v *CloudRunView) updateHeader() {
	// Format similar to k9s context line
	// Use tab spacing for better alignment
	const separator = "│"
	headerText := fmt.Sprintf(
		"\n "+ // Start with a newline for spacing
			"[white::b]Context[white:-:-]: [white::b]GCP[-:-:-] %s\n"+
			"[white::b]Project[white:-:-]: [white::b]%s[-:-:-] %s\n"+
			"[white::b]Region[white:-:-]:  [white::b]%s[-:-:-] %s    "+
			"[gray::-]Press ? for help[-:-:-]",
		separator,
		v.project, separator,
		v.region, separator)
	v.header.SetText(headerText)
}

// showServiceDescription displays detailed information about the selected service
func (v *CloudRunView) showServiceDescription() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	// TODO: Implement service description view
	v.app.QueueUpdateDraw(func() {
		v.SetTitle(fmt.Sprintf(" Cloud Run Services - Describing %s... ", serviceName))
	})
}

func (v *CloudRunView) showLogs() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	logView := NewLogView(v.client, v.app)

	pages := v.app.GetPages()
	pages.AddPage("logs", logView, true, true)

	if err := logView.ShowLogs(context.Background(), serviceName); err != nil {
		pages.RemovePage("logs")
		return
	}
}
