package view

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

type Options struct {
	IO              *iolib.IOStreams
	TFEClient       func() (*tfc.Client, error)
	Clock           *cmdutil.Clock
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	RunID       string
	Web         bool
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "view [run-id]",
		Short: "View run details and plan output",
		Long: text.Heredoc(`
			View detailed information about a Terraform run, including the plan
			output.

			By default, shows the current run for the workspace. You can also
			specify a run ID directly.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# View current run for workspace (using state.tf)
			$ tfc run view

			# View current run for explicit workspace
			$ tfc run view -W myorg/myworkspace

			# View a specific run by ID
			$ tfc run view run-abc123

			# Open run in browser
			$ tfc run view --web
		`),
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.RunID = args[0]
			}
			opts.Complete(cmd)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	cmd.Flags().BoolVarP(&opts.Web, "web", "w", false, "Open run in web browser")

	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command) {
	if opts.RunID == "" {
		cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)
	}
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return fmt.Errorf("failed to initialize TFE client: %w", err)
	}

	run, err := opts.resolveRun(ctx, client)
	if err != nil {
		return err
	}

	if opts.Web {
		return opts.openRunInBrowser(ctx, run)
	}

	return opts.displayRun(ctx, client, run)
}

func (opts *Options) resolveRun(ctx context.Context, client *tfc.Client) (*tfc.Run, error) {
	if opts.RunID != "" {
		run, err := client.Runs.ReadWithOptions(ctx, opts.RunID, &tfc.RunReadOptions{
			Include: []tfe.RunIncludeOpt{tfe.RunPlan, tfe.RunWorkspace},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read run %s: %w", opts.RunID, err)
		}
		return run, nil
	}

	if err := opts.WorkspaceID.Validate(); err != nil {
		return nil, fmt.Errorf(
			"run ID or workspace required: use [run-id] argument, -W ORG/WORKSPACE, or ensure state.tf exists",
		)
	}

	readOpts := &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{tfe.WSCurrentRun},
	}
	ws, err := client.Workspaces.ReadWithOptions(
		ctx, opts.WorkspaceID.Org, opts.WorkspaceID.Workspace, readOpts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace %s: %w", opts.WorkspaceID.String(), err)
	}

	if ws.CurrentRun == nil {
		return nil, fmt.Errorf("no current run for workspace %s", opts.WorkspaceID.String())
	}

	// Re-read the run with plan included
	run, err := client.Runs.ReadWithOptions(ctx, ws.CurrentRun.ID, &tfc.RunReadOptions{
		Include: []tfe.RunIncludeOpt{tfe.RunPlan, tfe.RunWorkspace},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read run %s: %w", ws.CurrentRun.ID, err)
	}

	return run, nil
}

func (opts *Options) displayRun(ctx context.Context, client *tfc.Client, run *tfc.Run) error {
	out := opts.IO.Out
	faintStyle := lipgloss.NewStyle().Faint(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	statusStyle := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(run.Status))

	runType := "Plan and apply"
	if run.PlanOnly {
		runType = "Plan only (speculative)"
	} else if run.IsDestroy {
		runType = "Destroy"
	} else if run.RefreshOnly {
		runType = "Refresh only"
	}

	// Run details section
	fmt.Fprintf(out, "%s\n", headerStyle.Render("RUN"))
	fmt.Fprintf(out, "  ID:       %s\n", faintStyle.Render(run.ID))
	fmt.Fprintf(out, "  Status:   %s\n", statusStyle.Render(string(run.Status)))
	fmt.Fprintf(out, "  Type:     %s\n", runType)
	if run.Message != "" {
		fmt.Fprintf(out, "  Message:  %s\n", run.Message)
	}
	createdTime := text.RelativeTimeAgo(opts.Clock.Now(), run.CreatedAt)
	fmt.Fprintf(out, "  Created:  %s\n", createdTime)

	// Plan summary section
	if run.Plan != nil {
		fmt.Fprintf(out, "\n%s\n", headerStyle.Render("PLAN SUMMARY"))
		fmt.Fprintf(out, "  Status:        %s\n", string(run.Plan.Status))

		hasResources := run.Plan.ResourceAdditions > 0 ||
			run.Plan.ResourceChanges > 0 ||
			run.Plan.ResourceDestructions > 0 ||
			run.Plan.ResourceImports > 0
		if hasResources {
			addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
			changeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
			destroyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

			fmt.Fprintf(out, "  Resources:     %s to add, %s to change, %s to destroy",
				addStyle.Render(fmt.Sprintf("%d", run.Plan.ResourceAdditions)),
				changeStyle.Render(fmt.Sprintf("%d", run.Plan.ResourceChanges)),
				destroyStyle.Render(fmt.Sprintf("%d", run.Plan.ResourceDestructions)),
			)
			if run.Plan.ResourceImports > 0 {
				fmt.Fprintf(out, ", %d to import", run.Plan.ResourceImports)
			}
			fmt.Fprintln(out)
		} else if !run.Plan.HasChanges {
			fmt.Fprintf(out, "  Resources:     No changes\n")
		}
	}

	// URL
	url := buildRunURL(run)
	if url != "" {
		fmt.Fprintf(out, "\n  URL:      %s\n", url)
	}

	// Plan logs
	if run.Plan != nil && run.Plan.ID != "" {
		fmt.Fprintf(out, "\n%s\n", headerStyle.Render("PLAN OUTPUT"))

		logs, err := client.Plans.Logs(ctx, run.Plan.ID)
		if err != nil {
			fmt.Fprintf(out, "  (unable to retrieve plan logs: %v)\n", err)
			return nil
		}

		if _, err := io.Copy(out, logs); err != nil {
			return fmt.Errorf("failed to read plan logs: %w", err)
		}
	}

	return nil
}

func buildRunURL(run *tfc.Run) string {
	if run.Workspace == nil {
		return ""
	}
	hostname := os.Getenv("TFE_HOSTNAME")
	if hostname == "" {
		hostname = "app.terraform.io"
	}
	org := ""
	if run.Workspace.Organization != nil {
		org = run.Workspace.Organization.Name
	}
	return fmt.Sprintf("https://%s/app/%s/workspaces/%s/runs/%s",
		hostname, org, run.Workspace.Name, run.ID)
}

func (opts *Options) openRunInBrowser(ctx context.Context, run *tfc.Run) error {
	url := buildRunURL(run)
	if url == "" {
		return fmt.Errorf("unable to determine run URL")
	}

	fmt.Fprintf(opts.IO.Out, "Opening run in browser:\n")
	fmt.Fprintf(opts.IO.Out, "%s\n", url)

	var browserCmd string
	switch runtime.GOOS {
	case "darwin":
		browserCmd = "open"
	case "windows":
		browserCmd = "start"
	default:
		browserCmd = "xdg-open"
	}

	cmd := exec.CommandContext(ctx, browserCmd, url)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(opts.IO.ErrOut, "Warning: Could not open browser: %v\n", err)
		fmt.Fprintf(opts.IO.Out, "\nPlease open the URL manually: %s\n", url)
		return nil
	}

	return nil
}
