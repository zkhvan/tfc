package runs

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/runs/list"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdRuns(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Manage Terraform runs",
		Long: text.Heredoc(`
			Manage Terraform runs.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))

	return cmd
}
