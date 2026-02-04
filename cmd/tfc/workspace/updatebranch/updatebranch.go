package updatebranch

import (
	"context"
	"fmt"

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
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	Branch      string
}

func NewCmdUpdateBranch(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:     "update-vcs-branch <branch>",
		Aliases: []string{"update-branch"},
		Short:   "Update the VCS branch of a workspace",
		Long: text.Heredoc(`
			Update the VCS branch that a workspace is connected to.

			This command updates the branch that Terraform Cloud uses to
			trigger runs and fetch configuration. The workspace must have
			VCS configured for this command to work.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# Update branch for workspace specified by flag
			$ tfc workspaces update-vcs-branch main -W myorg/myworkspace

			# Update branch using workspace from state.tf
			$ tfc workspaces update-vcs-branch feature-branch

			# Using alias
			$ tfc ws update-branch develop -W myorg/prod-app
		`),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	_ = cmdutil.MarkFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, args []string) {
	opts.Branch = args[0]
	cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)
}

func (opts *Options) Run(ctx context.Context) error {
	if err := opts.WorkspaceID.Validate(); err != nil {
		return fmt.Errorf("workspace required: use -W ORG/WORKSPACE or ensure state.tf exists")
	}

	if opts.Branch == "" {
		return fmt.Errorf("branch cannot be empty")
	}

	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	// Read the workspace first to check if it has VCS configured
	ws, err := client.Workspaces.Read(
		ctx,
		opts.WorkspaceID.Org,
		opts.WorkspaceID.Workspace,
	)
	if err != nil {
		return fmt.Errorf("failed to read workspace %s: %w", opts.WorkspaceID.String(), err)
	}

	if ws.VCSRepo == nil {
		return fmt.Errorf("workspace %q does not have VCS configured", ws.Name)
	}

	// Update the workspace with the new VCS branch
	updateOpts := tfe.WorkspaceUpdateOptions{
		VCSRepo: &tfe.VCSRepoOptions{
			Branch: ptr.String(opts.Branch),
		},
	}

	updatedWS, err := client.Workspaces.Update(
		ctx,
		opts.WorkspaceID.Org,
		opts.WorkspaceID.Workspace,
		updateOpts,
	)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Fprintf(
		opts.IO.Out,
		"Updated VCS branch for workspace %q to %q\n",
		updatedWS.Name,
		opts.Branch,
	)

	return nil
}
