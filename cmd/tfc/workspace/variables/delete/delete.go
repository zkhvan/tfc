package delete

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)

	Org        string
	Workspace  string
	Identifier string // Variable name or ID
}

func NewCmdDelete(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
	}

	cmd := &cobra.Command{
		Use:   "delete <ORG/WORKSPACE> <NAME|ID>",
		Short: "Delete a workspace variable",
		Long: text.Heredoc(`
			Delete a workspace variable.

			The variable can be identified by either its name or ID.
		`),
		Example: text.Heredoc(`
			# Delete a variable by name
			$ tfc workspaces variables delete myorg/myworkspace MY_VAR

			# Delete a variable by ID
			$ tfc workspaces variables delete myorg/myworkspace var-abc123
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(args)
			return opts.Run(cmd.Context())
		},
	}

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

func parse(workspace string) (string, string) {
	parts := strings.Split(workspace, "/")
	return parts[0], parts[1]
}
