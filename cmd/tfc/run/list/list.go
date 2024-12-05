package list

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/term/color"
	"github.com/zkhvan/tfc/pkg/text"
)

var (
	ColumnsDefault = []string{
		ColumnWorkspace,
		ColumnMessage,
		ColumnStatus,
		ColumnCreatedAt,
	}
	ColumnsAll = []string{
		ColumnID,
		ColumnCreatedAt,
		ColumnIsDestroy,
		ColumnHasChanges,
		ColumnMessage,
		ColumnPlanOnly,
		ColumnRefreshOnly,
		ColumnStatus,
		ColumnSource,
		ColumnWorkspace,
	}
)

var (
	TimeStyle = lipgloss.NewStyle().Foreground(color.LightBlack)
)

const (
	ColumnID          string = "ID"
	ColumnCreatedAt   string = "CREATED_AT"
	ColumnIsDestroy   string = "IS_DESTROY"
	ColumnHasChanges  string = "HAS_CHANGES"
	ColumnMessage     string = "MESSAGE"
	ColumnPlanOnly    string = "PLAN_ONLY"
	ColumnRefreshOnly string = "REFRESH_ONLY"
	ColumnStatus      string = "STATUS"
	ColumnSource      string = "SOURCE"
	ColumnWorkspace   string = "WORKSPACE"
)

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Terraform runs",
		Long: text.Heredoc(`
			List Terraform runs.
		`),
		Aliases:           []string{"ls"},
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.Organization, "org", "o", "", "Organization name.")
	cmd.Flags().StringVarP(&opts.Commit, "commit", "C", "", "Commit SHA.")

	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 20, "Limit the number of results.")
	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", ColumnsDefault, "Columns to show.", ColumnsAll)

	_ = cmdutil.MarkFlagsWithNoFileCompletions(cmd)

	return cmd
}

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)
	Clock     *cmdutil.Clock

	Limit        int
	Columns      []string
	Organization string
	Commit       string
}

func (*Options) Complete(_ *cobra.Command, _ []string) {
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	runs, err := listRuns(ctx, client, opts.Organization, &listOptions{
		Limit:  opts.Limit,
		Commit: opts.Commit,
	})
	if err != nil {
		return err
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)
	for _, run := range runs {
		fields := opts.ExtractFields(run)

		p.Write(fields)
	}
	p.Flush()

	return nil
}

func (opts *Options) ExtractFields(run *tfe.Run) map[string]string {
	renderTime := func(at time.Time) string {
		rat := text.RelativeTimeAgo(opts.Clock.Now(), at)

		return TimeStyle.Render(rat)
	}

	renderStatus := func(status tfe.RunStatus) string {
		s := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(status))

		return s.Render(string(status))
	}

	fields := map[string]string{
		ColumnID:        run.ID,
		ColumnCreatedAt: renderTime(run.CreatedAt),
		ColumnIsDestroy: strconv.FormatBool(run.IsDestroy),
		ColumnMessage:   run.Message,
		ColumnStatus:    renderStatus(run.Status),
	}

	if run.Workspace != nil {
		fields[ColumnWorkspace] = run.Workspace.Name
	}

	return fields
}

type OrganizationRunListOptions struct {
	tfe.ListOptions

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

type RunStatusGroup string

const (
	RunGroupNonFinal    RunStatusGroup = "non_final"
	RunGroupFinal       RunStatusGroup = "final"
	RunGroupDiscardable RunStatusGroup = "discardable"
)

type listOptions struct {
	Limit  int
	Commit string
}

func listRuns(
	ctx context.Context,
	client *tfc.Client,
	org string,
	opts *listOptions,
) ([]*tfe.Run, error) {
	f := func(lo tfe.ListOptions) ([]*tfe.Run, *tfe.Pagination, error) {
		o := &OrganizationRunListOptions{
			ListOptions: lo,
			Commit:      opts.Commit,
			Include: []tfe.RunIncludeOpt{
				tfe.RunWorkspace,
			},
		}

		u := fmt.Sprintf("organizations/%s/runs", url.PathEscape(org))
		req, err := client.NewRequest("GET", u, o)
		if err != nil {
			return nil, nil, err
		}

		var rl tfe.RunList
		if err := req.Do(ctx, &rl); err != nil {
			return nil, nil, err
		}

		return rl.Items, rl.Pagination, nil
	}

	pager := tfepaging.New(f)

	var runs []*tfe.Run
	for i, org := range pager.All() {
		if opts.Limit <= i {
			break
		}

		runs = append(runs, org)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}
