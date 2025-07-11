package table

import (
	"fmt"
	"io"
	"strings"

	"github.com/zkhvan/tfc/pkg/term"
	"github.com/zkhvan/tfc/pkg/text"
)

type Table interface {
	AddRow(...string)
	Render()
}

type table struct {
	w io.Writer

	colDefs []*ColumnDef
	rows    [][]string

	delimiter string
	maxWidth  int
}

type tableOption func(*table)

func New(w io.Writer, opts ...tableOption) *table {
	t := &table{
		w:         w,
		delimiter: "  ",
	}

	// Detect if the table should be written to a terminal, then use the
	// terminal width as the max width
	if f, ok := w.(term.File); ok {
		if width, _, err := term.GetSize(f.Fd()); err == nil && width > 0 {
			t.maxWidth = width
		}
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

func WithMaxWidth(n int) tableOption {
	return func(t *table) {
		t.maxWidth = n
	}
}

func WithColumns(cols ...*ColumnDef) tableOption {
	return func(t *table) {
		t.colDefs = cols
	}
}

func (t *table) AddRow(row ...string) {
	row = sanitize(row)
	t.updateCols(row)

	t.rows = append(t.rows, row)
}

func (t *table) Render() {
	t.truncateCols()

	t.renderHeader()
	for _, row := range t.rows {
		t.renderRow(row)
	}
}

func (t *table) renderHeader() {
	if !t.hasHeader() {
		return
	}

	row := make([]string, 0, len(t.colDefs))

	for _, col := range t.colDefs {
		row = append(row, col.Header())
	}

	t.renderRow(row)
}

func (t *table) renderRow(row []string) {
	for col, value := range row {
		if col > 0 && len(t.delimiter) > 0 {
			fmt.Fprint(t.w, t.delimiter)
		}

		t.renderField(col, value)
	}
	fmt.Fprintf(t.w, "\n")
}

func (t *table) renderField(col int, value string) {
	colDef := t.colDefs[col]

	switch {
	case text.DisplayWidth(value) > colDef.width:
		fmt.Fprint(t.w, text.Truncate(colDef.width, value))
	case text.DisplayWidth(value) <= colDef.width:
		if col == len(t.colDefs)-1 {
			fmt.Fprint(t.w, value)
		} else {
			fmt.Fprint(t.w, text.PadRight(colDef.width, value))
		}
	}
}

func (t *table) hasHeader() bool {
	for _, col := range t.colDefs {
		if len(col.Header()) > 0 {
			return true
		}
	}

	return false
}

func (t *table) updateCols(row []string) {
	if len(t.colDefs) == 0 {
		t.colDefs = make([]*ColumnDef, 0, len(row))
	}

	if len(t.colDefs) < len(row) {
		for range len(row) - len(t.colDefs) {
			t.colDefs = append(t.colDefs, NewColumnDef())
		}
	}

	for i, value := range row {
		n := text.DisplayWidth(value)
		if t.colDefs[i].width < n {
			t.colDefs[i].width = n
		}
	}
}

func (t *table) truncateCols() {
	if t.maxWidth == 0 {
		return
	}

	var totalWidth int
	for _, col := range t.colDefs {
		totalWidth += col.width
	}
	totalWidth += (len(t.colDefs) - 1) * len(t.delimiter)

	if totalWidth <= t.maxWidth {
		return
	}

	var truncateCols, nonTruncateCols []*ColumnDef
	for _, col := range t.colDefs {
		if col.Truncate() {
			truncateCols = append(truncateCols, col)
		} else {
			nonTruncateCols = append(nonTruncateCols, col)
		}
	}

	// The columns should be truncated since they occupy more than the max
	// width.
	maxAvailableWidth := t.maxWidth - ((len(t.colDefs) - 1) * len(t.delimiter))
	for _, col := range nonTruncateCols {
		maxAvailableWidth -= col.width
	}

	var contentWidth int
	for _, col := range truncateCols {
		contentWidth += col.width
	}

	for _, col := range truncateCols {
		col.width = col.width * maxAvailableWidth / contentWidth
	}
}

func sanitize(values []string) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = strings.ReplaceAll(v, "\n", `\n`)
	}
	return out
}
