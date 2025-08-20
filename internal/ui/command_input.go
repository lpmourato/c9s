package ui

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

// CommandInput represents a command input field with suggestions
type CommandInput struct {
	*tview.InputField
	app         *App
	suggestions []CommandSuggestion
	onCommand   func(cmd string, args []string) bool
	visible     bool
	mainTable   tview.Primitive // The main table to focus when hiding
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
			if input.onCommand != nil {
				parts := strings.Fields(cmd)
				if len(parts) > 0 {
					input.onCommand(parts[0], parts[1:])
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

// Show makes the command input visible
func (c *CommandInput) Show() {
	c.visible = true
	c.SetText("")
	c.app.SetFocus(c)
}

// Hide hides the command input and returns focus to the table
func (c *CommandInput) Hide() {
	c.visible = false
	c.SetText("")
	if c.app != nil && c.mainTable != nil {
		c.app.SetFocus(c.mainTable)
	}
}

// IsVisible returns whether the command input is visible
func (c *CommandInput) IsVisible() bool {
	return c.visible
}

// SetCommandHandler sets the function to handle command execution
func (c *CommandInput) SetCommandHandler(handler func(cmd string, args []string) bool) {
	c.onCommand = handler
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
