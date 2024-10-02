package iolib

import (
	"io"

	"github.com/zkhvan/tfc/pkg/term"
)

type Terminal interface {
	IsTerminalOutput() bool
	IsColorEnabled() bool
	Is256ColorSupported() bool
	IsTrueColorSupported() bool
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
