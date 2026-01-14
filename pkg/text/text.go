package text

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

const (
	ellipsis            = "..."
	minWidthForEllipsis = len(ellipsis) + 2
)

func DisplayWidth(s string) int {
	return lipgloss.Width(s)
}

// Truncate returns a copy of the string s that has been shortened to fit the
// maximum display width.
func Truncate(maxWidth int, s string) string {
	w := DisplayWidth(s)
	if w <= maxWidth {
		return s
	}
	tail := ""
	if maxWidth >= minWidthForEllipsis {
		tail = ellipsis
	}

	if maxWidth < 0 {
		return s
	}
	r := truncate.StringWithTail(s, uint(maxWidth), tail)
	if DisplayWidth(r) < maxWidth {
		r += " "
	}
	return r
}

// PadRight returns a copy of the string s that has been padded on the right
// with whitespace to fit the maximum display width.
func PadRight(maxWidth int, s string) string {
	if padWidth := maxWidth - DisplayWidth(s); padWidth > 0 {
		s += strings.Repeat(" ", padWidth)
	}
	return s
}

// Pluralize returns a concatenated string with num and the plural form of
// thing if necessary.
func Pluralize(num int, thing string) string {
	if num == 1 {
		return fmt.Sprintf("%d %s", num, thing)
	}
	return fmt.Sprintf("%d %ss", num, thing)
}

func RelativeTimeAgo(a, b time.Time) string {
	ago := a.Sub(b)

	if ago < time.Minute {
		return "less than a minute ago"
	}
	if ago < time.Hour {
		return fmtDuration(int(ago.Minutes()), "minute")
	}
	if ago < 24*time.Hour {
		return fmtDuration(int(ago.Hours()), "hour")
	}
	if ago < 30*24*time.Hour {
		return fmtDuration(int(ago.Hours())/24, "day")
	}
	if ago < 365*24*time.Hour {
		return fmtDuration(int(ago.Hours())/24/30, "month")
	}

	return fmtDuration(int(ago.Hours()/24/365), "year")
}

func fmtDuration(amount int, unit string) string {
	return fmt.Sprintf("about %s ago", Pluralize(amount, unit))
}

func Heredoc(raw string) string {
	return heredoc.Doc(raw)
}

func Heredocf(raw string, args ...any) string {
	return heredoc.Docf(raw, args...)
}

func IndentHeredoc(amount int, raw string) string {
	return Indent(amount, Heredoc(raw))
}

func IndentHeredocf(amount int, raw string, args ...any) string {
	return Indent(amount, Heredocf(raw, args...))
}

// Indent will add indentation to every line in the string.
func Indent(amount int, s string) string {
	ss := strings.Split(s, "\n")
	for i, line := range ss {
		ss[i] = fmt.Sprintf("%s%s", strings.Repeat(" ", amount), line)
	}
	return strings.Join(ss, "\n")
}

// Clamp constrains a value between a minimum and maximum.
func Clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// TruncateBounded truncates a string to a width that is clamped between min and max.
func TruncateBounded(s string, width, minVal, maxVal int) string {
	clampedWidth := Clamp(width, minVal, maxVal)
	return Truncate(clampedWidth, s)
}
