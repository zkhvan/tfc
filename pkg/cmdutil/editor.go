package cmdutil

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"

	shellquote "github.com/kballard/go-shellquote"

	"github.com/zkhvan/tfc/pkg/iolib"
)

// Editor provides interactive file editing using the user's configured editor.
type Editor struct {
	io *iolib.IOStreams
}

// NewEditor creates a new Editor instance that uses the provided IOStreams.
func NewEditor(io *iolib.IOStreams) *Editor {
	return &Editor{io: io}
}

// Edit opens the file at the given path in the user's configured editor and waits for it to exit.
// The context can be used to cancel the editor process.
// Checks TFC_EDITOR, VISUAL, EDITOR environment variables in order, falling back
// to vi (Unix) or notepad (Windows).
func (e *Editor) Edit(ctx context.Context, path string) error {
	editorArgs, err := resolveEditorCommand()
	if err != nil {
		return err
	}

	// #nosec G204 -- editorArgs comes from environment variables (TFC_EDITOR, VISUAL, EDITOR)
	// which are under user control. This is intentional to allow users to specify their editor.
	cmd := exec.CommandContext(ctx, editorArgs[0], append(editorArgs[1:], path)...) // #nosec G204
	cmd.Stdin = e.io.In
	cmd.Stdout = e.io.Out
	cmd.Stderr = e.io.ErrOut

	return cmd.Run()
}

// resolveEditorCommand returns the editor command with arguments parsed from
// shell-like syntax (e.g., "vim -n" or "code --wait").
func resolveEditorCommand() ([]string, error) {
	candidates := []string{
		os.Getenv("TFC_EDITOR"),
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
	}

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}

		args, err := shellquote.Split(candidate)
		if err != nil {
			return nil, err
		}
		if len(args) > 0 {
			return args, nil
		}
	}

	if runtime.GOOS == "windows" {
		return []string{"notepad"}, nil
	}

	return []string{"vi"}, nil
}
