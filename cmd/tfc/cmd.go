package main

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	workspacesCmd "github.com/zkhvan/tfc/cmd/tfc/workspaces"
	"github.com/zkhvan/tfc/pkg/cmdutil"
)

func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tfc",
		Short: "Terraform Cloud/Enterprise CLI",
		Long: heredoc.Doc(`
			Terraform Cloud/Enterprise CLI.

			A CLI for interacting with either the HCP Terraform platform or
			a Terraform Enterprise instance.
		`),
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	cmd.PersistentFlags().Bool("help", false, "Show help for command")

	cmd.AddCommand(workspacesCmd.NewCmdWorkspaces(f))

	return cmd
}
