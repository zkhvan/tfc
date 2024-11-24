package list_test

import (
	"bytes"
	"encoding/json"
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

func TestList_default(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"o","type":"organizations","attributes":{"name":"o"}}]}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, _ *http.Request) {
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

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, text.Heredoc(`
		ORG  NAME         RUN_STATUS  UPDATED_AT
		o    workspace-1              about 1 day ago
	`))
}

func TestList_pagination(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, _ *http.Request) {
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
		ORG  NAME         RUN_STATUS  UPDATED_AT
		o    workspace-1              about 1 day ago
		o    workspace-2              about 1 day ago
	`))
}

func TestList_pagination_with_client_side_filters(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"o","type":"organizations","attributes":{"name":"o"}}]}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, _ *http.Request) {
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
						},
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
							"current-page": 1,
							"prev-page": null,
							"next-page": null,
							"total-pages": 1,
							"total-count": 2
						}
					}
				}
			`)
		},
	)

	result := runCommand(t, client, "--limit", "1")

	test.BufferEmpty(t, result.ErrBuf)
	test.Buffer(t, result.OutBuf, text.Heredoc(`
		Showing top 1 results for org "o"
		ORG  NAME         RUN_STATUS  UPDATED_AT
		o    workspace-1              about 1 day ago
	`))
}

func TestList_multiple_organizations(t *testing.T) {
	t.Run("with errors in some organizations", func(t *testing.T) {
		client, mux, teardown := tfetest.Setup()
		defer teardown()

		mux.HandleFunc(
			"GET /api/v2/organizations",
			func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w,
					`{"data": [%s,%s,%s]}`,
					testOrg(t, "o1"),
					testOrg(t, "o2"),
					testOrg(t, "o3"),
				)
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
			ORG  NAME         RUN_STATUS  UPDATED_AT
			o1   workspace-1              about 1 day ago
		`))
	})

	t.Run("with limit for each org", func(t *testing.T) {
		client, mux, teardown := tfetest.Setup()
		defer teardown()

		mux.HandleFunc(
			"GET /api/v2/organizations",
			func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w,
					`{"data": [%s,%s,%s]}`,
					testOrg(t, "o1"),
					testOrg(t, "o2"),
					testOrg(t, "o3"),
				)
			},
		)

		mux.HandleFunc(
			"GET /api/v2/organizations/{organization}/workspaces",
			func(w http.ResponseWriter, r *http.Request) {
				org := r.PathValue("organization")

				testWorkspace := func(name, org string) string {
					t.Helper()

					return text.Heredocf(
						`
							{
								"id": "%[1]s",
								"type": "workspaces",
								"attributes": {
									"name": "%[1]s",
									"updated-at": "1999-12-31T12:00:00Z"
								},
								"relationships": {
									"organization": {
										"data": {
											"id": "%[2]s",
											"type": "organizations"
										}
									}
								}
							}
						`,
						name,
						org,
					)
				}

				switch org {
				case "o1":
					fmt.Fprintf(w, `{"data": [%s]}`, testWorkspace("workspace-1", "o1"))
				case "o2":
					fmt.Fprintf(w, `{"data": [%s,%s], "meta": {"pagination": %s}}`,
						testWorkspace("workspace-1", "o2"),
						testWorkspace("workspace-2", "o2"),
						testPagination(t, 1, 1, 2),
					)
				case "o3":
					fmt.Fprint(w, `{"data": []}`)
				}
			},
		)

		result := runCommand(t, client, "--limit", "1")

		test.BufferEmpty(t, result.ErrBuf)
		test.Buffer(t, result.OutBuf, text.Heredoc(`
			ORG  NAME         RUN_STATUS  UPDATED_AT
			o1   workspace-1              about 1 day ago
			Showing top 1 results for org "o2"
			ORG  NAME         RUN_STATUS  UPDATED_AT
			o2   workspace-1              about 1 day ago
			ORG  NAME  RUN_STATUS  UPDATED_AT
		`))
	})
}

var (
	referenceTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)
)

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

func testOrg(t *testing.T, name string) string {
	t.Helper()

	return text.Heredocf(
		`{"id":"%[1]s","type":"organizations","attributes":{"name":"%[1]s"}}`,
		name,
	)
}

func testPagination(t *testing.T, page, totalPages, totalCount int) string {
	t.Helper()

	type pagination struct {
		CurrentPage int  `json:"current-page"`
		PrevPage    *int `json:"prev-page"`
		NextPage    *int `json:"next-page"`
		TotalPages  int  `json:"total-pages"`
		TotalCount  int  `json:"total-count"`
	}

	var prevPage, nextPage int
	if page > 1 {
		prevPage = page - 1
	}
	if totalPages > page {
		nextPage = page + 1
	}

	p := pagination{
		CurrentPage: page,
		PrevPage:    &prevPage,
		NextPage:    &nextPage,
		TotalPages:  totalPages,
		TotalCount:  totalCount,
	}

	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}
