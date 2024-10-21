package clock

import (
	"time"
)

func Real() time.Time {
	return time.Now()
}

func FrozenClock(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}
