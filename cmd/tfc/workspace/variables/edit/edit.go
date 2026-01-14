package edit

import (
	"context"
	"fmt"
	"os"

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
	Editor          func() *cmdutil.Editor
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	Identifier  string // Variable name or ID
}

func NewCmdEdit(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Editor:          f.Editor,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "edit <NAME|ID>",
		Short: "Edit a workspace variable interactively",
		Long: text.Heredoc(`
			Edit a workspace variable interactively.

			The variable can be identified by either its name or ID. The current
			value will be loaded into a temporary file and opened in your
			preferred editor. After saving and closing the editor, the variable
			will be updated with the new contents.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
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

	// Create a temporary directory to isolate the file from LSP confusion
	tempDir, err := os.MkdirTemp("", "tfc-variable-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Use .tf extension for HCL variables to enable syntax highlighting
	ext := "txt"
	if targetVar.HCL {
		ext = "hcl"
	}

	pattern := fmt.Sprintf("variable-*.%s", ext)
	file, err := os.CreateTemp(tempDir, pattern)
	if err != nil {
		return err
	}

	if _, err = file.WriteString(targetVar.Value); err != nil {
		file.Close()
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}

	if targetVar.Sensitive && targetVar.Value == "" {
		fmt.Fprintf(
			opts.IO.Out,
			"Variable %q is marked sensitive; the editor will start empty. Provide a new value.\n",
			targetVar.Key,
		)
	}

	if err = opts.Editor().Edit(ctx, file.Name()); err != nil {
		return fmt.Errorf("failed to launch editor: %w", err)
	}

	updatedBytes, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}
	updatedValue := string(updatedBytes)

	if updatedValue == targetVar.Value {
		fmt.Fprintf(opts.IO.Out, "No changes made to variable %q\n", targetVar.Key)
		return nil
	}

	category := targetVar.Category
	updateOpts := tfe.VariableUpdateOptions{
		Value:     ptr.String(updatedValue),
		HCL:       ptr.Bool(targetVar.HCL),
		Sensitive: ptr.Bool(targetVar.Sensitive),
		Category:  &category,
	}

	if _, err := client.Variables.Update(ctx, ws.ID, targetVar.ID, updateOpts); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "Variable %q updated successfully\n", targetVar.Key)

	return nil
}
