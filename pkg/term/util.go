package term

import (
	"github.com/charmbracelet/x/term"
)

type (
	// State contains platform-specific state of a terminal.
	State = term.State

	// IsTerminal returns whether the given file descriptor is a terminal.
	File = term.File
)

var (
	// IsTerminal returns whether the given file descriptor is a terminal.
	IsTerminal = term.IsTerminal

	// MakeRaw puts the terminal connected to the given file descriptor into raw
	// mode and returns the previous state of the terminal so that it can be
	// restored.
	MakeRaw = term.MakeRaw

	// GetState returns the current state of a terminal which may be useful to
	// restore the terminal after a signal.
	GetState = term.GetState

	// SetState sets the given state of the terminal.
	SetState = term.SetState

	// Restore restores the terminal connected to the given file descriptor to a
	// previous state.
	Restore = term.Restore

	// GetSize returns the visible dimensions of the given terminal.
	//
	// These dimensions don't include any scrollback buffer height.
	GetSize = term.GetSize

	// ReadPassword reads a line of input from a terminal without local echo.  This
	// is commonly used for inputting passwords and other sensitive data. The slice
	// returned does not include the \n.
	ReadPassword = term.ReadPassword
)
