package main

import (
	"github.com/spf13/cobra"

	organizationCmd "github.com/zkhvan/tfc/cmd/tfc/organization"
	runCmd "github.com/zkhvan/tfc/cmd/tfc/run"
	versionCmd "github.com/zkhvan/tfc/cmd/tfc/version"
	workspaceCmd "github.com/zkhvan/tfc/cmd/tfc/workspace"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdRoot(f *cmdutil.Factory, version, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tfc",
		Short: "Terraform Cloud/Enterprise CLI",
		Annotations: map[string]string{
			"versionInfo": versionCmd.Format(f.ExecutableName, version, date),
		},
		Long: text.Heredoc(`
			Terraform Cloud/Enterprise CLI.

			A CLI for interacting with either the HCP Terraform platform or
			a Terraform Enterprise instance.
		`),
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	cmd.PersistentFlags().Bool("help", false, "Show help for command")

	cmd.AddCommand(versionCmd.NewCmdVersion(f, version, date))
	cmd.AddCommand(workspaceCmd.NewCmdWorkspace(f))
	cmd.AddCommand(organizationCmd.NewCmdOrganization(f))
	cmd.AddCommand(runCmd.NewCmdRun(f))

	return cmd
}
