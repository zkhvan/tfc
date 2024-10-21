package cmdutil_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc"

	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
)

func TestFieldPrinter_SingleField(t *testing.T) {
	io, _, out, _ := iolib.Test()

	t.Run("single row", func(t *testing.T) {
		rows := []map[string]string{
			{"KEY1": "1"},
		}

		p := cmdutil.FieldPrinter(io, "KEY1")
		for _, row := range rows {
			p.Write(row)
		}
		p.Flush()

		test.Buffer(t, out, heredoc.Doc(`
			KEY1
			1
		`))
	})
}
