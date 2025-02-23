package tfc

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
)

type OrganizationsService service

// Type aliases for convenience
type (
	Organization = tfe.Organization
	Run          = tfe.Run
)

func (s *OrganizationsService) Read(
	ctx context.Context,
	name string,
) (*Organization, error) {
	return s.tfe.Organizations.Read(ctx, name)
}

type OrganizationListOptions struct {
	ListOptions

	// Optional: A query string used to filter organizations. Organizations
	// with a name or email partially matching this value will be returned.
	Query string `url:"q,omitempty"`
}

func (s *OrganizationsService) List(
	ctx context.Context,
	opts *OrganizationListOptions,
) ([]*Organization, *Pagination, error) {
	o := tfe.OrganizationListOptions{
		Query: opts.Query,
	}

	f := func(lo tfe.ListOptions) ([]*Organization, *tfe.Pagination, error) {
		o.ListOptions = lo
		result, err := s.tfe.Organizations.List(ctx, &o)
		if err != nil {
			return nil, nil, err
		}

		return result.Items, result.Pagination, nil
	}

	current := Pagination{}
	pager := tfepaging.New(f)

	var organizations []*Organization
	for i, ws := range pager.All() {
		current.Pagination = *pager.Current()

		if opts.Limit == 0 {
			opts.Limit = 20
		}

		if opts.Limit <= len(organizations) {
			if i < current.TotalCount {
				current.ReachedLimit = true
			}
			break
		}

		organizations = append(organizations, ws)
	}

	if err := pager.Err(); err != nil {
		return nil, nil, err
	}

	return organizations, &current, nil
}

type OrganizationRunListOptions struct {
	ListOptions

	// Optional: Searches runs that matches the supplied VCS username.
	User string `url:"search[user],omitempty"`

	// Optional: Searches runs that matches the supplied commit sha.
	Commit string `url:"search[commit],omitempty"`

	// Optional: Searches runs that matches the supplied VCS username, commit sha, run_id, and run message.
	// The presence of search[commit] or search[user] takes priority over this parameter and will be omitted.
	Search string `url:"search[basic],omitempty"`

	// Optional: Comma-separated list of acceptable run statuses.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-states,
	// or as constants with the RunStatus string type.
	Status string `url:"filter[status],omitempty"`

	// Optional: A single status group. The result lists runs whose status
	// falls under this status group. For details on options, refer to Run
	// status groups.
	// Options are listed at https://developer.hashicorp.com/terraform/enterprise/api-docs/run#run-status-groups,
	// or as constants with the RunStatusGroup string type.
	StatusGroup string `url:"filter[status_group],omitempty"`

	// Optional: Comma-separated list of acceptable run sources.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-sources,
	// or as constants with the RunSource string type.
	Source string `url:"filter[source],omitempty"`

	// Optional: Comma-separated list of acceptable run operation types.
	// Options are listed at https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#run-operations,
	// or as constants with the RunOperation string type.
	Operation string `url:"filter[operation],omitempty"`

	// Optional: Comma-separated list of acceptable agent pool names. The
	// result lists runs that use one of the agent pools you specify.
	AgentPoolNames string `url:"filter[agent_pool_names],omitempty"`

	// Optional: A comma-separated list of workspace names. The result lists
	// runs that belong to one of the workspaces your specify.
	WorkspaceNames string `url:"filter[workspace_names],omitempty"`

	// Optional: A list of relations to include. See available resources:
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/run#available-related-resources
	Include []tfe.RunIncludeOpt `url:"include,omitempty"`
}

func (s *OrganizationsService) ListRuns(
	ctx context.Context,
	org string,
	opts *OrganizationRunListOptions,
) ([]*Run, *Pagination, error) {
	o := *opts

	if o.Limit == 0 {
		o.Limit = 20
	}

	f := func(lo tfe.ListOptions) ([]*Run, *tfe.Pagination, error) {
		o.ListOptions.ListOptions = lo

		u := fmt.Sprintf("organizations/%s/runs", url.PathEscape(org))
		req, err := s.tfe.NewRequest("GET", u, o)
		if err != nil {
			return nil, nil, err
		}

		var rl tfe.RunList
		if err := req.Do(ctx, &rl); err != nil {
			return nil, nil, err
		}

		return rl.Items, rl.Pagination, nil
	}

	current := Pagination{}
	pager := tfepaging.New(f)

	var runs []*tfe.Run
	for i, org := range pager.All() {
		current.Pagination = *pager.Current()

		if o.Limit <= len(runs) {
			if i < current.TotalCount {
				current.ReachedLimit = true
			}
			break
		}

		runs = append(runs, org)
	}

	if err := pager.Err(); err != nil {
		return nil, nil, err
	}

	return runs, &current, nil
}
