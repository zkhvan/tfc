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
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)
	Clock     *cmdutil.Clock

	Org       string
	Workspace string

	Message     string
	PlanOnly    bool
	IsDestroy   bool
	RefreshOnly bool
}

func NewCmdTrigger(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "trigger <ORG/WORKSPACE>",
		Short: "Trigger a new Terraform run",
		Long: text.Heredoc(`
			Trigger a new Terraform run in a workspace.

			By default, creates a standard plan-and-apply run that requires
			manual approval. Use flags to create different run types.
		`),
		Example: text.Heredoc(`
			# Create a standard run
			$ tfc run trigger myorg/myworkspace

			# Create a run with a custom message
			$ tfc run trigger myorg/myworkspace -m "Deploying version 2.0"

			# Create a speculative plan-only run
			$ tfc run trigger myorg/myworkspace --plan-only

			# Create a destroy run
			$ tfc run trigger myorg/myworkspace --destroy

			# Create a refresh-only run
			$ tfc run trigger myorg/myworkspace --refresh-only
		`),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompletionOrgWorkspace(opts.TFEClient),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.Message, "message", "m", "", "Run message (default: \"Triggered via CLI\")")
	cmd.Flags().BoolVar(&opts.PlanOnly, "plan-only", false, "Create a speculative plan-only run")
	cmd.Flags().BoolVar(&opts.IsDestroy, "destroy", false, "Create a destroy run")
	cmd.Flags().BoolVar(&opts.RefreshOnly, "refresh-only", false, "Create a refresh-only run")

	cmd.MarkFlagsMutuallyExclusive("plan-only", "destroy", "refresh-only")
	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(args []string) {
	parsed := tfc.ParseOrgWorkspace(args[0])
	opts.Org = parsed.Org
	opts.Workspace = parsed.Workspace

	if opts.Message == "" {
		opts.Message = "Triggered via CLI"
	}
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return fmt.Errorf("failed to initialize TFE client: %w", err)
	}

	ws, err := client.Workspaces.Read(ctx, opts.Org, opts.Workspace)
	if err != nil {
		return fmt.Errorf("failed to read workspace %s/%s: %w", opts.Org, opts.Workspace, err)
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
	url := buildRunURL(opts.Org, opts.Workspace, run.ID)

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
