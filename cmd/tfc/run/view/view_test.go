package view_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/zkhvan/tfc/cmd/tfc/run/view"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/clock"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
)

var referenceTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)

func TestView_byRunID(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/runs/{run_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": {
					"id": "run-abc123",
					"type": "runs",
					"attributes": {
						"status": "planned",
						"message": "Apply requested",
						"is-destroy": false,
						"plan-only": false,
						"refresh-only": false,
						"has-changes": true,
						"created-at": "2000-01-01T11:00:00Z"
					},
					"relationships": {
						"plan": {
							"data": { "id": "plan-xyz789", "type": "plans" }
						},
						"workspace": {
							"data": { "id": "ws-1", "type": "workspaces" }
						}
					}
				},
				"included": [
					{
						"id": "plan-xyz789",
						"type": "plans",
						"attributes": {
							"status": "finished",
							"has-changes": true,
							"resource-additions": 3,
							"resource-changes": 1,
							"resource-destructions": 0,
							"resource-imports": 0
						}
					},
					{
						"id": "ws-1",
						"type": "workspaces",
						"attributes": {
							"name": "my-workspace"
						},
						"relationships": {
							"organization": {
								"data": { "id": "org-1", "type": "organizations" }
							}
						}
					}
				]
			}`)
		},
	)

	result := runCommand(t, client, "run-abc123")

	test.BufferEmpty(t, result.ErrBuf)

	out := result.OutBuf.String()

	if !bytes.Contains(result.OutBuf.Bytes(), []byte("run-abc123")) {
		t.Errorf("expected output to contain run ID, got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("planned")) {
		t.Errorf("expected output to contain status, got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("3 to add")) {
		t.Errorf("expected output to contain resource summary, got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("1 to change")) {
		t.Errorf("expected output to contain resource changes, got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("PLAN SUMMARY")) {
		t.Errorf("expected output to contain plan summary section, got:\n%s", out)
	}
}

func TestView_byWorkspace(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{org}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": {
					"id": "ws-1",
					"type": "workspaces",
					"attributes": {
						"name": "my-workspace"
					},
					"relationships": {
						"current-run": {
							"data": { "id": "run-abc123", "type": "runs" }
						},
						"organization": {
							"data": { "id": "org-1", "type": "organizations" }
						}
					}
				}
			}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/runs/{run_id}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": {
					"id": "run-abc123",
					"type": "runs",
					"attributes": {
						"status": "planned",
						"message": "Queued from VCS",
						"is-destroy": false,
						"plan-only": false,
						"refresh-only": false,
						"has-changes": false,
						"created-at": "2000-01-01T11:00:00Z"
					},
					"relationships": {
						"plan": {
							"data": { "id": "plan-xyz789", "type": "plans" }
						},
						"workspace": {
							"data": { "id": "ws-1", "type": "workspaces" }
						}
					}
				},
				"included": [
					{
						"id": "plan-xyz789",
						"type": "plans",
						"attributes": {
							"status": "finished",
							"has-changes": false,
							"resource-additions": 0,
							"resource-changes": 0,
							"resource-destructions": 0,
							"resource-imports": 0
						}
					},
					{
						"id": "ws-1",
						"type": "workspaces",
						"attributes": {
							"name": "my-workspace"
						},
						"relationships": {
							"organization": {
								"data": { "id": "org-1", "type": "organizations" }
							}
						}
					}
				]
			}`)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace")

	test.BufferEmpty(t, result.ErrBuf)

	out := result.OutBuf.String()

	if !bytes.Contains(result.OutBuf.Bytes(), []byte("No changes")) {
		t.Errorf("expected output to contain 'No changes', got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("Queued from VCS")) {
		t.Errorf("expected output to contain message, got:\n%s", out)
	}
}

func TestView_noCurrentRun(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{org}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": {
					"id": "ws-1",
					"type": "workspaces",
					"attributes": {
						"name": "my-workspace"
					},
					"relationships": {
						"organization": {
							"data": { "id": "org-1", "type": "organizations" }
						}
					}
				}
			}`)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace")

	errOut := result.ErrBuf.String()
	if !bytes.Contains(result.ErrBuf.Bytes(), []byte("no current run")) {
		t.Errorf("expected error about no current run, got:\n%s", errOut)
	}
}

func runCommand(t *testing.T, client *tfc.Client, args ...string) *tfetest.CmdOut {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams: ios,
		TFEClient: func() (*tfc.Client, error) { return client, nil },
		Clock:     cmdutil.NewClock(clock.FrozenClock(referenceTime)),
	}

	cmd := view.NewCmdView(f)
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
