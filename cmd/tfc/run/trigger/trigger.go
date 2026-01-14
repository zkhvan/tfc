package trigger

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/ptr"
	"github.com/zkhvan/tfc/pkg/text"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

type Options struct {
	IO              *iolib.IOStreams
	TFEClient       func() (*tfc.Client, error)
	Clock           *cmdutil.Clock
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier

	Message     string
	PlanOnly    bool
	IsDestroy   bool
	RefreshOnly bool
}

func NewCmdTrigger(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger a new Terraform run",
		Long: text.Heredoc(`
			Trigger a new Terraform run in a workspace.

			By default, creates a standard plan-and-apply run that requires
			manual approval. Use flags to create different run types.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# Create a standard run (using state.tf)
			$ tfc run trigger

			# Create a run with explicit workspace
			$ tfc run trigger -W myorg/myworkspace

			# Create a run with a custom message
			$ tfc run trigger -W myorg/myworkspace -m "Deploying version 2.0"

			# Create a speculative plan-only run
			$ tfc run trigger --plan-only

			# Create a destroy run
			$ tfc run trigger --destroy

			# Create a refresh-only run
			$ tfc run trigger --refresh-only
		`),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Complete(cmd)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	cmd.Flags().StringVarP(&opts.Message, "message", "m", "", "Run message (default: \"Triggered via CLI\")")
	cmd.Flags().BoolVar(&opts.PlanOnly, "plan-only", false, "Create a speculative plan-only run")
	cmd.Flags().BoolVar(&opts.IsDestroy, "destroy", false, "Create a destroy run")
	cmd.Flags().BoolVar(&opts.RefreshOnly, "refresh-only", false, "Create a refresh-only run")

	cmd.MarkFlagsMutuallyExclusive("plan-only", "destroy", "refresh-only")
	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command) {
	cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)

	if opts.Message == "" {
		opts.Message = "Triggered via CLI"
	}
}

func (opts *Options) Run(ctx context.Context) error {
	if err := opts.WorkspaceID.Validate(); err != nil {
		return fmt.Errorf("workspace required: use -W ORG/WORKSPACE or ensure state.tf exists")
	}

	client, err := opts.TFEClient()
	if err != nil {
		return fmt.Errorf("failed to initialize TFE client: %w", err)
	}

	ws, err := client.Workspaces.Read(ctx, opts.WorkspaceID.Org, opts.WorkspaceID.Workspace)
	if err != nil {
		return fmt.Errorf("failed to read workspace %s: %w", opts.WorkspaceID.String(), err)
	}

	runOpts := tfc.RunCreateOptions{
		Workspace: &tfc.Workspace{ID: ws.ID},
		Message:   ptr.String(opts.Message),
	}

	if opts.PlanOnly {
		runOpts.PlanOnly = ptr.Bool(true)
	}
	if opts.IsDestroy {
		runOpts.IsDestroy = ptr.Bool(true)
	}
	if opts.RefreshOnly {
		runOpts.RefreshOnly = ptr.Bool(true)
	}

	run, err := client.Runs.Create(ctx, runOpts)
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}

	return opts.displayRun(run)
}

func (opts *Options) displayRun(run *tfc.Run) error {
	statusStyle := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(run.Status))

	runType := "Plan and apply"
	if run.PlanOnly {
		runType = "Plan only (speculative)"
	} else if run.IsDestroy {
		runType = "Destroy"
	} else if run.RefreshOnly {
		runType = "Refresh only"
	}

	// Build and display URL
	url := buildRunURL(opts.WorkspaceID.Org, opts.WorkspaceID.Workspace, run.ID)

	fmt.Fprintf(opts.IO.Out, "Run created successfully\n\n")
	fmt.Fprintf(opts.IO.Out, "  Run ID:     %s\n", run.ID)
	fmt.Fprintf(opts.IO.Out, "  Message:    %s\n", run.Message)
	fmt.Fprintf(opts.IO.Out, "  Status:     %s\n", statusStyle.Render(string(run.Status)))
	fmt.Fprintf(opts.IO.Out, "  Type:       %s\n", runType)
	fmt.Fprintf(opts.IO.Out, "\n  View run:   %s\n", url)

	return nil
}

func buildRunURL(org, workspace, runID string) string {
	hostname := os.Getenv("TFE_HOSTNAME")
	if hostname == "" {
		hostname = "app.terraform.io"
	}
	return fmt.Sprintf("https://%s/app/%s/workspaces/%s/runs/%s",
		hostname, org, workspace, runID)
}
