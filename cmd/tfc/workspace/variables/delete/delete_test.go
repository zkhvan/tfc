package delete_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/zkhvan/tfc/cmd/tfc/workspace/variables/delete"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

func TestDelete_by_name(t *testing.T) {
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
								"value": "some-value",
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

	mux.HandleFunc(
		"DELETE /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "MY_VAR")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"MY_VAR\" deleted successfully\n")
}

func TestDelete_by_id(t *testing.T) {
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
							"id": "var-abc123",
							"type": "vars",
							"attributes": {
								"key": "MY_VAR",
								"value": "some-value",
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

	mux.HandleFunc(
		"DELETE /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "var-abc123")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"MY_VAR\" deleted successfully\n")
}

func TestDelete_variable_not_found(t *testing.T) {
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
								"key": "OTHER_VAR",
								"value": "some-value",
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

	result := runCommand(t, client, "-W", "myorg/my-workspace", "NONEXISTENT_VAR")

	test.BufferEmpty(t, result.OutBuf)
	test.Buffer(t, result.ErrBuf, "variable \"NONEXISTENT_VAR\" not found\n")
}

func runCommand(t *testing.T, client *tfc.Client, args ...string) *tfetest.CmdOut {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams:       ios,
		TFEClient:       func() (*tfc.Client, error) { return client, nil },
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := delete.NewCmdDelete(f)
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
