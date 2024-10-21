package cmdutil

import (
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/table"
)

type Printer interface {
	Fields() []string
	Write(v map[string]string)
}

type fieldPrinter struct {
	table table.Table

	fields []string
}

func FieldPrinter(streams *iolib.IOStreams, fields ...string) *fieldPrinter {
	p := &fieldPrinter{
		fields: fields,
	}

	columns := make([]*table.ColumnDef, 0, len(fields))
	for _, field := range fields {
		columns = append(columns, table.Column(field))
	}

	p.table = table.New(
		streams.Out,
		table.WithMaxWidth(streams.TerminalWidth()),
		table.WithColumns(columns...),
	)

	return p
}

func (p *fieldPrinter) Fields() []string {
	return p.fields
}

func (p *fieldPrinter) Write(v map[string]string) {
	if v == nil || len(p.fields) == 0 {
		return
	}

	row := make([]string, 0, len(p.fields))
	for _, field := range p.fields {
		if v, ok := v[field]; ok {
			row = append(row, v)
		} else {
			row = append(row, "")
		}

	}
	p.table.AddRow(row...)
}

func (p *fieldPrinter) Flush() {
	p.table.Render()
}
