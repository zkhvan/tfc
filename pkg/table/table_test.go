package table_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/pkg/table"
	"github.com/zkhvan/tfc/pkg/text"
)

func Test_SingleColumn(t *testing.T) {
	newTable := func(w io.Writer) table.Table {
		return table.New(w,
			table.WithColumns(
				table.Column("ID"),
			),
		)
	}

	t.Run("single row", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID
			1
		`))
	})

	t.Run("no rows", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID
		`))
	})

	t.Run("single row with extra columns", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1", "2")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID  
			1   2
		`))
	})

	t.Run("second row with extra columns", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("XXXXX")
		tbl.AddRow("1", "2")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID     
			XXXXX
			1      2
		`))
	})

	t.Run("second row with extra long columns", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("XXXXX")
		tbl.AddRow("1", "XXXXX")
		tbl.AddRow("2")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID     
			XXXXX
			1      XXXXX
			2    
		`))
	})

	t.Run("single row, with multi-line content", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1\n2")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID
			1\n2
		`))
	})

	t.Run("second row, with multi-line content", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("XXXXX")
		tbl.AddRow("1\n2")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			ID
			XXXXX
			1\n2
		`))
	})
}

func Test_AutomaticColumns(t *testing.T) {
	newTable := func(w io.Writer) table.Table {
		return table.New(w)
	}

	t.Run("single column and single row", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			1
		`))
	})

	t.Run("single column and multiple rows", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1")
		tbl.AddRow("2")
		tbl.AddRow("3")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			1
			2
			3
		`))
	})

	t.Run("multiple column and single row", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("123", "456", "789")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			123  456  789
		`))
	})

	t.Run("multiple column and multiple rows", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1", "2", "3")
		tbl.AddRow("123", "456", "789")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			1    2    3
			123  456  789
		`))
	})

	t.Run("multiple column and multiple rows, with multi-line content", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1", "2", "3")
		tbl.AddRow("A\nB", "C\nD", "E\nF")
		tbl.AddRow("123", "456", "789")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			1     2     3
			A\nB  C\nD  E\nF
			123   456   789
		`))
	})

	t.Run("inconsistent columns with multiple rows", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1", "2")
		tbl.AddRow("123", "456", "789")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			1    2  
			123  456  789
		`))
	})
}

func Test_TruncateAutomaticColumns(t *testing.T) {
	newTable := func(w io.Writer) table.Table {
		return table.New(w,
			table.WithMaxWidth(9),
		)
	}

	t.Run("single column and single row", func(t *testing.T) {
		var buf bytes.Buffer

		tbl := newTable(&buf)
		tbl.AddRow("1234567890")
		tbl.Render()

		test.Buffer(t, &buf, text.Heredoc(`
			123456...
		`))
	})
}
