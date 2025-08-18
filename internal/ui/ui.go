package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// App represents the application UI
type App struct {
	*tview.Application
	pages *tview.Pages
}

// NewApp creates a new UI application
func NewApp() *App {
	app := &App{
		Application: tview.NewApplication(),
		pages:      tview.NewPages(),
	}

	app.SetRoot(app.pages, true)
	return app
}

// TextView represents a text view
type TextView struct {
	*tview.TextView
}

// NewTextView creates a new text view
func NewTextView() *TextView {
	tv := &TextView{
		TextView: tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true).
			SetWordWrap(true),
	}

	tv.SetBorder(true)
	tv.SetTitleAlign(tview.AlignLeft)
	return tv
}

// KeyBinding represents a keyboard shortcut binding
type KeyBinding struct {
	Key         tcell.Key
	Description string
	Action      func(event *tcell.EventKey)
}
