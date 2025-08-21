package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// HeaderTable represents a styled header table
type HeaderTable struct {
	*tview.Table
}

// NewHeaderTable creates a new header table with standard styling
func NewHeaderTable() *HeaderTable {
	table := &HeaderTable{
		Table: tview.NewTable().
			SetBorders(false).
			SetSelectable(false, false),
	}

	table.SetBackgroundColor(tcell.ColorBlack)
	table.SetBorder(true)
	table.SetBorderColor(tcell.ColorGray)

	return table
}

// AddLabelValueRow adds a label-value pair row to the header
func (h *HeaderTable) AddLabelValueRow(row int, label, value string) {
	// Label cell
	h.SetCell(row, 0, tview.NewTableCell(label).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))

	// Value cell
	h.SetCell(row, 1, tview.NewTableCell(value).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))
}

// AddSeparator adds a vertical separator column
func (h *HeaderTable) AddSeparator(col int, rows int) {
	for i := 0; i < rows; i++ {
		h.SetCell(i, col, tview.NewTableCell("│").
			SetTextColor(tcell.ColorGray).
			SetBackgroundColor(tcell.ColorBlack).
			SetSelectable(false).
			SetAlign(tview.AlignCenter))
	}
}

// AddCommandHint adds a command hint row
func (h *HeaderTable) AddCommandHint(row int, text string, isActive bool) {
	textColor := tcell.ColorGray
	if isActive {
		textColor = tcell.ColorWhite
	}

	h.SetCell(row, 0, tview.NewTableCell("").
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))

	h.SetCell(row, 1, tview.NewTableCell(text).
		SetTextColor(textColor).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))
}

// AddSection adds a titled section with content
func (h *HeaderTable) AddSection(row, col int, title, content string) {
	// Title
	h.SetCell(row, col, tview.NewTableCell(title).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))

	// Content
	h.SetCell(row, col+1, tview.NewTableCell(content).
		SetTextColor(tcell.ColorGray).
		SetExpansion(1).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false))
}
