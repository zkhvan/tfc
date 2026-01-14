package edit_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zkhvan/tfc/cmd/tfc/workspace/variables/edit"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

func TestEdit_updates_variable_value(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping interactive editor test on Windows")
	}

	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "ws-123",
						"type": "workspaces",
						"attributes": {
							"name": "my-workspace"
						}
					}
				}
			`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": [
						{
							"id": "var-123",
							"type": "vars",
							"attributes": {
								"key": "MY_VAR",
								"value": "old-value",
								"category": "terraform",
								"hcl": false,
								"sensitive": false
							}
						}
					]
				}
			`)
		},
	)

	var (
		patchCalled bool
		gotValue    string
		patchErr    error
	)

	mux.HandleFunc(
		"PATCH /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, r *http.Request) {
			patchCalled = true

			defer r.Body.Close()

			var payload struct {
				Data struct {
					Attributes struct {
						Value string `json:"value"`
					} `json:"attributes"`
				} `json:"data"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				patchErr = err
			} else {
				gotValue = payload.Data.Attributes.Value
			}

			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-123",
						"type": "vars",
						"attributes": {
							"key": "MY_VAR",
							"value": "new-value",
							"category": "terraform",
							"hcl": false,
							"sensitive": false
						}
					}
				}
			`)
		},
	)

	script := createEditorScript(t, "new-value")

	t.Setenv("TFC_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "MY_VAR")

	if patchErr != nil {
		t.Fatalf("failed to decode patch payload: %v", patchErr)
	}
	if !patchCalled {
		t.Fatalf("expected PATCH request to be made")
	}
	if gotValue != "new-value" {
		t.Fatalf("expected updated value to be \"new-value\", got %q", gotValue)
	}

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"MY_VAR\" updated successfully\n")
}

func TestEdit_no_changes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping interactive editor test on Windows")
	}

	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "ws-123",
						"type": "workspaces",
						"attributes": {
							"name": "my-workspace"
						}
					}
				}
			`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": [
						{
							"id": "var-123",
							"type": "vars",
							"attributes": {
								"key": "MY_VAR",
								"value": "unchanged",
								"category": "terraform",
								"hcl": false,
								"sensitive": false
							}
						}
					]
				}
			`)
		},
	)

	var patchCalled bool

	mux.HandleFunc(
		"PATCH /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			patchCalled = true
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
		},
	)

	script := createEditorScript(t, "unchanged")

	t.Setenv("TFC_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "MY_VAR")

	if patchCalled {
		t.Fatalf("expected no PATCH request to be made when value is unchanged")
	}

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "No changes made to variable \"MY_VAR\"\n")
}

func TestEdit_variable_not_found(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping interactive editor test on Windows")
	}

	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "ws-123",
						"type": "workspaces",
						"attributes": {
							"name": "my-workspace"
						}
					}
				}
			`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": []
				}
			`)
		},
	)

	script := createEditorScript(t, "does-not-matter")

	t.Setenv("TFC_EDITOR", "")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "missing")

	test.Buffer(t, result.ErrBuf, "variable \"missing\" not found\n")
	test.BufferEmpty(t, result.OutBuf)
}

func runCommand(t *testing.T, client *tfc.Client, args ...string) *tfetest.CmdOut {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams: ios,
		TFEClient: func() (*tfc.Client, error) { return client, nil },
		Editor: func() *cmdutil.Editor {
			return cmdutil.NewEditor(ios)
		},
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := edit.NewCmdEdit(f)
	cmd.SetArgs(args)

	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	_, err := cmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(stderr, err)
	}

	return &tfetest.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}
}

func createEditorScript(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "editor.sh")

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create editor script: %v", err)
	}

	escaped := strings.ReplaceAll(content, "'", "'\"'\"'")

	if _, err := fmt.Fprintf(file, "#!/bin/sh\nprintf '%%s' '%s' > \"$1\"\n", escaped); err != nil {
		file.Close()
		t.Fatalf("failed to write editor script: %v", err)
	}

	if err := file.Close(); err != nil {
		t.Fatalf("failed to close editor script: %v", err)
	}

	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatalf("failed to chmod editor script: %v", err)
	}

	return path
}
