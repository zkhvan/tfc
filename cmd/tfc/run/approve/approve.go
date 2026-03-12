package approve

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
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
	RunID       string
	Comment     string
}

func NewCmdApprove(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "approve [run-id]",
		Short: "Approve a run for apply",
		Long: text.Heredoc(`
			Approve a Terraform run for apply.

			By default, approves the current run for the workspace. You can also
			specify a run ID directly.

			The run must be in an approvable state (e.g., planned, cost_estimated,
			policy_checked). If the run is not in an approvable state, the command
			will exit with an error.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# Approve current run for workspace (using state.tf)
			$ tfc run approve

			# Approve current run for explicit workspace
			$ tfc run approve -W myorg/myworkspace

			# Approve a specific run by ID
			$ tfc run approve run-abc123

			# Approve with a comment
			$ tfc run approve -m "LGTM, approved for production deploy"
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

	cmd.Flags().StringVarP(&opts.Comment, "comment", "m", "", "Comment to include with the approval")

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

	if !tfc.IsApprovable(run.Status) {
		return fmt.Errorf(
			"run %s is in status %q and cannot be approved (must be in a planned/pending state)",
			run.ID, run.Status,
		)
	}

	applyOpts := tfc.RunApplyOptions{}
	if opts.Comment != "" {
		applyOpts.Comment = ptr.String(opts.Comment)
	}

	if err := client.Runs.Apply(ctx, run.ID, applyOpts); err != nil {
		return fmt.Errorf("failed to approve run %s: %w", run.ID, err)
	}

	return opts.displayApproval(run)
}

func (opts *Options) resolveRun(ctx context.Context, client *tfc.Client) (*tfc.Run, error) {
	if opts.RunID != "" {
		run, err := client.Runs.ReadWithOptions(ctx, opts.RunID, &tfc.RunReadOptions{
			Include: []tfe.RunIncludeOpt{tfe.RunPlan},
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

	run, err := client.Runs.ReadWithOptions(ctx, ws.CurrentRun.ID, &tfc.RunReadOptions{
		Include: []tfe.RunIncludeOpt{tfe.RunPlan},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read run %s: %w", ws.CurrentRun.ID, err)
	}

	return run, nil
}

func (opts *Options) displayApproval(run *tfc.Run) error {
	statusStyle := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(run.Status))

	fmt.Fprintf(opts.IO.Out, "Run approved for apply\n\n")
	fmt.Fprintf(opts.IO.Out, "  Run ID:   %s\n", run.ID)
	fmt.Fprintf(opts.IO.Out, "  Status:   %s\n", statusStyle.Render(string(run.Status)))
	if run.Message != "" {
		fmt.Fprintf(opts.IO.Out, "  Message:  %s\n", run.Message)
	}
	if opts.Comment != "" {
		fmt.Fprintf(opts.IO.Out, "  Comment:  %s\n", opts.Comment)
	}

	return nil
}
