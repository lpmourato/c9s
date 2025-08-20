package ui

import (
	"github.com/derailed/tview"
)

// App represents the application
type App struct {
	*tview.Application
	pages      *tview.Pages
	mainView   tview.Primitive // The main services view
	activeView tview.Primitive // The currently displayed view
}

// NewApp creates a new application instance
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:       tview.NewPages(),
	}

	return app
}

// SetMainView sets the main/home view of the application
func (a *App) SetMainView(view tview.Primitive) {
	a.mainView = view
	a.activeView = view
	a.SetRoot(view, true)
}

// GetMainView returns the main/home view
func (a *App) GetMainView() tview.Primitive {
	return a.mainView
}

// SwitchToView temporarily switches to another view
func (a *App) SwitchToView(view tview.Primitive) {
	a.activeView = view
	a.SetRoot(view, true)
}

// ReturnToMain returns to the main view
func (a *App) ReturnToMain() {
	a.activeView = a.mainView
	a.SetRoot(a.mainView, true)
}

// GetPages returns the application pages
func (a *App) GetPages() *tview.Pages {
	return a.pages
}

// Stop stops the application
func (a *App) Stop() {
	a.Application.Stop()
}
