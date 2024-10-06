package table

import (
	"github.com/zkhvan/tfc/pkg/ptr"
)

type columnDef struct {
	width    int
	header   *string
	truncate *bool
}

func (c *columnDef) Truncate() bool {
	if c.truncate == nil {
		return true
	}

	return *c.truncate
}

func (c *columnDef) Header() string {
	if c.header == nil {
		return ""
	}

	return *c.header
}

type columnOption func(*columnDef)

func Column(header string, opts ...columnOption) *columnDef {
	c := NewColumnDef(
		append(
			opts,
			ColumnWithHeader(header),
		)...,
	)
	return c
}

func NewColumnDef(opts ...columnOption) *columnDef {
	c := &columnDef{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func ColumnWithHeader(s string) columnOption {
	return func(c *columnDef) {
		c.header = ptr.String(s)

		n := len(s)
		if c.width < n {
			c.width = n
		}
	}
}

func ColumnWithTruncation() columnOption {
	return func(c *columnDef) {
		c.truncate = ptr.Bool(true)
	}
}

func ColumnWithNoTruncation() columnOption {
	return func(c *columnDef) {
		c.truncate = ptr.Bool(false)
	}
}
