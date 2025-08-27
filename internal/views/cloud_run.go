package views

import (
	"fmt"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/lpmourato/c9s/internal/config"
	"github.com/lpmourato/c9s/internal/datasource"
	"github.com/lpmourato/c9s/internal/model"
	"github.com/lpmourato/c9s/internal/ui"
)

// CloudRunView represents the Cloud Run services view
type CloudRunView struct {
	*ui.Table
	app          *ui.App
	headerTable  *ui.HeaderTable
	commandInput *ui.CommandInput
	config       *config.CloudRunConfig
	dataSource   datasource.DataSource
	services     []model.Service
	filter       string // Current service name filter
}

// Verify CloudRunView implements CommandHandler interface
var _ ui.CommandHandler = (*CloudRunView)(nil)

// HandleRegion implements CommandHandler
func (v *CloudRunView) HandleRegion(region string) error {
	v.config.Region = region
	if err := v.loadServices(); err != nil {
		return err
	}
	v.updateHeader()
	return nil
}

// HandleProject implements CommandHandler
func (v *CloudRunView) HandleProject(project string) error {
	if project == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	// Create config for the new data source
	cfg := &datasource.Config{
		Type:      datasource.GCP,
		ProjectID: project,
		Region:    v.config.Region,
	}

	// Create new data source with the new project
	newDS, err := datasource.Factory(cfg)
	if err != nil {
		return fmt.Errorf("failed to switch to project %s: %v", project, err)
	}

	// Update view with new project and data source
	v.config.ProjectID = project
	v.dataSource = newDS

	// Reload services for the new project
	if err := v.loadServices(); err != nil {
		return fmt.Errorf("failed to load services for project %s: %v", project, err)
	}

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
func NewCloudRunView(app *ui.App, cfg *config.CloudRunConfig, ds datasource.DataSource) *CloudRunView {
	table := ui.NewTable()
	table.SetApp(app)
	table.SetSelectable(true, false)

	// Create header table for session info
	headerTable := ui.NewHeaderTable()
	headerTable.SetTitle(" Cloud Run Context ")

	view := &CloudRunView{
		Table:       table,
		app:         app,
		headerTable: headerTable,
		config:      cfg,
		dataSource:  ds,
	}

	// Set up the table columns and style
	view.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})
	view.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorNavy))

	// Create main content flex (header + table)
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headerTable, 5, 0, false).
		AddItem(table, 0, 1, true)

	// Create command container with keyboard handling
	cmdContainer := ui.NewCommandContainer(app, mainFlex, view)
	view.commandInput = cmdContainer.GetCommandInput()

	// Set up additional keyboard shortcuts
	originalHandler := cmdContainer.GetInputCapture()
	cmdContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Let command input handle its events first
		if view.commandInput.IsVisible() {
			return event
		}

		// Handle view-specific shortcuts
		switch event.Key() {
		case tcell.KeyEnter:
			view.showLogs()
			return nil
		case tcell.KeyRune:
			if event.Rune() == 'd' || event.Rune() == 'D' {
				view.showDeploymentDetails()
				return nil
			}
			if event.Rune() == 's' || event.Rune() == 'S' {
				view.showServiceDescription()
				return nil
			}
		}

		// Pass through other events to the default handler
		if originalHandler != nil {
			return originalHandler(event)
		}
		return event
	})

	// Set as main view and update
	app.SetMainView(cmdContainer)
	app.SetFocus(view)
	view.updateHeader()
	if err := view.loadServices(); err != nil {
		app.Stop()
		return nil
	}

	return view
}

// loadServices loads services from the provider
func (v *CloudRunView) loadServices() error {
	var err error
	if v.config.Region != "" {
		v.services, err = v.dataSource.GetServicesByRegion(v.config.Region)
	} else {
		v.services, err = v.dataSource.GetServices()
	}
	if err != nil {
		return err
	}

	// Clear and reload table
	v.Clear()
	v.SetColumns([]string{"Name", "Region", "URL", "Status", "Last Deploy", "Traffic"})

	for i, svc := range v.services {
		if v.filter == "" || strings.Contains(strings.ToLower(svc.GetName()), strings.ToLower(v.filter)) {
			v.updateServiceRow(i+1, svc)
		}
	}

	if len(v.services) > 0 {
		v.Select(1, 0)
	}
	return nil
}

// updateServiceRow updates a single row in the table with service data
func (v *CloudRunView) updateServiceRow(row int, svc model.Service) {
	cells := []ui.TableCell{
		{
			Text:      svc.GetName(),
			Expansion: 1,
		},
		{
			Text:      svc.GetRegion(),
			Expansion: 1,
		},
		{
			Text:      svc.GetURL(),
			Expansion: 2,
		},
		{
			Text:      svc.GetStatus(),
			TextColor: ui.StatusColor(svc.GetStatus()),
			Expansion: 1,
		},
		{
			Text:      svc.GetLastDeploy().Format("2006-01-02 15:04:05"),
			Expansion: 1,
			Align:     tview.AlignRight,
		},
		{
			Text:      svc.GetTraffic(),
			TextColor: ui.TrafficColor(svc.GetTraffic()),
			Expansion: 2,
		},
	}
	v.AddStyledRow(row, cells)
}

// updateHeader updates the header content with session info
func (v *CloudRunView) updateHeader() {
	v.headerTable.Clear()

	// Left column: Project and Region info
	v.headerTable.AddLabelValueRow(0, "Project ID", v.config.ProjectID)
	v.headerTable.AddLabelValueRow(1, "Region", v.config.Region)

	// Add separator
	v.headerTable.AddSeparator(2, 3)

	// Right column: Shortcuts and Commands
	v.headerTable.AddSection(0, 3, "Keyboard Shortcuts", "Enter(Logs), D(Service Details)")
	v.headerTable.AddSection(1, 3, "Commands", ":region(rg) :project(proj) :service(svc) :clear(cl) :quit(q)")

	// Command input/hint row
	cmdHint := "Type Shift+: for commands"
	if v.commandInput.IsVisible() {
		cmdHint = v.commandInput.GetText()
	}
	v.headerTable.AddCommandHint(2, cmdHint, v.commandInput.IsVisible())
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

	// Create the log view
	logView, err := NewLogView(v.app, v.config.ProjectID, serviceName, region)
	if err != nil {
		// Show error in status bar
		v.app.ShowError(fmt.Sprintf("Failed to open logs: %v", err))
		return
	}

	// Start streaming immediately since NewLogView now sets up the streamer
	go logView.StreamLogs()

	// Switch to log view
	v.app.SwitchToView(logView)
}

// showDeploymentDetails displays deployment details for the selected service
func (v *CloudRunView) showDeploymentDetails() {
	row, _ := v.GetSelection()
	if row == 0 {
		return // Header row
	}

	serviceName := v.GetCell(row, 0).Text
	region := v.GetCell(row, 1).Text

	// Create deployment view
	deployView := NewDeploymentView(v.app, serviceName, region)

	// Start loading details using the provider from the data source
	deployView.LoadDetails(v.dataSource.GetProvider())

	// Switch to deployment view
	v.app.SwitchToView(deployView)
}
