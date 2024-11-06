package list_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/cmd/tfc/workspaces/list"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfe/tfetest"
	"github.com/zkhvan/tfc/pkg/clock"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

var (
	referenceTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)
)

func TestList_default(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"o","type":"organizations","attributes":{"name":"o"}}]}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
				{
					"data": [
						{
							"id": "ws-1",
							"type": "workspaces",
							"attributes": {
								"name": "workspace-1",
								"updated-at": "1999-12-31T12:00:00Z"
							},
							"relationships": {
								"organization": {
									"data": {
										"id": "o",
										"type": "organizations"
									}
								}
							}
						}
					]
				}
			`)
		},
	)

	result := runCommand(t, client)

	test.Buffer(t, result.OutBuf, text.Heredoc(`
		ORG  NAME         UPDATED_AT
		o    workspace-1  about 1 day ago
	`))
}

func TestList_pagination(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
				{
					"data": [
						{
							"id": "o",
							"type": "organizations",
							"attributes": {
								"name": "o"
							}
						}
					]
				}
			`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			params := r.URL.Query()
			page := params.Get("page[number]")

			switch page {
			case "2":
				fmt.Fprint(w, `
					{
						"data": [
							{
								"id": "ws-2",
								"type": "workspaces",
								"attributes": {
									"name": "workspace-2",
									"updated-at": "1999-12-31T12:00:00Z"
								},
								"relationships": {
									"organization": {
										"data": {
											"id": "o",
											"type": "organizations"
										}
									}
								}
							}
						],
						"meta": {
							"pagination": {
								"current-page": 2,
								"prev-page": 1,
								"next-page": null,
								"total-pages": 2,
								"total-count": 2
							}
						}
					}
				`)
			default:
				fmt.Fprint(w, `
					{
						"data": [
							{
								"id": "ws-1",
								"type": "workspaces",
								"attributes": {
									"name": "workspace-1",
									"updated-at": "1999-12-31T12:00:00Z"
								},
								"relationships": {
									"organization": {
										"data": {
											"id": "o",
											"type": "organizations"
										}
									}
								}
							}
						],
						"meta": {
							"pagination": {
								"current-page": 1,
								"prev-page": null,
								"next-page": 2,
								"total-pages": 2,
								"total-count": 2
							}
						}
					}
				`)
			}
		},
	)

	result := runCommand(t, client)

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, text.Heredoc(`
		ORG  NAME         UPDATED_AT
		o    workspace-1  about 1 day ago
		o    workspace-2  about 1 day ago
	`))
}

func TestList_multiple_organizations(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `
				{
					"data": [
						{"id":"o1","type":"organizations","attributes":{"name":"o1"}},
						{"id":"o2","type":"organizations","attributes":{"name":"o2"}},
						{"id":"o3","type":"organizations","attributes":{"name":"o3"}}
					]
				}
			`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			org := r.PathValue("organization")

			switch org {
			case "o1":
				fmt.Fprint(w, `
					{
						"data": [
							{
								"id": "ws-1",
								"type": "workspaces",
								"attributes": {
									"name": "workspace-1",
									"updated-at": "1999-12-31T12:00:00Z"
								},
								"relationships": {
									"organization": {
										"data": {
											"id": "o1",
											"type": "organizations"
										}
									}
								}
							}
						]
					}
				`)
			case "o2":
				http.NotFound(w, r)
			case "o3":
				http.NotFound(w, r)
			}
		},
	)

	result := runCommand(t, client)
	test.Buffer(t, result.ErrBuf, text.Heredoc(`
		error listing workspaces for "o2": resource not found
		error listing workspaces for "o3": resource not found
	`))

	test.Buffer(t, result.OutBuf, text.Heredoc(`
		ORG  NAME         UPDATED_AT
		o1   workspace-1  about 1 day ago
	`))
}

func runCommand(t *testing.T, client *tfe.Client, args ...string) *tfetest.CmdOut {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams: ios,
		TFEClient: func() (*tfe.Client, error) { return client, nil },
		Clock:     cmdutil.NewClock(clock.FrozenClock(referenceTime)),
	}

	cmd := list.NewCmdList(f)
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
