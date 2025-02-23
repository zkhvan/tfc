package variables

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/workspace/variables/list"
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

	return cmd
}
