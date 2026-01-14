package set

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
	Identifier  string // Variable name or ID
	Value       string
	Description string
	Category    string
	HCL         bool
	Sensitive   bool
}

func NewCmdSet(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "set <NAME|ID>",
		Short: "Set a workspace variable's value",
		Long: text.Heredoc(`
			Set a workspace variable's value.

			The variable can be identified by either its name or ID. If the
			variable does not exist, it will be created.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# Set a variable value by name
			$ tfc workspaces variables set MY_VAR --value "new-value"

			# Set a variable with explicit org/workspace
			$ tfc workspaces variables set MY_VAR -W myorg/myworkspace --value "new-value"

			# Set a sensitive variable
			$ tfc workspaces variables set AWS_SECRET --value "secret" --sensitive

			# Set an HCL variable
			$ tfc workspaces variables set config --value '{"key": "value"}' --hcl

			# Set a variable by ID
			$ tfc workspaces variables set var-abc123 --value "new-value"
		`),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompletionVariableNamesFromWorkspaceFlag(opts.TFEClient, opts.TerraformConfig),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	cmd.Flags().StringVarP(&opts.Value, "value", "v", "", "Variable value (required)")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Variable description")
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "terraform", "Variable category: terraform or env")
	cmd.Flags().BoolVar(&opts.HCL, "hcl", false, "Parse the value as HCL")
	cmd.Flags().BoolVar(&opts.Sensitive, "sensitive", false, "Mark the variable as sensitive")

	_ = cmd.MarkFlagRequired("value")
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

	var existingVar *tfc.Variable
	for _, v := range vars {
		if v.ID == opts.Identifier || v.Key == opts.Identifier {
			existingVar = v
			break
		}
	}

	category := tfe.CategoryTerraform
	if opts.Category == "env" {
		category = tfe.CategoryEnv
	}

	if existingVar != nil {
		updateOpts := tfe.VariableUpdateOptions{
			Value:     ptr.String(opts.Value),
			HCL:       ptr.Bool(opts.HCL),
			Sensitive: ptr.Bool(opts.Sensitive),
			Category:  &category,
		}

		if opts.Description != "" {
			updateOpts.Description = ptr.String(opts.Description)
		}

		updatedVar, err := client.Variables.Update(ctx, ws.ID, existingVar.ID, updateOpts)
		if err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.Out, "Variable %q updated successfully\n", updatedVar.Key)
	} else {
		createOpts := tfe.VariableCreateOptions{
			Key:       ptr.String(opts.Identifier),
			Value:     ptr.String(opts.Value),
			Category:  &category,
			HCL:       ptr.Bool(opts.HCL),
			Sensitive: ptr.Bool(opts.Sensitive),
		}

		if opts.Description != "" {
			createOpts.Description = ptr.String(opts.Description)
		}

		newVar, err := client.Variables.Create(ctx, ws.ID, createOpts)
		if err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.Out, "Variable %q created successfully\n", newVar.Key)
	}

	return nil
}
