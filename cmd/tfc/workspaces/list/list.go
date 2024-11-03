package list

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

const MAX_PAGE_SIZE = 100

const (
	ColumnID            string = "ID"
	ColumnName          string = "NAME"
	ColumnOrg           string = "ORG"
	ColumnUpdatedAt     string = "UPDATED_AT"
	ColumnVCSRepo       string = "VCS_REPO"
	ColumnVCSRepoURL    string = "VCS_REPO_URL"
	ColumnTFVersion     string = "TF_VERSION"
	ColumnResourceCount string = "RESOURCE_COUNT"
	ColumnWorkingDir    string = "WORKING_DIR"
)

var ColumnAll = []string{
	ColumnID,
	ColumnName,
	ColumnOrg,
	ColumnUpdatedAt,
	ColumnVCSRepo,
	ColumnVCSRepoURL,
	ColumnResourceCount,
	ColumnTFVersion,
	ColumnWorkingDir,
}

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfe.Client, error)
	Clock     *cmdutil.Clock
	Printer   cmdutil.Printer

	Organization      string
	OrganizationExact bool
	Name              string
	Limit             int
	Columns           []string
	WithVariables     []string
}

var (
	DefaultColumns = []string{
		string(ColumnName),
		string(ColumnUpdatedAt),
	}
)

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "list organization",
		Short: "List Terraform workspaces",
		Long: heredoc.Doc(`
			List Terraform workspaces.

			Workspaces always belong to a single organization.
		`),
		Example: heredoc.Doc(`
			# List the workspaces in all organizations
			tfc workspaces list

			# List the workspaces in one organization
			tfc workspaces list --org example-org
		`),
		Aliases:           []string{"ls"},
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Search by the workspace name.")
	cmd.RegisterFlagCompletionFunc("name", cobra.NoFileCompletions)

	cmd.Flags().StringVarP(&opts.Organization, "org", "o", "", "Search by the organization name.")
	cmd.RegisterFlagCompletionFunc("org", cobra.NoFileCompletions)

	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Limit the number of results.")
	cmd.RegisterFlagCompletionFunc("limit", cobra.NoFileCompletions)

	cmd.Flags().StringSliceVarP(&opts.Columns, "columns", "c", DefaultColumns, "Columns to show.")
	cmd.RegisterFlagCompletionFunc("columns", cmdutil.GenerateOptionCompletionFunc(ColumnAll))

	cmd.Flags().StringSliceVar(&opts.WithVariables, "with-variables", []string{}, "Retrieve workspace variables to display as columns (expensive).")
	cmd.RegisterFlagCompletionFunc("with-variables", cobra.NoFileCompletions)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, args []string) {
	// Check if there's any wildcard characters
	if strings.ContainsAny(opts.Organization, "%*") {
		opts.Organization = strings.ReplaceAll(opts.Organization, "*", "%")
	}

	if len(opts.Organization) > 0 && !strings.Contains(opts.Organization, "%") {
		opts.OrganizationExact = true
	}

	if len(opts.WithVariables) > 0 {
		opts.Columns = append(opts.Columns, opts.WithVariables...)
	}
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	// Filter the results to the organizations that the user has access to.
	orgs, err := listOrganizations(ctx, client, opts.Organization, opts.OrganizationExact)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		return fmt.Errorf("no matching organizations")
	}

	// Add ORG column intelligently:
	//   - if the user customized the columns, don't add the column
	//   - if the user specified an organization, chances are they know what
	//     they're looking for, only add the column if they end up with
	//     results that have more than one organization
	if slices.Equal(opts.Columns, DefaultColumns) {
		if len(opts.Organization) == 0 || len(orgs) > 1 {
			opts.Columns = append([]string{"ORG"}, opts.Columns...)
		}
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)

	var errs []error
	limit := opts.Limit
	for _, org := range orgs {
		result, err := listWorkspaces(ctx, client, org.Name, &ListOptions{
			Name:  opts.Name,
			Limit: limit,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("error listing workspaces for %q: %w", org.Name, err))
			continue
		}

		if len(result.Items) < result.TotalCount {
			fmt.Fprintf(opts.IO.Out, "Showing %d of %d results for org %q\n", opts.Limit, result.TotalCount, org.Name)
		}

		limit -= len(result.Items)

		for _, ws := range result.Items {
			var wsVars []*tfe.Variable
			if len(opts.WithVariables) > 0 {
				vars, err := listWorkspacesVariables(ctx, client, ws.ID)
				if err != nil {
					errs = append(errs, fmt.Errorf("error retrieving workspace variables for %q: %w", ws.ID, err))
					continue
				}

				wsVars = append(wsVars, vars.Items...)
			}

			fields := opts.extractWorkspaceFields(ws, wsVars)
			p.Write(fields)
		}
	}

	p.Flush()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (opts *Options) extractWorkspaceFields(ws *tfe.Workspace, wsVars []*tfe.Variable) map[string]string {
	v := map[string]string{
		ColumnID:            ws.ID,
		ColumnName:          ws.Name,
		ColumnUpdatedAt:     text.RelativeTimeAgo(opts.Clock.Now(), ws.UpdatedAt),
		ColumnWorkingDir:    ws.WorkingDirectory,
		ColumnTFVersion:     ws.TerraformVersion,
		ColumnResourceCount: strconv.Itoa(ws.ResourceCount),
	}

	if ws.Organization != nil {
		v[ColumnOrg] = ws.Organization.Name
	}

	if ws.VCSRepo != nil {
		v[ColumnVCSRepo] = ws.VCSRepo.DisplayIdentifier
		v[ColumnVCSRepoURL] = ws.VCSRepo.RepositoryHTTPURL
	}

	if len(opts.WithVariables) > 0 {
		wsVarsMap := make(map[string]*tfe.Variable, 0)
		for _, wsVar := range wsVars {
			wsVarsMap[wsVar.Key] = wsVar
		}

		for _, key := range opts.WithVariables {
			if wsVar, ok := wsVarsMap[key]; ok {
				v[key] = wsVar.Value
			}
		}
	}

	return v
}

func filterOrganizations(orgs *tfe.OrganizationList, name string) *tfe.OrganizationList {
	result := tfe.OrganizationList{
		Pagination: orgs.Pagination,
	}
	for _, org := range orgs.Items {
		if org.Name == name {
			result.Items = append(result.Items, org)
		}
	}
	return &result
}

type ListOptions struct {
	Name  string
	Limit int
}

func listWorkspaces(ctx context.Context, client *tfe.Client, org string, opts *ListOptions) (*tfe.WorkspaceList, error) {
	listOpts := &tfe.WorkspaceListOptions{
		Search: opts.Name,
	}

	if opts.Limit < MAX_PAGE_SIZE {
		listOpts.PageSize = opts.Limit
		return client.Workspaces.List(ctx, org, listOpts)
	} else {
		listOpts.PageSize = MAX_PAGE_SIZE
	}

	list := &tfe.WorkspaceList{
		Pagination: &tfe.Pagination{},
		Items:      make([]*tfe.Workspace, 0, opts.Limit),
	}
	for count := 0; count < opts.Limit; {
		response, err := client.Workspaces.List(ctx, org, listOpts)
		if err != nil {
			return nil, err
		}
		list.Pagination = response.Pagination

		// Take the necessary amount of items to reach the limit.
		n := min(
			opts.Limit-count,
			len(response.Items),
		)
		count += n
		list.Items = append(list.Items, response.Items[:n]...)

		// Check if there's a next page.
		if list.NextPage == 0 {
			break
		}
		listOpts.PageNumber = list.NextPage
	}

	return list, nil
}

func listWorkspacesVariables(ctx context.Context, client *tfe.Client, id string) (*tfe.VariableList, error) {
	response, err := client.Variables.List(ctx, id, &tfe.VariableListOptions{})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func listOrganizations(ctx context.Context, client *tfe.Client, name string, nameExact bool) ([]*tfe.Organization, error) {
	if nameExact {
		org, err := client.Organizations.Read(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("get organization: %w", err)
		}

		return []*tfe.Organization{org}, nil
	}

	listOpts := tfe.OrganizationListOptions{
		Query: name,
		ListOptions: tfe.ListOptions{
			PageSize: MAX_PAGE_SIZE,
		},
	}

	list := &tfe.OrganizationList{
		Pagination: &tfe.Pagination{},
		Items:      make([]*tfe.Organization, 0),
	}
	for {
		response, err := client.Organizations.List(ctx, &listOpts)
		if err != nil {
			return nil, fmt.Errorf("list organizations: %w", err)
		}

		list.Pagination = response.Pagination
		list.Items = append(list.Items, response.Items...)

		if list.NextPage == 0 {
			break
		}
		listOpts.PageNumber = list.NextPage
	}

	return list.Items, nil
}
