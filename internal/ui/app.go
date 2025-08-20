package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// App represents the application
type App struct {
	*tview.Application
	pages *tview.Pages
}

// NewApp creates a new application instance
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
	}

	// Force periodic screen updates
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// Force redraw every time
		screen.Clear()
		return false
	})

	app.SetRoot(app.pages, true)
	return app
}

// GetPages returns the application pages
func (a *App) GetPages() *tview.Pages {
	return a.pages
}

// SwitchToView switches to a view
func (a *App) SwitchToView(name string, data interface{}) {
	if name == "logs" {
		// Log view can handle its own page management
		return
	}
	a.pages.SwitchToPage(name)
}

// Stop stops the application
func (a *App) Stop() {
	a.Application.Stop()
}

// HandleInputCapture handles key events coming from embedded components
func (a *App) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// Let all keys pass through to be handled by the focused primitive
	return event
}
