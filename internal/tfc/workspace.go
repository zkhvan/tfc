package tfc

import (
	"context"
	"slices"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
)

type WorkspacesService service

type Workspace = tfe.Workspace

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

	workspaces := make([]*Workspace, 0)
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
