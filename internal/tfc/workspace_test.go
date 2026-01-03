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

func TestParseOrgWorkspace(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedOrg       string
		expectedWorkspace string
		expectedComplete  bool
		expectedHasOrg    bool
		expectedString    string
	}{
		{
			name:              "empty string",
			input:             "",
			expectedOrg:       "",
			expectedWorkspace: "",
			expectedComplete:  false,
			expectedHasOrg:    false,
			expectedString:    "",
		},
		{
			name:              "org only",
			input:             "myorg",
			expectedOrg:       "",
			expectedWorkspace: "myorg",
			expectedComplete:  false,
			expectedHasOrg:    false,
			expectedString:    "myorg",
		},
		{
			name:              "org with slash",
			input:             "myorg/",
			expectedOrg:       "myorg",
			expectedWorkspace: "",
			expectedComplete:  false,
			expectedHasOrg:    true,
			expectedString:    "myorg/",
		},
		{
			name:              "org and workspace",
			input:             "myorg/workspace",
			expectedOrg:       "myorg",
			expectedWorkspace: "workspace",
			expectedComplete:  true,
			expectedHasOrg:    true,
			expectedString:    "myorg/workspace",
		},
		{
			name:              "org and partial workspace",
			input:             "myorg/work",
			expectedOrg:       "myorg",
			expectedWorkspace: "work",
			expectedComplete:  true,
			expectedHasOrg:    true,
			expectedString:    "myorg/work",
		},
		{
			name:              "org with multiple slashes in workspace",
			input:             "myorg/my/workspace",
			expectedOrg:       "myorg",
			expectedWorkspace: "my/workspace",
			expectedComplete:  true,
			expectedHasOrg:    true,
			expectedString:    "myorg/my/workspace",
		},
		{
			name:              "single character org",
			input:             "o",
			expectedOrg:       "",
			expectedWorkspace: "o",
			expectedComplete:  false,
			expectedHasOrg:    false,
			expectedString:    "o",
		},
		{
			name:              "single character org with slash",
			input:             "o/",
			expectedOrg:       "o",
			expectedWorkspace: "",
			expectedComplete:  false,
			expectedHasOrg:    true,
			expectedString:    "o/",
		},
		{
			name:              "single character org and workspace",
			input:             "o/w",
			expectedOrg:       "o",
			expectedWorkspace: "w",
			expectedComplete:  true,
			expectedHasOrg:    true,
			expectedString:    "o/w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := tfc.ParseOrgWorkspace(tt.input)

			if parsed.Org != tt.expectedOrg {
				t.Errorf("Expected org %q, got %q", tt.expectedOrg, parsed.Org)
			}

			if parsed.Workspace != tt.expectedWorkspace {
				t.Errorf("Expected workspace %q, got %q", tt.expectedWorkspace, parsed.Workspace)
			}

			if parsed.IsComplete() != tt.expectedComplete {
				t.Errorf("Expected IsComplete() %v, got %v", tt.expectedComplete, parsed.IsComplete())
			}

			if parsed.HasOrg() != tt.expectedHasOrg {
				t.Errorf("Expected HasOrg() %v, got %v", tt.expectedHasOrg, parsed.HasOrg())
			}

			if parsed.String() != tt.expectedString {
				t.Errorf("Expected String() %q, got %q", tt.expectedString, parsed.String())
			}
		})
	}
}

func TestOrgWorkspace_Validate(t *testing.T) {
	tests := []struct {
		name          string
		orgWorkspace  tfc.OrgWorkspace
		expectError   bool
		errorContains string
	}{
		{
			name: "valid complete",
			orgWorkspace: tfc.OrgWorkspace{
				Org:       "myorg",
				Workspace: "myworkspace",
			},
			expectError: false,
		},
		{
			name: "missing org",
			orgWorkspace: tfc.OrgWorkspace{
				Org:       "",
				Workspace: "myworkspace",
			},
			expectError:   true,
			errorContains: "organization cannot be empty",
		},
		{
			name: "missing workspace",
			orgWorkspace: tfc.OrgWorkspace{
				Org:       "myorg",
				Workspace: "",
			},
			expectError:   true,
			errorContains: "workspace cannot be empty",
		},
		{
			name: "both empty",
			orgWorkspace: tfc.OrgWorkspace{
				Org:       "",
				Workspace: "",
			},
			expectError:   true,
			errorContains: "organization and workspace cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.orgWorkspace.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorContains != "" {
				// Simple contains check
				found := false
				errStr := err.Error()
				for i := 0; i <= len(errStr)-len(tt.errorContains); i++ {
					if errStr[i:i+len(tt.errorContains)] == tt.errorContains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, errStr)
				}
			}
		})
	}
}
