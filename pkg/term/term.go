package term

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

// Term represents information about the terminal that a process is connected
// to.
type Term struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer

	profile colorprofile.Profile
}

func New() Term {
	t := Term{
		in:     os.Stdin,
		out:    colorprofile.NewWriter(os.Stdout, os.Environ()),
		errOut: colorprofile.NewWriter(os.Stderr, os.Environ()),
	}

	t.profile = colorprofile.Detect(t.out, os.Environ())

	return t
}

// In is the reader reading from standard input.
func (t Term) In() io.Reader {
	return t.in
}

// Out is the writer writing to standard output.
func (t Term) Out() io.Writer {
	return t.out
}

// ErrOut is the writer writing to standard error.
func (t Term) ErrOut() io.Writer {
	return t.errOut
}

// IsTerminalOutput returns true if standard output is connected to a terminal.
func (t Term) IsTerminalOutput() bool {
	return t.profile != colorprofile.NoTTY
}

// IsColorEnabled reports whether it's safe to output ANSI color sequences,
// depending on IsTerminalOutput and environment variables.
func (t Term) IsColorEnabled() bool {
	return t.profile != colorprofile.NoTTY &&
		t.profile != colorprofile.Ascii
}

// Is256ColorSupported reports whether the terminal advertises ANSI 256 color
// codes.
func (t Term) Is256ColorSupported() bool {
	return t.profile == colorprofile.TrueColor ||
		t.profile == colorprofile.ANSI256
}

// IsTrueColorSupported reports whether the terminal advertises support for
// ANSI true color sequences.
func (t Term) IsTrueColorSupported() bool {
	return t.profile == colorprofile.TrueColor
}
