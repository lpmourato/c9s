package views

import (
	"fmt"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/model"
	"github.com/lpmourato/c9s/internal/ui"
)

// CloudRunView represents the Cloud Run services view
type CloudRunView struct {
	*ui.Table
	app         *ui.App
	headerTable *tview.Table
	project     string
	region      string
}

// NewCloudRunView returns a new Cloud Run view
func NewCloudRunView(app *ui.App) *CloudRunView {
	table := ui.NewTable()
	table.SetApp(app)
	table.SetSelectable(true, false)

	// Create header table for session info
	headerTable := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)

	// Set styling similar to services table
	headerTable.SetBackgroundColor(tcell.ColorBlack)
	headerTable.SetBorder(true)
	headerTable.SetBorderColor(tcell.ColorGray)
	headerTable.SetTitle(" Cloud Run Context ")

	// Remove selection highlighting
	headerTable.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)) // Create main flex layout

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(tcell.ColorDefault)

	// Add components with more height for header
	flex.AddItem(headerTable, 5, 0, false) // Header height for 2 rows + shortcuts
	flex.AddItem(table, 0, 1, true)        // Table takes remaining space

	view := &CloudRunView{
		Table:       table,
		app:         app,
		headerTable: headerTable,
		project:     "dev-tla-cm",
		region:      "europe-west4",
	}

	// Set up the table
	view.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
	view.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorNavy))

	// Set up input handlers at the table level
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlD:
			view.showServiceDescription()
			return nil
		case tcell.KeyEnter:
			view.showLogs()
			return nil
		}
		return event
	})

	// Update header immediately
	view.updateHeader()

	// Set the flex layout as the main view and give the table focus
	app.SetMainView(flex)
	app.SetFocus(view)

	// Load mock data
	view.loadMockData()

	return view
}

// loadMockData loads mock service data into the table
func (v *CloudRunView) loadMockData() {
	services := model.GetMockServices()
	for i, svc := range services {
		row := i + 1 // Skip header row
		v.updateServiceRow(row, svc)
	}
	v.Select(1, 0) // Select first row
}

// updateServiceRow updates a single row in the table with service data
func (v *CloudRunView) updateServiceRow(row int, svc model.Service) {
	// Basic service info
	v.SetCell(row, 0, ui.NewTableCell(svc.GetName()).
		SetExpansion(1))

	v.SetCell(row, 1, ui.NewTableCell(svc.GetRegion()).
		SetExpansion(1))

	v.SetCell(row, 2, ui.NewTableCell(svc.GetURL()).
		SetExpansion(2))

	// Status with color coding
	statusCell := ui.NewTableCell(svc.GetStatus()).
		SetExpansion(1)

	switch svc.GetStatus() {
	case "Ready":
		statusCell.SetTextColor(tcell.ColorGreen)
	case "Failed":
		statusCell.SetTextColor(tcell.ColorRed)
	case "Updating":
		statusCell.SetTextColor(tcell.ColorYellow)
	default:
		statusCell.SetTextColor(tcell.ColorGray)
	}
	v.SetCell(row, 3, statusCell)

	// Last deploy time
	lastDeploy := svc.GetLastDeploy().Format("2006-01-02 15:04:05")
	v.SetCell(row, 4, ui.NewTableCell(lastDeploy).
		SetExpansion(1).
		SetAlign(tview.AlignRight))

	// Traffic with color coding
	trafficCell := ui.NewTableCell(svc.GetTraffic()).
		SetExpansion(2)

	switch {
	case svc.GetTraffic() == "No traffic (failed)":
		trafficCell.SetTextColor(tcell.ColorRed)
	case svc.GetTraffic() == "No traffic (stopped)":
		trafficCell.SetTextColor(tcell.ColorYellow)
	case svc.GetTraffic() == "No traffic":
		trafficCell.SetTextColor(tcell.ColorGray)
	default:
		trafficCell.SetTextColor(tcell.ColorGreen)
	}
	v.SetCell(row, 5, trafficCell)
}

// updateHeader updates the header content with session info
func (v *CloudRunView) updateHeader() {
	// Clear the header table
	v.headerTable.Clear()

	// Define the rows
	rows := []struct {
		field string
		value string
	}{
		{"Project ID", v.project},
		{"Region", v.region},
	}

	// Add the rows
	for i, row := range rows {
		// Field column
		fieldCell := tview.NewTableCell(row.field).
			SetTextColor(tcell.ColorWhite).
			SetExpansion(0).
			SetAlign(tview.AlignLeft).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false)
		v.headerTable.SetCell(i, 0, fieldCell)

		// Value column
		valueCell := tview.NewTableCell(row.value).
			SetTextColor(tcell.ColorWhite).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false)
		v.headerTable.SetCell(i, 1, valueCell)

		// Add empty cell to fill the rest of the row
		emptyCell := tview.NewTableCell("").
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false)
		v.headerTable.SetCell(i, 2, emptyCell)
	}

	// Add shortcuts in a new row
	shortcutsCell := tview.NewTableCell("Enter(Logs) Ctrl+D(Description)").
		SetTextColor(tcell.ColorGray).
		SetExpansion(1).
		SetAlign(tview.AlignRight).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(len(rows), 1, shortcutsCell)

	// Add empty cell in the shortcuts row
	emptyShortcutCell := tview.NewTableCell("").
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(len(rows), 0, emptyShortcutCell)
	v.headerTable.SetCell(len(rows), 2, emptyShortcutCell)
} // showServiceDescription displays detailed information about the selected service
func (v *CloudRunView) showServiceDescription() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	v.SetTitle(fmt.Sprintf(" Cloud Run Services - %s ", serviceName))
}

// showLogs displays logs for the selected service
func (v *CloudRunView) showLogs() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	region := v.GetCell(row, 1).Text

	// Create and show the log view
	logView := NewLogView(v.app, serviceName, region)
	v.app.SwitchToView(logView)
}
