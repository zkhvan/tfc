package tfc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
)

type WorkspacesService service

type Workspace = tfe.Workspace

// OrgWorkspace represents a parsed organization/workspace identifier.
type OrgWorkspace struct {
	Org       string
	Workspace string
}

// IsComplete returns true if both org and workspace are present.
func (ow OrgWorkspace) IsComplete() bool {
	return ow.Org != "" && ow.Workspace != ""
}

// HasOrg returns true if an organization name is present.
func (ow OrgWorkspace) HasOrg() bool {
	return ow.Org != ""
}

// String returns the org/workspace format.
func (ow OrgWorkspace) String() string {
	if ow.HasOrg() {
		return fmt.Sprintf("%s/%s", ow.Org, ow.Workspace)
	}
	return ow.Workspace
}

// Validate returns an error if the org/workspace is not complete.
func (ow OrgWorkspace) Validate() error {
	if !ow.IsComplete() {
		if ow.Org == "" && ow.Workspace == "" {
			return fmt.Errorf("organization and workspace cannot be empty")
		}
		if ow.Org == "" {
			return fmt.Errorf("organization cannot be empty")
		}
		return fmt.Errorf("workspace cannot be empty")
	}
	return nil
}

// ParseOrgWorkspace parses an "org/workspace" identifier.
// Handles partial input for shell completion scenarios.
//
// Examples:
//   - "myorg/workspace" -> OrgWorkspace{Org: "myorg", Workspace: "workspace"}
//   - "myorg/" -> OrgWorkspace{Org: "myorg", Workspace: ""}
//   - "myorg" -> OrgWorkspace{Org: "", Workspace: "myorg"}
//   - "" -> OrgWorkspace{Org: "", Workspace: ""}
func ParseOrgWorkspace(input string) OrgWorkspace {
	if idx := strings.Index(input, "/"); idx != -1 {
		return OrgWorkspace{
			Org:       input[:idx],
			Workspace: input[idx+1:],
		}
	}
	return OrgWorkspace{
		Org:       "",
		Workspace: input,
	}
}

type WorkspaceListOptions struct {
	ListOptions

	// Optional: A search string (partial workspace name) used to filter the
	// results.
	Search string `url:"search[name],omitempty"`

	// Optional: A search string (comma-separated tag names) used to filter
	// the results.
	Tags string `url:"search[tags],omitempty"`

	// Optional: A search string (comma-separated tag names to exclude) used
	// to filter the results.
	ExcludeTags string `url:"search[exclude-tags],omitempty"`

	// Optional: A search on substring matching to filter the results.
	WildcardName string `url:"search[wildcard-name],omitempty"`

	// Optional: A filter string to list all the workspaces linked to a given
	// project id in the organization.
	ProjectID string `url:"filter[project][id],omitempty"`

	// Optional: A filter string to list all the workspaces filtered by
	// current run status.
	CurrentRunStatus string `url:"filter[current-run][status],omitempty"`

	// Optional: A filter string to list workspaces filtered by key/value tags.
	// These are not annotated and therefore not encoded by go-querystring
	TagBindings []*tfe.TagBinding

	// Optional: A list of relations to include. See available resources
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspaces#available-related-resources
	Include []tfe.WSIncludeOpt `url:"include,omitempty"`

	// Optional: May sort on "name" (the default) and "current-run.created-at"
	// (which sorts by the time of the current run) Prepending a hyphen to the
	// sort parameter will reverse the order (e.g. "-name" to reverse the
	// default order)
	Sort string `url:"sort,omitempty"`

	// Optional: A list of VCS repositories to filter by (client-side).
	VCSRepos []string
}

func (s *WorkspacesService) Read(
	ctx context.Context,
	org string,
	workspace string,
) (*Workspace, error) {
	return s.tfe.Workspaces.Read(ctx, org, workspace)
}

func (s *WorkspacesService) ReadWithOptions(
	ctx context.Context,
	org string,
	workspace string,
	opts *tfe.WorkspaceReadOptions,
) (*Workspace, error) {
	return s.tfe.Workspaces.ReadWithOptions(ctx, org, workspace, opts)
}

func (s *WorkspacesService) List(
	ctx context.Context,
	org string,
	opts *WorkspaceListOptions,
) ([]*Workspace, *Pagination, error) {
	o := tfe.WorkspaceListOptions{
		Search:           opts.Search,
		Tags:             opts.Tags,
		ExcludeTags:      opts.ExcludeTags,
		WildcardName:     opts.WildcardName,
		ProjectID:        opts.ProjectID,
		CurrentRunStatus: opts.CurrentRunStatus,
		TagBindings:      opts.TagBindings,
		Include:          opts.Include,
		Sort:             opts.Sort,
	}

	f := func(lo tfe.ListOptions) ([]*Workspace, *tfe.Pagination, error) {
		o.ListOptions = lo
		result, err := s.tfe.Workspaces.List(ctx, org, &o)
		if err != nil {
			return nil, nil, err
		}

		return result.Items, result.Pagination, nil
	}

	current := Pagination{}
	pager := tfepaging.New(f)

	var workspaces []*Workspace
	for i, ws := range pager.All() {
		current.Pagination = *pager.Current()

		if opts.Limit == 0 {
			opts.Limit = 20
		}

		if opts.Limit <= len(workspaces) {
			if i < current.TotalCount {
				current.ReachedLimit = true
			}
			break
		}

		if len(opts.VCSRepos) > 0 {
			if ws.VCSRepo == nil {
				continue
			}

			if !slices.Contains(opts.VCSRepos, ws.VCSRepo.Identifier) {
				continue
			}
		}

		workspaces = append(workspaces, ws)
	}

	if err := pager.Err(); err != nil {
		return nil, nil, err
	}

	return workspaces, &current, nil
}

func (s *WorkspacesService) Update(
	ctx context.Context,
	org string,
	workspace string,
	opts tfe.WorkspaceUpdateOptions,
) (*Workspace, error) {
	return s.tfe.Workspaces.Update(ctx, org, workspace, opts)
}
