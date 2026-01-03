package cmdutil_test

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfetest"
	"github.com/zkhvan/tfc/pkg/cmdutil"
)

func TestCompleteColumns(t *testing.T) {
	options := []string{
		"ID",
		"Name",
		"Org",
		"Org1",
		"Org2",
	}

	complete := cmdutil.GenerateOptionCompletionFunc(options)

	t.Run("with no characters should return all options",
		func(t *testing.T) {
			toComplete := ""

			opts, _ := complete(nil, nil, toComplete)

			test.StringSlice(t, opts, options)
		},
	)

	t.Run("with one character",
		func(t *testing.T) {
			t.Run("should return the single filtered option",
				func(t *testing.T) {
					toComplete := "I"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"ID"})
				},
			)

			t.Run("should return multiple filtered options",
				func(t *testing.T) {
					toComplete := "O"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"Org", "Org1", "Org2"})
				},
			)
		},
	)

	t.Run("with multiple characters",
		func(t *testing.T) {
			t.Run("should return the single filtered option",
				func(t *testing.T) {
					toComplete := "ID"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"ID"})
				},
			)

			t.Run("should return the multiple filtered options",
				func(t *testing.T) {
					toComplete := "Org"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"Org", "Org1", "Org2"})
				},
			)
		},
	)

	t.Run("with one item and a comma",
		func(t *testing.T) {
			t.Run("should return the correctly filtered options and remove the completed option (ID)",
				func(t *testing.T) {
					toComplete := "ID,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name",
						"ID,Org",
						"ID,Org1",
						"ID,Org2",
					})
				},
			)

			t.Run("should return the correctly filtered options and remove the completed option (Org)",
				func(t *testing.T) {
					toComplete := "Org,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"Org,ID",
						"Org,Name",
						"Org,Org1",
						"Org,Org2",
					})
				},
			)
		},
	)

	t.Run("with one item, a comma, and some partial text",
		func(t *testing.T) {
			t.Run("should filter the options using the partial text",
				func(t *testing.T) {
					toComplete := "ID,N"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name",
					})
				},
			)
		},
	)

	t.Run("with multiple items and a comma",
		func(t *testing.T) {
			t.Run("should return the correctly filtered options and remove the completed option (ID)",
				func(t *testing.T) {
					toComplete := "ID,Name,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name,Org",
						"ID,Name,Org1",
						"ID,Name,Org2",
					})
				},
			)
		},
	)
}

func TestCompletionOrgWorkspace(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"o","type":"organizations","attributes":{"name":"o"}}]}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionOrgWorkspace(tfeClient)

	t.Run("should complete org names with trailing slash when no org specified",
		func(t *testing.T) {
			completions, directive := completionFunc(cmd, []string{}, "")

			if len(completions) != 1 {
				t.Errorf("Expected 1 completion, got %d", len(completions))
			}

			expectedCompletions := []string{"o/"}
			for _, expected := range expectedCompletions {
				found := slices.Contains(completions, expected)
				if !found {
					t.Errorf("Expected completion %q not found in %v", expected, completions)
				}
			}

			expectedDirective := cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			if directive != expectedDirective {
				t.Errorf("Expected directive %d (NoFileComp | NoSpace), got %d", expectedDirective, directive)
			}
		},
	)

	t.Run("should return empty when args already provided",
		func(t *testing.T) {
			completions, _ := completionFunc(cmd, []string{"o/w1"}, "")

			if len(completions) != 0 {
				t.Errorf("Expected 0 completions when arg already provided, got %d", len(completions))
			}
		},
	)
}

func TestCompletionOrgWorkspace_WithSlash(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/o/workspaces",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": [
					{"id":"ws-1","type":"workspaces","attributes":{"name":"w1"}},
					{"id":"ws-2","type":"workspaces","attributes":{"name":"w2"}}
				]
			}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionOrgWorkspace(tfeClient)

	t.Run("should list workspaces when org with slash provided",
		func(t *testing.T) {
			completions, directive := completionFunc(cmd, []string{}, "o/")

			if len(completions) != 2 {
				t.Errorf("Expected 2 completions, got %d", len(completions))
			}

			expectedCompletions := []string{"o/w1", "o/w2"}
			for _, expected := range expectedCompletions {
				found := slices.Contains(completions, expected)
				if !found {
					t.Errorf("Expected completion %q not found in %v", expected, completions)
				}
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("Expected ShellCompDirectiveNoFileComp, got %d", directive)
			}
		},
	)
}

func TestCompletionOrgWorkspace_WithOrgPrefix(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/myorg/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			searchParam := r.URL.Query().Get("search[name]")
			if searchParam != "prod" {
				t.Errorf("Expected search[name]=prod, got %q", searchParam)
			}

			fmt.Fprint(w, `{
				"data": [
					{"id":"ws-1","type":"workspaces","attributes":{"name":"production"}},
					{"id":"ws-2","type":"workspaces","attributes":{"name":"prod-staging"}}
				]
			}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionOrgWorkspace(tfeClient)

	t.Run("should filter workspaces by prefix with server-side search parameter",
		func(t *testing.T) {
			completions, directive := completionFunc(cmd, []string{}, "myorg/prod")

			if len(completions) != 2 {
				t.Errorf("Expected 2 completions, got %d", len(completions))
			}

			expectedCompletions := []string{"myorg/production", "myorg/prod-staging"}
			for _, expected := range expectedCompletions {
				found := slices.Contains(completions, expected)
				if !found {
					t.Errorf("Expected completion %q not found in %v", expected, completions)
				}
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("Expected ShellCompDirectiveNoFileComp, got %d", directive)
			}
		},
	)
}

func TestCompletionOrgWorkspace_WithOrgQueryFilter(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, r *http.Request) {
			queryParam := r.URL.Query().Get("q")
			if queryParam != "my" {
				t.Errorf("Expected q=my, got %q", queryParam)
			}

			fmt.Fprint(w, `{
				"data": [
					{"id":"myorg","type":"organizations","attributes":{"name":"myorg"}},
					{"id":"mycompany","type":"organizations","attributes":{"name":"mycompany"}}
				]
			}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionOrgWorkspace(tfeClient)

	t.Run("should filter organizations by query parameter and return with trailing slash",
		func(t *testing.T) {
			completions, directive := completionFunc(cmd, []string{}, "my")

			if len(completions) != 2 {
				t.Errorf("Expected 2 completions, got %d", len(completions))
			}

			expectedCompletions := []string{"myorg/", "mycompany/"}
			for _, expected := range expectedCompletions {
				found := slices.Contains(completions, expected)
				if !found {
					t.Errorf("Expected completion %q not found in %v", expected, completions)
				}
			}

			expectedDirective := cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
			if directive != expectedDirective {
				t.Errorf("Expected directive %d (NoFileComp | NoSpace), got %d", expectedDirective, directive)
			}
		},
	)
}

func TestCompletionVariableNames(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":{"id":"ws-123","type":"workspaces","attributes":{"name":"w1"}}}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{
				"data": [
					{"id":"var-1","type":"vars","attributes":{"key":"MY_VAR","value":"value1","category":"terraform"}},
					{"id":"var-2","type":"vars","attributes":{"key":"AWS_SECRET","value":"secret","category":"env"}}
				]
			}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionVariableNames(tfeClient)

	t.Run("should return variable names when valid org/workspace provided",
		func(t *testing.T) {
			completions, directive := completionFunc(cmd, []string{"o/w1"}, "")

			if len(completions) != 2 {
				t.Errorf("Expected 2 completions, got %d", len(completions))
			}

			expectedCompletions := []string{"MY_VAR", "AWS_SECRET"}
			for _, expected := range expectedCompletions {
				found := slices.Contains(completions, expected)
				if !found {
					t.Errorf("Expected completion %q not found in %v", expected, completions)
				}
			}

			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("Expected ShellCompDirectiveNoFileComp, got %d", directive)
			}
		},
	)

	t.Run("should return empty when no args provided",
		func(t *testing.T) {
			completions, _ := completionFunc(cmd, []string{}, "")
			if len(completions) != 0 {
				t.Errorf("Expected 0 completions with no args, got %d", len(completions))
			}
		},
	)

	t.Run("should return empty when invalid format provided",
		func(t *testing.T) {
			completions, _ := completionFunc(cmd, []string{"invalid"}, "")
			if len(completions) != 0 {
				t.Errorf("Expected 0 completions with invalid format, got %d", len(completions))
			}
		},
	)
}

func TestCompletionVariableNames_Timeout(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces/{workspace}",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data":{"id":"ws-123","type":"workspaces","attributes":{"name":"w1"}}}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/workspaces/{workspace_id}/vars",
		func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, `{"data": []}`)
		},
	)

	tfeClient := func() (*tfc.Client, error) {
		return client, nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	completionFunc := cmdutil.CompletionVariableNames(tfeClient)

	t.Run("should complete within timeout",
		func(t *testing.T) {
			done := make(chan bool)
			go func() {
				completions, _ := completionFunc(cmd, []string{"o/w1"}, "")
				_ = completions
				done <- true
			}()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("Completion function did not complete within 2 seconds")
			}
		},
	)
}

