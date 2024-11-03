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
		Short:   "Interact with Terraform workspaces",
		Long: text.Heredoc(`
			Interact with Terraform workspaces. A workspace is a group of
			infrastructure resources managed by Terraform.

			HCP Terraform manages infrastructure collections with workspaces
			instead of directories. A workspace contains everything Terraform
			needs to manage a given collection of infrastructure, and separate
			workspaces function like completely separate working directories.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))

	return cmd
}
