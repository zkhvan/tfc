package run

import (
	"github.com/spf13/cobra"

	approveCmd "github.com/zkhvan/tfc/cmd/tfc/run/approve"
	listCmd "github.com/zkhvan/tfc/cmd/tfc/run/list"
	triggerCmd "github.com/zkhvan/tfc/cmd/tfc/run/trigger"
	viewCmd "github.com/zkhvan/tfc/cmd/tfc/run/view"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdRun(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage Terraform runs",
		Long: text.Heredoc(`
			Manage Terraform runs.
		`),
	}

	cmd.AddCommand(approveCmd.NewCmdApprove(f))
	cmd.AddCommand(listCmd.NewCmdList(f))
	cmd.AddCommand(triggerCmd.NewCmdTrigger(f))
	cmd.AddCommand(viewCmd.NewCmdView(f))

	return cmd
}
