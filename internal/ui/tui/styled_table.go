package tui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// TableCell represents a styled table cell configuration
type TableCell struct {
	Text      string
	TextColor tcell.Color
	Expansion int
	Align     int
}

// TableRow represents a collection of styled cells
type TableRow struct {
	Cells []TableCell
}

// AddStyledRow adds a row of styled cells to the table
func (t *Table) AddStyledRow(row int, cells []TableCell) {
	for col, cell := range cells {
		t.SetCell(row, col, NewTableCell(cell.Text).
			SetTextColor(cell.TextColor).
			SetExpansion(cell.Expansion).
			SetAlign(cell.Align))
	}
}

// AddHeaderCell adds a standardized header cell to the table
func (t *Table) AddHeaderCell(row, col int, text string) *tview.TableCell {
	cell := NewTableCell(text).
		SetTextColor(tcell.ColorWhite).
		SetExpansion(0).
		SetAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack).
		SetSelectable(false)
	t.SetCell(row, col, cell)
	return cell
}

// StatusColor returns the appropriate color for a given status
func StatusColor(status string) tcell.Color {
	switch status {
	case "Ready":
		return tcell.ColorGreen
	case "Not Ready":
		return tcell.ColorRed
	case "Unknown":
		return tcell.ColorYellow
	default:
		return tcell.ColorGray
	}
}

// TrafficColor returns the appropriate color for a traffic status
func TrafficColor(traffic string) tcell.Color {
	switch traffic {
	case "No traffic (failed)":
		return tcell.ColorRed
	case "No traffic (stopped)":
		return tcell.ColorYellow
	case "No traffic":
		return tcell.ColorGray
	default:
		return tcell.ColorGreen
	}
}
