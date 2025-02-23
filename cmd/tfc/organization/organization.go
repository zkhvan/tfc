package organization

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/organization/list"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdOrganization(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs"},
		Short:   "Manage Terraform organizations",
		Long: text.Heredoc(`
			Manage Terraform organizations.

			Organizations are the top-level entities that encompass managed
			Terraform resources.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))

	return cmd
}
