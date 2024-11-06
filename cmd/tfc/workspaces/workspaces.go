package workspaces

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/workspaces/list"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdWorkspaces(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspaces",
		Aliases: []string{"ws"},
		Short:   "Manage Terraform workspaces",
		Long: text.Heredoc(`
			Manage Terraform workspaces.

			A workspace groups resources that are managed by Terraform.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))

	return cmd
}
