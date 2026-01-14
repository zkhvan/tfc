package delete

import (
	"context"
	"fmt"

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
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	Identifier  string // Variable name or ID
}

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "delete <NAME|ID>",
		Short: "Delete a workspace variable",
		Long: text.Heredoc(`
			Delete a workspace variable.

			The variable can be identified by either its name or ID.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# Delete a variable by name
			$ tfc workspaces variables delete MY_VAR

			# Delete a variable with explicit org/workspace
			$ tfc workspaces variables delete MY_VAR -W myorg/myworkspace

			# Delete a variable by ID
			$ tfc workspaces variables delete var-abc123
		`),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompletionVariableNamesFromWorkspaceFlag(opts.TFEClient, opts.TerraformConfig),
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
	opts.Identifier = args[0]
	cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)
}

func (opts *Options) Run(ctx context.Context) error {
	if err := opts.WorkspaceID.Validate(); err != nil {
		return fmt.Errorf("workspace required: use -W ORG/WORKSPACE or ensure state.tf exists")
	}

	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	ws, err := client.Workspaces.Read(ctx, opts.WorkspaceID.Org, opts.WorkspaceID.Workspace)
	if err != nil {
		return err
	}

	// Try to find the variable by name or ID
	vars, _, err := client.Variables.List(ctx, ws.ID, &tfc.VariableListOptions{ListOptions: tfc.ListOptions{Limit: 1000}})
	if err != nil {
		return err
	}

	var targetVar *tfe.Variable
	for _, v := range vars {
		if v.ID == opts.Identifier || v.Key == opts.Identifier {
			targetVar = v
			break
		}
	}

	if targetVar == nil {
		return fmt.Errorf("variable %q not found", opts.Identifier)
	}

	err = client.Variables.Delete(ctx, ws.ID, targetVar.ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "Variable %q deleted successfully\n", targetVar.Key)

	return nil
}
