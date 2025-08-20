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

// Stop stops the application
func (a *App) Stop() {
	a.Application.Stop()
}
