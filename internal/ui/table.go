package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// Table represents a table view
type Table struct {
	*tview.Table
}

// NewTable creates a new table instance
func NewTable() *Table {
	t := &Table{
		Table: tview.NewTable().SetBorders(true),
	}
	t.SetBorder(true)
	t.SetTitleAlign(tview.AlignLeft)
	return t
}

// NewTableCell creates a new table cell
func NewTableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text)
}

// SetColumns sets the table columns
func (t *Table) SetColumns(columns []string) {
	for col, header := range columns {
		cell := NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAttributes(tcell.AttrBold).
			SetSelectable(false)
		t.SetCell(0, col, cell)
	}
}

// SetCell sets a cell in the table
func (t *Table) SetCell(row, col int, cell *tview.TableCell) {
	t.Table.SetCell(row, col, cell)
}

// GetCell gets a cell from the table
func (t *Table) GetCell(row, col int) *tview.TableCell {
	return t.Table.GetCell(row, col)
}

// Clear clears the table content
func (t *Table) Clear() {
	t.Table.Clear()
}
