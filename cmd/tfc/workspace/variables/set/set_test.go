package set_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/zkhvan/tfc/cmd/tfc/workspace/variables/set"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

func TestSet_update_existing_by_name(t *testing.T) {
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

	mux.HandleFunc(
		"PATCH /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, _ *http.Request) {
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

	result := runCommand(t, client, "-W", "myorg/my-workspace", "MY_VAR", "--value", "new-value")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"MY_VAR\" updated successfully\n")
}

func TestSet_update_existing_by_id(t *testing.T) {
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

	mux.HandleFunc(
		"PATCH /api/v2/workspaces/{workspace_id}/vars/{var_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-abc123",
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

	result := runCommand(t, client, "-W", "myorg/my-workspace", "var-abc123", "--value", "new-value")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"MY_VAR\" updated successfully\n")
}

func TestSet_create_new_variable(t *testing.T) {
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

	mux.HandleFunc(
		"POST /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-new123",
						"type": "vars",
						"attributes": {
							"key": "NEW_VAR",
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

	result := runCommand(t, client, "-W", "myorg/my-workspace", "NEW_VAR", "--value", "new-value")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"NEW_VAR\" created successfully\n")
}

func TestSet_with_hcl_flag(t *testing.T) {
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

	mux.HandleFunc(
		"POST /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-hcl123",
						"type": "vars",
						"attributes": {
							"key": "CONFIG",
							"value": "{\"key\": \"value\"}",
							"category": "terraform",
							"hcl": true,
							"sensitive": false
						}
					}
				}
			`)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "CONFIG", "--value", `{"key": "value"}`, "--hcl")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"CONFIG\" created successfully\n")
}

func TestSet_with_sensitive_flag(t *testing.T) {
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

	mux.HandleFunc(
		"POST /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-secret123",
						"type": "vars",
						"attributes": {
							"key": "SECRET",
							"value": "secret-value",
							"category": "terraform",
							"hcl": false,
							"sensitive": true
						}
					}
				}
			`)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "SECRET", "--value", "secret-value", "--sensitive")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"SECRET\" created successfully\n")
}

func TestSet_with_env_category(t *testing.T) {
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

	mux.HandleFunc(
		"POST /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `
				{
					"data": {
						"id": "var-env123",
						"type": "vars",
						"attributes": {
							"key": "PATH",
							"value": "/usr/local/bin",
							"category": "env",
							"hcl": false,
							"sensitive": false
						}
					}
				}
			`)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace", "PATH", "--value", "/usr/local/bin", "--category", "env")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, "Variable \"PATH\" created successfully\n")
}

func runCommand(t *testing.T, client *tfc.Client, args ...string) *tfetest.CmdOut {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams:       ios,
		TFEClient:       func() (*tfc.Client, error) { return client, nil },
		TerraformConfig: func() *tfconfig.TerraformConfig { return nil },
	}

	cmd := set.NewCmdSet(f)
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
