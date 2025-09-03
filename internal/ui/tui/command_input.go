package tui

import (
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// CommandSuggestion represents a command suggestion
type CommandSuggestion struct {
	Command     string
	Alias       string
	Description string
}

// CommandHandler interface for components that want to handle commands
type CommandHandler interface {
	HandleRegion(region string) error
	HandleProject(project string) error
	HandleService(service string) error
	HandleClear() error
	HandleQuit()
}

// CommandInput represents a command input field with suggestions
type CommandInput struct {
	*tview.InputField
	app         *App
	suggestions []CommandSuggestion
	handler     CommandHandler
	visible     bool
	mainTable   tview.Primitive // The main table to focus when hiding
	container   *tview.Flex     // Reference to parent container for layout management
}

// NewCommandInput creates a new command input
func NewCommandInput(app *App, mainTable tview.Primitive) *CommandInput {
	input := &CommandInput{
		InputField: tview.NewInputField(),
		app:        app,
		mainTable:  mainTable,
		visible:    false,
	}

	input.
		SetLabel(":").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorWhite).
		SetFieldTextColor(tcell.ColorWhite)

	// Default suggestions
	input.suggestions = []CommandSuggestion{
		{Command: "region", Alias: "rg", Description: "Switch to a different region"},
		{Command: "service", Alias: "svc", Description: "Filter services by name"},
		{Command: "project", Alias: "proj", Description: "Switch to a different project"},
		{Command: "clear", Alias: "cl", Description: "Clear the current service filter"},
		{Command: "quit", Alias: "q", Description: "Exit the application"},
	}

	// Handle input events
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			input.Hide()
			return nil
		case tcell.KeyEnter:
			cmd := input.GetText()
			input.Hide()

			if input.handler != nil {
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					switch parts[0] {
					case "region", "rg":
						if len(parts) > 1 {
							input.handler.HandleRegion(parts[1])
						}
					case "project", "proj":
						if len(parts) > 1 {
							input.Show()
							input.ShowMessage("Switching to project " + parts[1] + "...")
							// Run project switching asynchronously to keep UI responsive
							go func(projectName string) {
								input.handler.HandleProject(projectName)
								// Hide the message after completion
								input.app.QueueUpdateDraw(func() {
									input.Hide()
								})
							}(parts[1])
						}
					case "service", "svc":
						if len(parts) > 1 {
							input.handler.HandleService(parts[1])
						}
					case "quit", "q":
						input.handler.HandleQuit()
					case "clear", "cl":
						input.handler.HandleClear()
					}
				}
			}
			return nil
		}
		return event
	})

	// Handle text changes for suggestions
	input.SetChangedFunc(func(text string) {
		// TODO: Show suggestions in a dropdown or status line
	})

	return input
}

// SetContainer sets the parent container reference for layout management
func (c *CommandInput) SetContainer(container *tview.Flex) {
	c.container = container
}

// SetMainTable sets the main table reference for layout management
func (c *CommandInput) SetMainTable(mainTable tview.Primitive) {
	c.mainTable = mainTable
}

// Show makes the command input visible
func (c *CommandInput) Show() {
	if !c.visible {
		c.visible = true
		c.SetText("")
		if c.container != nil {
			// Clear and rebuild the layout
			c.container.Clear()
			c.container.AddItem(c.mainTable, 0, 1, true)
			c.container.AddItem(c, 1, 0, false)
		}
		c.app.SetFocus(c)
	}
}

// Hide hides the command input and returns focus to the table
func (c *CommandInput) Hide() {
	if c.visible {
		c.visible = false
		c.SetText("")
		if c.container != nil {
			// Clear and rebuild the layout without the command input
			c.container.Clear()
			c.container.AddItem(c.mainTable, 0, 1, true)
		}
		if c.app != nil && c.mainTable != nil {
			c.app.SetFocus(c.mainTable)
		}
	}
}

// IsVisible returns whether the command input is visible
func (c *CommandInput) IsVisible() bool {
	return c.visible
}

// SetCommandHandler sets the handler for command execution
func (c *CommandInput) SetCommandHandler(handler CommandHandler) {
	c.handler = handler
}

// ShowMessage displays a message in the command input field
func (c *CommandInput) ShowMessage(msg string) {
	if !c.visible {
		c.visible = true
		if c.container != nil {
			// Clear and rebuild the layout
			c.container.Clear()
			c.container.AddItem(c.mainTable, 0, 1, true)
			c.container.AddItem(c, 1, 0, false)
		}
	}
	c.SetText(msg)
}

// GetSuggestions returns matching command suggestions for the given input
func (c *CommandInput) GetSuggestions(input string) []CommandSuggestion {
	if input == "" {
		return c.suggestions
	}

	var matches []CommandSuggestion
	input = strings.ToLower(input)
	for _, s := range c.suggestions {
		if strings.HasPrefix(strings.ToLower(s.Command), input) ||
			strings.HasPrefix(strings.ToLower(s.Alias), input) {
			matches = append(matches, s)
		}
	}
	return matches
}
