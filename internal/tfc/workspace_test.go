package tfc_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
)

func TestWorkspacesList(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			test.PathValue(t, r, "organization", "o")
			fmt.Fprint(w, `
				{
					"data": [
						{
							"id": "ws-1",
							"type": "workspaces",
							"attributes": {
								"name": "workspace-1"
							}
						}
					]
				}
			`)
		},
	)

	wss, _, err := client.Workspaces.List(context.TODO(), "o", &tfc.WorkspaceListOptions{})
	if err != nil {
		t.Error(err)
	}

	want := []*tfc.Workspace{
		{ID: "ws-1", Name: "workspace-1"},
	}

	if !cmp.Equal(wss, want) {
		t.Errorf("Workspaces.List diff: %s", cmp.Diff(wss, want))
	}
}
