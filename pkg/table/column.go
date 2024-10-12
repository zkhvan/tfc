package table

import (
	"github.com/zkhvan/tfc/pkg/ptr"
)

type ColumnDef struct {
	width    int
	header   *string
	truncate *bool
}

func (c *ColumnDef) Truncate() bool {
	if c.truncate == nil {
		return true
	}

	return *c.truncate
}

func (c *ColumnDef) Header() string {
	if c.header == nil {
		return ""
	}

	return *c.header
}

type columnOption func(*ColumnDef)

func Column(header string, opts ...columnOption) *ColumnDef {
	c := NewColumnDef(
		append(
			opts,
			ColumnWithHeader(header),
		)...,
	)
	return c
}

func NewColumnDef(opts ...columnOption) *ColumnDef {
	c := &ColumnDef{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func ColumnWithHeader(s string) columnOption {
	return func(c *ColumnDef) {
		c.header = ptr.String(s)

		n := len(s)
		if c.width < n {
			c.width = n
		}
	}
}

func ColumnWithTruncation() columnOption {
	return func(c *ColumnDef) {
		c.truncate = ptr.Bool(true)
	}
}

func ColumnWithNoTruncation() columnOption {
	return func(c *ColumnDef) {
		c.truncate = ptr.Bool(false)
	}
}
