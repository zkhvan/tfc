package init

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)

	Organization string
	Workspace    string
	Project      string
	Force        bool
}

func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a state.tf file with cloud backend configuration",
		Long: text.Heredoc(`
			Generate a state.tf file with Terraform Cloud backend configuration.

			This command creates a state.tf file in the current directory containing
			the terraform.cloud block with organization and workspace settings.
		`),
		Example: text.IndentHeredoc(2, `
			# Generate state.tf with org and workspace
			tfc init --org my-org --workspace my-workspace

			# Include a project
			tfc init --org my-org --workspace my-workspace --project my-project

			# Overwrite an existing state.tf
			tfc init --org my-org --workspace my-workspace --force
		`),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}
			return opts.Run()
		},
	}

	cmd.Flags().StringVarP(&opts.Organization, "org", "o", "", "Organization name (required)")
	cmd.Flags().StringVarP(&opts.Workspace, "workspace", "w", "", "Workspace name (required)")
	cmd.Flags().StringVarP(&opts.Project, "project", "p", "", "Project name (optional)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing state.tf file")

	_ = cmd.MarkFlagRequired("org")
	_ = cmd.MarkFlagRequired("workspace")

	_ = cmd.RegisterFlagCompletionFunc("org", cmdutil.CompletionOrganizations(opts.TFEClient))
	_ = cmd.RegisterFlagCompletionFunc("workspace", cmdutil.CompletionWorkspacesFromOrgFlag(opts.TFEClient))
	_ = cmdutil.MarkFlagsWithNoFileCompletions(cmd, "project", "force")

	return cmd
}

func (opts *Options) Validate() error {
	if opts.Organization == "" {
		return fmt.Errorf("--org is required")
	}
	if opts.Workspace == "" {
		return fmt.Errorf("--workspace is required")
	}
	return nil
}

func (opts *Options) Run() error {
	const filename = "state.tf"

	if !opts.Force {
		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("%s already exists; use --force to overwrite", filename)
		}
	}

	content := generateHCL(opts.Organization, opts.Workspace, opts.Project)

	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing %s: %w", filename, err)
	}

	fmt.Fprintf(opts.IO.ErrOut, "Wrote %s\n", filename)
	return nil
}

func generateHCL(org, workspace, project string) string {
	var b strings.Builder

	b.WriteString("terraform {\n")
	b.WriteString("  cloud {\n")
	b.WriteString(fmt.Sprintf("    organization = %q\n", org))
	b.WriteString("\n")
	b.WriteString("    workspaces {\n")
	b.WriteString(fmt.Sprintf("      name    = %q\n", workspace))
	if project != "" {
		b.WriteString(fmt.Sprintf("      project = %q\n", project))
	}
	b.WriteString("    }\n")
	b.WriteString("  }\n")
	b.WriteString("}\n")

	return b.String()
}
