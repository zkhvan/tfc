package approve_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/zkhvan/tfc/cmd/tfc/run/approve"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/clock"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
)

var referenceTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)

func TestApprove_byRunID(t *testing.T) {
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
						"has-changes": true,
						"created-at": "2000-01-01T11:00:00Z"
					},
					"relationships": {
						"plan": {
							"data": { "id": "plan-xyz", "type": "plans" }
						}
					}
				},
				"included": [
					{
						"id": "plan-xyz",
						"type": "plans",
						"attributes": {
							"status": "finished",
							"has-changes": true,
							"resource-additions": 1,
							"resource-changes": 0,
							"resource-destructions": 0
						}
					}
				]
			}`)
		},
	)

	mux.HandleFunc(
		"POST /api/v2/runs/{run_id}/actions/apply",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		},
	)

	result := runCommand(t, client, "run-abc123")

	test.BufferEmpty(t, result.ErrBuf)

	out := result.OutBuf.String()
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("Run approved for apply")) {
		t.Errorf("expected approval confirmation, got:\n%s", out)
	}
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("run-abc123")) {
		t.Errorf("expected run ID in output, got:\n%s", out)
	}
}

func TestApprove_withComment(t *testing.T) {
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
						"has-changes": true,
						"created-at": "2000-01-01T11:00:00Z"
					},
					"relationships": {
						"plan": {
							"data": { "id": "plan-xyz", "type": "plans" }
						}
					}
				},
				"included": [
					{
						"id": "plan-xyz",
						"type": "plans",
						"attributes": {
							"status": "finished",
							"has-changes": true
						}
					}
				]
			}`)
		},
	)

	mux.HandleFunc(
		"POST /api/v2/runs/{run_id}/actions/apply",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		},
	)

	result := runCommand(t, client, "run-abc123", "-m", "LGTM")

	test.BufferEmpty(t, result.ErrBuf)

	out := result.OutBuf.String()
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("LGTM")) {
		t.Errorf("expected comment in output, got:\n%s", out)
	}
}

func TestApprove_notApprovable(t *testing.T) {
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
						"status": "applied",
						"message": "Already applied",
						"has-changes": false,
						"created-at": "2000-01-01T11:00:00Z"
					}
				}
			}`)
		},
	)

	result := runCommand(t, client, "run-abc123")

	errOut := result.ErrBuf.String()
	if !bytes.Contains(result.ErrBuf.Bytes(), []byte("cannot be approved")) {
		t.Errorf("expected error about non-approvable status, got:\n%s", errOut)
	}
}

func TestApprove_byWorkspace(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{org}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": {
					"id": "ws-1",
					"type": "workspaces",
					"attributes": { "name": "my-workspace" },
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
						"message": "Plan complete",
						"has-changes": true,
						"created-at": "2000-01-01T11:00:00Z"
					},
					"relationships": {
						"plan": {
							"data": { "id": "plan-xyz", "type": "plans" }
						}
					}
				},
				"included": [
					{
						"id": "plan-xyz",
						"type": "plans",
						"attributes": {
							"status": "finished",
							"has-changes": true
						}
					}
				]
			}`)
		},
	)

	mux.HandleFunc(
		"POST /api/v2/runs/{run_id}/actions/apply",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		},
	)

	result := runCommand(t, client, "-W", "myorg/my-workspace")

	test.BufferEmpty(t, result.ErrBuf)

	out := result.OutBuf.String()
	if !bytes.Contains(result.OutBuf.Bytes(), []byte("Run approved for apply")) {
		t.Errorf("expected approval confirmation, got:\n%s", out)
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

	cmd := approve.NewCmdApprove(f)
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
