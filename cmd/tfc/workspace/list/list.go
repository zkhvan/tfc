package list

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

const MaxPageSize = 100

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
	ColumnRunStatus     string = "RUN_STATUS"
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
	ColumnRunStatus,
}

var (
	TimeStyle             = lipgloss.NewStyle().Faint(true)
	RunStatusPendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	RunStatusErroredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	RunStatusRunningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // blue
	RunStatusHoldingStyle = lipgloss.NewStyle().Faint(true)
	RunStatusAppliedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)
	Clock     *cmdutil.Clock
	Printer   cmdutil.Printer

	Organization      string
	OrganizationExact bool
	Name              string
	Tags              []string
	ExcludeTags       []string
	VCSRepos          []string

	// Filter the results based on various groups of run statuses.
	Pending bool
	Errored bool
	Running bool
	Holding bool
	Applied bool

	Limit          int
	Columns        []string
	ColumnsChanged bool
	WithVariables  []string
}

var (
	DefaultColumns = []string{
		ColumnName,
		ColumnRunStatus,
		ColumnUpdatedAt,
	}
)

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Terraform workspaces",
		Long: text.Heredoc(`
			List Terraform workspaces.

			Workspaces always belong to a single organization.
		`),
		Example: text.IndentHeredoc(2, `
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
	cmd.Flags().StringVarP(&opts.Organization, "org", "o", "", "Search by the organization name.")
	cmd.Flags().StringSliceVarP(&opts.Tags, "tags", "t", []string{}, "Search by the tags.")
	cmd.Flags().StringSliceVarP(&opts.ExcludeTags, "exclude-tags", "T", []string{}, "Search by excluding the tags.")
	cmd.Flags().StringSliceVarP(&opts.VCSRepos, "vcs-repos", "r", []string{}, "Search by the VCS repository name.")

	cmd.Flags().BoolVar(&opts.Pending, "pending", false, "Search for workspaces with pending runs.")
	cmd.Flags().BoolVar(&opts.Errored, "errored", false, "Search for workspaces with errored runs.")
	cmd.Flags().BoolVar(&opts.Running, "running", false, "Search for workspaces with running runs.")
	cmd.Flags().BoolVar(&opts.Holding, "holding", false, "Search for workspaces with holding runs.")
	cmd.Flags().BoolVar(&opts.Applied, "applied", false, "Search for workspaces with applied runs.")

	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 20, "Limit the number of results.")
	cmd.Flags().StringSliceVarP(&opts.WithVariables, "with-variables", "v", []string{},
		"Retrieve workspace variables to display as columns (expensive).",
	)
	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", DefaultColumns, "Columns to show.", ColumnAll)

	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, _ []string) {
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

	opts.ColumnsChanged = cmd.Flags().Changed("columns")
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
	if !opts.ColumnsChanged {
		if len(opts.Organization) == 0 || len(orgs) > 1 {
			opts.Columns = append([]string{"ORG"}, opts.Columns...)
		}
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)

	var errs []error
	for _, org := range orgs {
		o := tfc.WorkspaceListOptions{
			ListOptions: tfc.ListOptions{
				Limit: opts.Limit,
			},
			Search:           opts.Name,
			Tags:             strings.Join(opts.Tags, ","),
			ExcludeTags:      strings.Join(opts.ExcludeTags, ","),
			CurrentRunStatus: opts.runStatus(),
			VCSRepos:         opts.VCSRepos,
		}

		if slices.Contains(opts.Columns, ColumnRunStatus) {
			o.Include = append(o.Include, tfe.WSCurrentRun)
		}

		workspaces, paging, err := client.Workspaces.List(ctx, org.Name, &o)
		if err != nil {
			errs = append(errs, fmt.Errorf("error listing workspaces for %q: %w", org.Name, err))
			continue
		}

		if paging.ReachedLimit {
			fmt.Fprintf(opts.IO.Out, "Showing top %d results for org %q\n\n", opts.Limit, org.Name)
		}

		for _, ws := range workspaces {
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

func (opts *Options) runStatus() string {
	var statuses []tfe.RunStatus

	if opts.Pending {
		statuses = append(statuses, tfc.RunStatusesInGroup(tfc.RunStatusGroupPending)...)
	}
	if opts.Errored {
		statuses = append(statuses, tfc.RunStatusesInGroup(tfc.RunStatusGroupErrored)...)
	}
	if opts.Running {
		statuses = append(statuses, tfc.RunStatusesInGroup(tfc.RunStatusGroupRunning)...)
	}
	if opts.Holding {
		statuses = append(statuses, tfc.RunStatusesInGroup(tfc.RunStatusGroupHolding)...)
	}
	if opts.Applied {
		statuses = append(statuses, tfc.RunStatusesInGroup(tfc.RunStatusGroupApplied)...)
	}

	result := make([]string, 0, len(statuses))
	for _, s := range statuses {
		result = append(result, string(s))
	}
	return strings.Join(result, ",")
}

func (opts *Options) extractWorkspaceFields(ws *tfe.Workspace, wsVars []*tfe.Variable) map[string]string {
	renderTime := func(at time.Time) string {
		rat := text.RelativeTimeAgo(opts.Clock.Now(), at)

		return TimeStyle.Render(rat)
	}

	renderStatus := func(status tfe.RunStatus) string {
		s := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(status))

		return s.Render(string(status))
	}

	v := map[string]string{
		ColumnID:            ws.ID,
		ColumnName:          ws.Name,
		ColumnUpdatedAt:     renderTime(ws.UpdatedAt),
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

	if ws.CurrentRun != nil {
		v[ColumnRunStatus] = renderStatus(ws.CurrentRun.Status)
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

func listWorkspacesVariables(ctx context.Context, client *tfc.Client, id string) (*tfe.VariableList, error) {
	response, err := client.Variables.List(ctx, id, &tfe.VariableListOptions{})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func listOrganizations(
	ctx context.Context,
	client *tfc.Client,
	name string,
	nameExact bool,
) ([]*tfe.Organization, error) {
	if nameExact {
		org, err := client.Organizations.Read(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("get organization: %w", err)
		}

		return []*tfe.Organization{org}, nil
	}

	orgs, _, err := client.Organizations.List(ctx, &tfc.OrganizationListOptions{
		Query: name,
	})
	if err != nil {
		return nil, err
	}

	return orgs, nil
}
