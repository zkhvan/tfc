package variables

import (
	"github.com/spf13/cobra"

	deleteCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/variables/delete"
	listCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/variables/list"
	setCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/variables/set"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdVariables(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variables",
		Short:   "Manage a workspace's variables",
		Aliases: []string{"vars"},
		Long: text.Heredoc(`
			Manage a workspace's variables.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))
	cmd.AddCommand(setCmd.NewCmdSet(f))
	cmd.AddCommand(deleteCmd.NewCmdDelete(f))

	return cmd
}
