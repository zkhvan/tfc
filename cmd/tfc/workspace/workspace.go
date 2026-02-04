package workspace

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/list"
	updatebranchCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/updatebranch"
	variablesCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/variables"
	viewCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/view"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdWorkspace(f *cmdutil.Factory) *cobra.Command {
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
	cmd.AddCommand(updatebranchCmd.NewCmdUpdateBranch(f))
	cmd.AddCommand(variablesCmd.NewCmdVariables(f))
	cmd.AddCommand(viewCmd.NewCmdView(f))

	return cmd
}
