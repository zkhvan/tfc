package set

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
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

	Org         string
	Workspace   string
	Identifier  string // Variable name or ID
	Value       string
	Description string
	Category    string
	HCL         bool
	Sensitive   bool
}

func NewCmdSet(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
	}

	cmd := &cobra.Command{
		Use:   "set <ORG/WORKSPACE> <NAME|ID>",
		Short: "Set a workspace variable's value",
		Long: text.Heredoc(`
			Set a workspace variable's value.

			The variable can be identified by either its name or ID. If the
			variable does not exist, it will be created.
		`),
		Example: text.Heredoc(`
			# Set a variable value by name
			$ tfc workspaces variables set myorg/myworkspace MY_VAR --value "new-value"

			# Set a sensitive variable
			$ tfc workspaces variables set myorg/myworkspace AWS_SECRET --value "secret" --sensitive

			# Set an HCL variable
			$ tfc workspaces variables set myorg/myworkspace config --value '{"key": "value"}' --hcl

			# Set a variable by ID
			$ tfc workspaces variables set myorg/myworkspace var-abc123 --value "new-value"
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.Value, "value", "v", "", "Variable value (required)")
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Variable description")
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "terraform", "Variable category: terraform or env")
	cmd.Flags().BoolVar(&opts.HCL, "hcl", false, "Parse the value as HCL")
	cmd.Flags().BoolVar(&opts.Sensitive, "sensitive", false, "Mark the variable as sensitive")

	_ = cmd.MarkFlagRequired("value")

	return cmd
}

func (opts *Options) Complete(args []string) {
	opts.Org, opts.Workspace = parse(args[0])
	opts.Identifier = args[1]
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	ws, err := client.Workspaces.Read(ctx, opts.Org, opts.Workspace)
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

func parse(workspace string) (string, string) {
	parts := strings.Split(workspace, "/")
	return parts[0], parts[1]
}
