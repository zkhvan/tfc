package organizations

import (
	"github.com/spf13/cobra"

	listCmd "github.com/zkhvan/tfc/cmd/tfc/organizations/list"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdOrganizations(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "organizations",
		Aliases: []string{"orgs"},
		Short:   "Interact with Terraform organizations",
		Long: text.Heredoc(`
			Interact with Terraform organizations.
		`),
	}

	cmd.AddCommand(listCmd.NewCmdList(f))

	return cmd
}
