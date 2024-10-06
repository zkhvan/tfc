package iolib

import (
	"io"

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

	In     io.Reader // think os.Stdin
	Out    io.Writer // think os.Stdout
	ErrOut io.Writer // think os.Stderr
}

func System() *IOStreams {
	terminal := term.New()

	streams := &IOStreams{
		In:     terminal.In(),
		Out:    terminal.Out(),
		ErrOut: terminal.ErrOut(),

		term: &terminal,
	}

	return streams
}

// TerminalWidth returns the width of the terminal that controls the process
func (s *IOStreams) TerminalWidth() int {
	w, _, err := s.term.Size()
	if err == nil && w > 0 {
		return w
	}
	return DefaultWidth
}
