package list

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/term/color"
	"github.com/zkhvan/tfc/pkg/text"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

var (
	ColumnsDefault = []string{
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
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
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

	cmd.Flags().StringVarP(&opts.Org, "org", "o", "", "Organization name.")
	cmd.Flags().StringVarP(&opts.Workspace, "workspace", "w", "", "Workspace name.")
	cmd.Flags().StringVarP(&opts.Commit, "commit", "C", "", "Commit SHA.")

	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 20, "Limit the number of results.")
	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", ColumnsDefault, "Columns to show.", ColumnsAll)

	_ = cmdutil.MarkFlagsWithNoFileCompletions(cmd)

	return cmd
}

type Options struct {
	IO              *iolib.IOStreams
	TFEClient       func() (*tfc.Client, error)
	Clock           *cmdutil.Clock
	TerraformConfig func() *tfconfig.TerraformConfig

	Limit          int
	Columns        []string
	ColumnsChanged bool
	Commit         string

	Org       string
	Workspace string
}

func (opts *Options) Complete(cmd *cobra.Command, _ []string) {
	orgChanged := cmd.Flags().Changed("org")
	workspaceChanged := cmd.Flags().Changed("workspace")

	if !orgChanged || !workspaceChanged {
		if cfg := opts.TerraformConfig(); cfg != nil && cfg.IsValid() {
			if !orgChanged {
				opts.Org = cfg.Organization
			}
			if !workspaceChanged {
				opts.Workspace = cfg.Workspace.Name
			}
		}
	}

	opts.ColumnsChanged = cmd.Flags().Changed("columns")
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	if opts.Workspace != "" {
		return opts.listWorkspaceRuns(ctx, client)
	}

	return opts.listOrgRuns(ctx, client)
}

func (opts *Options) listWorkspaceRuns(ctx context.Context, client *tfc.Client) error {
	ws, err := client.Workspaces.Read(ctx, opts.Org, opts.Workspace)
	if err != nil {
		return err
	}

	runOpts := &tfc.WorkspaceRunListOptions{
		Limit: opts.Limit,
	}
	runs, paging, err := client.Runs.List(ctx, ws.ID, runOpts)
	if err != nil {
		return err
	}

	return opts.displayRuns(runs, paging, false)
}

func (opts *Options) listOrgRuns(ctx context.Context, client *tfc.Client) error {
	o := tfc.OrganizationRunListOptions{
		ListOptions: tfc.ListOptions{
			Limit: opts.Limit,
		},
		Commit:  opts.Commit,
		Include: []tfe.RunIncludeOpt{tfe.RunWorkspace},
	}
	runs, paging, err := client.Organizations.ListRuns(ctx, opts.Org, &o)
	if err != nil {
		return err
	}

	return opts.displayRuns(runs, paging, true)
}

func (opts *Options) displayRuns(runs []*tfe.Run, paging *tfc.Pagination, showWorkspace bool) error {
	if paging.ReachedLimit {
		fmt.Fprintf(opts.IO.Out, "Showing top %d results\n\n", opts.Limit)
	}

	columns := opts.Columns
	if !opts.ColumnsChanged && showWorkspace {
		columns = append([]string{ColumnWorkspace}, columns...)
	}

	p := cmdutil.FieldPrinter(opts.IO, columns...)
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

	renderMessage := func(msg string) string {
		// Truncate multiline messages
		if idx := strings.Index(msg, "\n"); idx != -1 {
			msg = msg[:idx]
		}
		targetWidth := opts.IO.TerminalWidth() * 7 / 10
		return text.TruncateBounded(msg, targetWidth, 20, 200)
	}

	fields := map[string]string{
		ColumnID:        run.ID,
		ColumnCreatedAt: renderTime(run.CreatedAt),
		ColumnIsDestroy: strconv.FormatBool(run.IsDestroy),
		ColumnMessage:   renderMessage(run.Message),
		ColumnStatus:    renderStatus(run.Status),
	}

	if run.Workspace != nil {
		fields[ColumnWorkspace] = run.Workspace.Name
	}

	return fields
}
