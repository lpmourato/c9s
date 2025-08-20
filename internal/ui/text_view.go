package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

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
