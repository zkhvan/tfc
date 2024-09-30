package list

import (
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/pkg/cmdutil"
)

type Options struct {
	TFEClient func() (*tfe.Client, error)

	Organization string
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		TFEClient: f.TFEClient,
	}

	cmd := &cobra.Command{
		Use:   "list organization",
		Short: "List Terraform workspaces",
		Long: heredoc.Doc(`
			List Terraform workspaces.

			Workspaces always belong to a single organization.
		`),
		Example: heredoc.Doc(`
			tfc workspaces list <organization>
		`),
		Aliases:           []string{"ls"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, args []string) {
	opts.Organization = args[0]
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	list, err := client.Workspaces.List(ctx, opts.Organization, &tfe.WorkspaceListOptions{})
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		println(item.Name)
	}

	return nil
}
