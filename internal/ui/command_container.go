package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// CommandContainer manages the command input and its layout
type CommandContainer struct {
	*tview.Flex
	app         *App
	mainView    tview.Primitive
	commandView *CommandInput
}

// NewCommandContainer creates a new command container
func NewCommandContainer(app *App, mainView tview.Primitive, handler CommandHandler) *CommandContainer {
	container := &CommandContainer{
		Flex:     tview.NewFlex().SetDirection(tview.FlexRow),
		app:      app,
		mainView: mainView,
	}

	// Create and set up command input
	cmdInput := NewCommandInput(app, mainView)
	cmdInput.SetCommandHandler(handler)
	container.commandView = cmdInput

	// Set up the flex layout
	container.AddItem(mainView, 0, 1, true)
	container.AddItem(cmdInput, 1, 0, false)

	// Set up input capture for the container
	container.SetInputCapture(container.handleKeyEvents)

	return container
}

// handleKeyEvents handles keyboard events for the command container
func (c *CommandContainer) handleKeyEvents(event *tcell.EventKey) *tcell.EventKey {
	// Handle Shift+: to show command input
	if event.Key() == tcell.KeyRune && event.Rune() == ':' {
		if !c.commandView.IsVisible() {
			c.commandView.Show()
			return nil
		}
	}

	// If command input is visible, let it handle all events
	if c.commandView.IsVisible() {
		return event
	}

	// Otherwise, pass the event through
	return event
}

// GetCommandInput returns the command input component
func (c *CommandContainer) GetCommandInput() *CommandInput {
	return c.commandView
}

// SetMainFocus sets focus to the main view
func (c *CommandContainer) SetMainFocus() {
	c.app.SetFocus(c.mainView)
}
