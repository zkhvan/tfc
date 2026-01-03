package term

import (
	"bytes"
	"fmt"
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

func System() Term {
	t := Term{
		in:     os.Stdin,
		out:    os.Stdout,
		errOut: os.Stderr,
	}

	t.profile = colorprofile.Detect(os.Stdout, os.Environ())

	return t
}
func Test(in, out, errOut *bytes.Buffer, environ []string) Term {
	if environ == nil {
		environ = os.Environ()
	}

	t := Term{
		in:     in,
		out:    out,
		errOut: errOut,
	}

	// For tests, detect color profile from environment only
	// since we don't have a real terminal
	t.profile = colorprofile.Detect(out, environ)

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

// Size returns the width and height of the terminal that the current process
// is attached to. In case of errors, the numeric values returned are -1.
func (t Term) Size() (int, int, error) {
	// Try to get file descriptor from out
	var fd uintptr
	if f, ok := t.out.(*os.File); ok {
		fd = f.Fd()
	} else if fder, ok := t.out.(interface{ Fd() uintptr }); ok {
		fd = fder.Fd()
	} else {
		return 0, 0, fmt.Errorf("not connected to a terminal")
	}

	if !IsTerminal(fd) {
		return 0, 0, fmt.Errorf("not connected to a terminal")
	}

	width, height, err := GetSize(fd)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}
