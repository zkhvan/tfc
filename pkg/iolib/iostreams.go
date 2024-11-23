package iolib

import (
	"bytes"
	"io"
	"os"

	"github.com/zkhvan/tfc/pkg/term"
)

const (
	DefaultWidth int = 80
)

type Terminal interface {
	IsTerminalOutput() bool
	IsColorEnabled() bool
	Is256ColorSupported() bool
	IsTrueColorSupported() bool
	Size() (int, int, error)
}

type IOStreams struct {
	term Terminal

	widthOverride int

	In     io.Reader // think os.Stdin
	Out    io.Writer // think os.Stdout
	ErrOut io.Writer // think os.Stderr

}

func System() *IOStreams {
	terminal := term.System()

	streams := &IOStreams{
		In:     terminal.In(),
		Out:    terminal.Out(),
		ErrOut: terminal.ErrOut(),

		term: &terminal,
	}

	return streams
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	var in, out, errOut bytes.Buffer

	terminal := term.Test(&in, &out, &errOut, os.Environ())

	streams := &IOStreams{
		In:     terminal.In(),
		Out:    terminal.Out(),
		ErrOut: terminal.ErrOut(),

		term: &terminal,
	}

	return streams, &in, &out, &errOut
}

func (s *IOStreams) OverrideTerminalWidth(w int) {
	s.widthOverride = w
}

// TerminalWidth returns the width of the terminal that controls the process
func (s *IOStreams) TerminalWidth() int {
	if s.widthOverride > 0 {
		return s.widthOverride
	}

	w, _, err := s.term.Size()
	if err == nil && w > 0 {
		return w
	}
	return DefaultWidth
}
