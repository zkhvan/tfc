package cmdutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zkhvan/tfc/pkg/iolib"
)

func TestResolveEditorCommand_Priority(t *testing.T) {
	tests := []struct {
		name          string
		tfcEditor     string
		visual        string
		editor        string
		expected      []string
		expectedError bool
	}{
		{
			name:      "TFC_EDITOR takes priority",
			tfcEditor: "tfc-vim",
			visual:    "visual-vim",
			editor:    "editor-vim",
			expected:  []string{"tfc-vim"},
		},
		{
			name:     "VISUAL is second priority",
			visual:   "visual-vim",
			editor:   "editor-vim",
			expected: []string{"visual-vim"},
		},
		{
			name:     "EDITOR is third priority",
			editor:   "editor-vim",
			expected: []string{"editor-vim"},
		},
		{
			name:      "TFC_EDITOR with arguments",
			tfcEditor: "vim -n --clean",
			expected:  []string{"vim", "-n", "--clean"},
		},
		{
			name:     "VISUAL with quoted arguments",
			visual:   `code --wait "/path/with spaces"`,
			expected: []string{"code", "--wait", "/path/with spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all editor env vars first
			t.Setenv("TFC_EDITOR", "")
			t.Setenv("VISUAL", "")
			t.Setenv("EDITOR", "")

			// Set the test values
			if tt.tfcEditor != "" {
				t.Setenv("TFC_EDITOR", tt.tfcEditor)
			}
			if tt.visual != "" {
				t.Setenv("VISUAL", tt.visual)
			}
			if tt.editor != "" {
				t.Setenv("EDITOR", tt.editor)
			}

			got, err := resolveEditorCommand()

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestResolveEditorCommand_Fallback(t *testing.T) {
	// Clear all editor env vars
	t.Setenv("TFC_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	got, err := resolveEditorCommand()
	require.NoError(t, err)

	var expected string
	if runtime.GOOS == "windows" {
		expected = "notepad"
	} else {
		expected = "vi"
	}

	assert.Equal(t, []string{expected}, got)
}

func TestEditor_Edit(t *testing.T) {
	// Create a mock editor script that writes content to the file
	mockEditor := createMockEditorScript(t, "edited content\n")
	t.Setenv("TFC_EDITOR", mockEditor)

	// Create a temp file to edit
	tmpFile, err := os.CreateTemp("", "editor-test-*.txt")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write initial content
	_, err = tmpFile.WriteString("initial content\n")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// Create editor with test IOStreams
	ios, _, _, _ := iolib.Test()
	editor := NewEditor(ios)

	// Edit the file
	ctx := context.Background()
	err = editor.Edit(ctx, tmpPath)
	require.NoError(t, err)

	// Verify the file was modified
	content, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, "edited content\n", string(content))
}

func TestEditor_Edit_WithArguments(t *testing.T) {
	// Create a mock editor script that accepts arguments
	mockEditor := createMockEditorScript(t, "content with args\n")
	t.Setenv("TFC_EDITOR", mockEditor+" --flag")

	tmpFile, err := os.CreateTemp("", "editor-test-*.txt")
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	_, err = tmpFile.WriteString("initial\n")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	ios, _, _, _ := iolib.Test()
	editor := NewEditor(ios)

	ctx := context.Background()
	err = editor.Edit(ctx, tmpPath)
	require.NoError(t, err)

	content, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, "content with args\n", string(content))
}

// createMockEditorScript creates an executable shell script that writes
// content to the file path passed as an argument.
func createMockEditorScript(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	var scriptPath string
	var scriptContent string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, "editor.bat")
		// Windows batch script - use last argument
		scriptContent = `@echo off
:loop
if "%2"=="" goto :last
shift
goto :loop
:last
echo ` + content + `> %1
`
	} else {
		scriptPath = filepath.Join(dir, "editor.sh")
		// Unix shell script - use last argument
		scriptContent = `#!/bin/sh
for last; do true; done
printf '` + content + `' > "$last"
`
	}

	// #nosec G306 -- Script must be executable (0755) to be run as editor command
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	return scriptPath
}
