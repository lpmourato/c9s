package interfaces

import "github.com/derailed/tview"

// UIController defines the interface for UI control operations
type UIController interface {
	SwitchToView(view tview.Primitive)
	ReturnToMain()
	ShowError(msg string)
	QueueUpdateDraw(func())
}
