package main

import (
	"github.com/spf13/cobra"

	versionCmd "github.com/zkhvan/tfc/cmd/tfc/version"
	workspacesCmd "github.com/zkhvan/tfc/cmd/tfc/workspaces"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/text"
)

func NewCmdRoot(f *cmdutil.Factory, version, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tfc",
		Short: "Terraform Cloud/Enterprise CLI",
		Annotations: map[string]string{
			"versionInfo": versionCmd.Format(version, date),
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
	cmd.AddCommand(workspacesCmd.NewCmdWorkspaces(f))

	return cmd
}
