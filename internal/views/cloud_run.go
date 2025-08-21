package views

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/model"
	"github.com/lpmourato/c9s/internal/ui"
)

// CloudRunView represents the Cloud Run services view
type CloudRunView struct {
	*ui.Table
	app          *ui.App
	headerTable  *tview.Table
	commandInput *ui.CommandInput
	project      string
	region       string
	services     []model.Service // Store all services
	filter       string          // Current service name filter
}

// Verify CloudRunView implements CommandHandler interface
var _ ui.CommandHandler = (*CloudRunView)(nil)

// HandleRegion implements CommandHandler
func (v *CloudRunView) HandleRegion(region string) error {
	v.region = region
	v.updateHeader()
	return nil
}

// HandleProject implements CommandHandler
func (v *CloudRunView) HandleProject(project string) error {
	v.project = project
	v.updateHeader()
	return nil
}

// HandleService implements CommandHandler
func (v *CloudRunView) HandleService(service string) error {
	v.filter = service
	// Clear and reload the table with the filter
	v.Clear()
	v.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})

	// Apply filter and update table
	rowIndex := 1 // Skip header
	for _, svc := range v.services {
		if service == "" || strings.Contains(strings.ToLower(svc.GetName()), strings.ToLower(service)) {
			v.updateServiceRow(rowIndex, svc)
			rowIndex++
		}
	}

	if rowIndex > 1 {
		v.Select(1, 0) // Select first matching row
	}

	v.updateHeader()
	return nil
}

// HandleClear implements CommandHandler
func (v *CloudRunView) HandleClear() error {
	v.filter = ""
	return v.HandleService("") // Reuse service handler with empty filter
}

// HandleQuit implements CommandHandler
func (v *CloudRunView) HandleQuit() {
	v.app.Stop()
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

	// Create command input
	cmdInput := ui.NewCommandInput(app, table)

	view := &CloudRunView{
		Table:        table,
		app:          app,
		headerTable:  headerTable,
		commandInput: cmdInput,
		project:      "dev-tla-cm",
		region:       "europe-west4",
	}

	// Set command handler
	cmdInput.SetCommandHandler(view)

	// Set up the table
	view.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
	view.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorNavy))

	// Update header immediately
	view.updateHeader()

	// Create flex layout
	var mainFlex, commandFlex *tview.Flex

	mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headerTable, 5, 0, false).
		AddItem(table, 0, 1, true)

	commandFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).
		AddItem(view.commandInput, 1, 0, false)

	// Set up input handlers at the root level
	commandFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Shift+: is usually ':' character
		if event.Key() == tcell.KeyRune {
			// Check for ':' character (which is what Shift+: produces)
			if event.Rune() == ':' {
				if !cmdInput.IsVisible() {
					cmdInput.Show()
					return nil
				}
			}
		}
		// If command input is visible, let it handle the event
		if cmdInput.IsVisible() {
			return event
		}

		// Handle other keyboard shortcuts
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

	// Set the command flex as the main view and give the table focus
	app.SetMainView(commandFlex)
	app.SetFocus(view) // Load mock data
	view.loadMockData()

	return view
}

// loadMockData loads mock service data into the table
func (v *CloudRunView) loadMockData() {
	v.services = model.GetMockServices()
	for i, svc := range v.services {
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

	// Left column: Project and Region info
	fieldCell := tview.NewTableCell("Project ID").
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(0, 0, fieldCell)

	valueCell := tview.NewTableCell(v.project).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(0, 1, valueCell)

	fieldCell = tview.NewTableCell("Region").
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(1, 0, fieldCell)

	valueCell = tview.NewTableCell(v.region).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(1, 1, valueCell)

	// Right column: Shortcuts and Commands
	shortcutsTitle := tview.NewTableCell("Keyboard Shortcuts").
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(0, 3, shortcutsTitle)

	shortcuts := tview.NewTableCell("Enter(Logs) Ctrl+D(Description)").
		SetTextColor(tcell.ColorGray).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(0, 4, shortcuts)

	commandsTitle := tview.NewTableCell("Commands").
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(1, 3, commandsTitle)

	commands := tview.NewTableCell(":region(rg) :project(proj) :service(svc) :clear(cl) :quit(q)").
		SetTextColor(tcell.ColorGray).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	v.headerTable.SetCell(1, 4, commands)

	// Command input/hint row
	hintRow := 2
	cmdHint := "Type Shift+: for commands"
	if v.commandInput.IsVisible() {
		cmdHint = v.commandInput.GetText()
	}

	textColor := tcell.ColorGray
	if v.commandInput.IsVisible() {
		textColor = tcell.ColorWhite
	}
	cmdHintCell := tview.NewTableCell(cmdHint).
		SetTextColor(textColor).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)

	v.headerTable.SetCell(hintRow, 0, tview.NewTableCell("").
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))
	v.headerTable.SetCell(hintRow, 1, cmdHintCell)

	// Add separator between left and right columns
	for i := 0; i < 3; i++ {
		v.headerTable.SetCell(i, 2, tview.NewTableCell("│").
			SetTextColor(tcell.ColorGray).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false).
			SetAlign(tview.AlignCenter))
	}
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
