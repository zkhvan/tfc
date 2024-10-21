package cmdutil

import (
	"time"

	"github.com/zkhvan/tfc/pkg/clock"
)

type TimeProvider func() time.Time

type Clock struct {
	now TimeProvider
}

func NewClock(p TimeProvider) *Clock {
	if p == nil {
		p = clock.Real
	}
	return &Clock{
		now: p,
	}
}

func (c *Clock) Now() time.Time {
	return c.now()
}
