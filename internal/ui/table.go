package ui

import (
	"sync"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// TableWidget interface represents a table widget with app reference
type TableWidget interface {
	tview.Primitive
	SetApp(app *App)
	SetTitle(title string)
	SetColumns(columns []string)
	SetSelectedStyle(style tcell.Style)
	Select(row, col int)
	GetSelection() (row, col int)
	SetCell(row, col int, cell *tview.TableCell)
	GetCell(row, col int) *tview.TableCell
	Clear()
	GetColumnCount() int
}

// Table represents a table view
type Table struct {
	*tview.Table
	app *App
	mx  sync.RWMutex
}

// NewTable creates a new table instance
func NewTable() *Table {
	t := &Table{
		Table: tview.NewTable(),
	}
	t.SetBorder(true)
	t.SetBorders(true)
	t.SetTitleAlign(tview.AlignLeft)
	t.SetSelectable(true, false)
	t.SetFixed(1, 0)
	t.SetSeparator(tview.Borders.Vertical)

	return t
}

// SetApp sets the app reference
func (t *Table) SetApp(app *App) {
	t.app = app
}

// SetTitle sets the table title
func (t *Table) SetTitle(title string) {
	t.Table.SetTitle(title)
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

// SetSelectedStyle sets the selected row style
func (t *Table) SetSelectedStyle(style tcell.Style) {
	t.Table.SetSelectedStyle(style)
}

// Select selects a table cell
func (t *Table) Select(row, col int) {
	t.Table.Select(row, col)
}

// GetSelection returns the current selection
func (t *Table) GetSelection() (row, col int) {
	return t.Table.GetSelection()
}

// SetCell sets a table cell
func (t *Table) SetCell(row, col int, cell *tview.TableCell) {
	t.mx.Lock()
	defer t.mx.Unlock()
	t.Table.SetCell(row, col, cell)
}

// GetCell gets a table cell
func (t *Table) GetCell(row, col int) *tview.TableCell {
	return t.Table.GetCell(row, col)
}

// Clear clears the table content except for headers
func (t *Table) Clear() {
	headers := make([]*tview.TableCell, t.GetColumnCount())
	for i := 0; i < t.GetColumnCount(); i++ {
		headers[i] = t.GetCell(0, i)
	}
	t.Table.Clear()
	for i, header := range headers {
		if header != nil {
			t.SetCell(0, i, header)
		}
	}
}

// GetColumnCount returns the number of columns
func (t *Table) GetColumnCount() int {
	return t.Table.GetColumnCount()
}

// NewTableCell creates a new table cell
func NewTableCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetExpansion(1).
		SetAlign(tview.AlignLeft)
}
