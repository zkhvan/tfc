package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

// WorkspaceIdentifier holds the parsed org/workspace from the -W flag.
type WorkspaceIdentifier struct {
	Raw string // The raw flag value
	tfc.OrgWorkspace
}

// AddWorkspaceFlag adds the -W/--workspace flag to a command that requires
// both organization and workspace. The flag accepts "org/workspace" format.
//
// Example:
//
//	var wsID cmdutil.WorkspaceIdentifier
//	cmdutil.AddWorkspaceFlag(cmd, &wsID, opts.TFEClient)
func AddWorkspaceFlag(
	cmd *cobra.Command,
	target *WorkspaceIdentifier,
	tfeClient func() (*tfc.Client, error),
) {
	cmd.Flags().StringVarP(
		&target.Raw,
		"workspace",
		"W",
		"",
		"Workspace identifier in ORG/WORKSPACE format",
	)

	// Register shell completion
	_ = cmd.RegisterFlagCompletionFunc(
		"workspace",
		CompletionOrgWorkspace(tfeClient),
	)
}

// CompleteWorkspaceIdentifier parses the -W flag and applies state.tf fallback.
// This should be called in the command's Complete() function.
//
// Example:
//
//	func (opts *Options) Complete(cmd *cobra.Command) {
//	    cmdutil.CompleteWorkspaceIdentifier(cmd, &opts.WorkspaceID, opts.TerraformConfig)
//	    // ... rest of completion logic
//	}
func CompleteWorkspaceIdentifier(
	_ *cobra.Command,
	wsID *WorkspaceIdentifier,
	terraformConfig func() *tfconfig.TerraformConfig,
) error {
	if wsID.Raw == "" {
		// Flag not provided, try state.tf fallback
		if cfg := terraformConfig(); cfg != nil && cfg.IsValid() {
			wsID.Org = cfg.Organization
			wsID.Workspace = cfg.Workspace.Name
			return nil
		}
		return fmt.Errorf("workspace must be specified via -W flag or state.tf file")
	}

	// Parse the flag value
	wsID.OrgWorkspace = tfc.ParseOrgWorkspace(wsID.Raw)

	// Validate that both org and workspace are present
	return wsID.Validate()
}

// CompleteWorkspaceIdentifierSilent is like CompleteWorkspaceIdentifier but doesn't
// return an error if the workspace is not specified. This is useful for commands
// that want to check the workspace later or provide a custom error message.
func CompleteWorkspaceIdentifierSilent(
	_ *cobra.Command,
	wsID *WorkspaceIdentifier,
	terraformConfig func() *tfconfig.TerraformConfig,
) {
	if wsID.Raw == "" {
		// Flag not provided, try state.tf fallback
		if cfg := terraformConfig(); cfg != nil && cfg.IsValid() {
			wsID.Org = cfg.Organization
			wsID.Workspace = cfg.Workspace.Name
		}
		return
	}

	// Parse the flag value
	wsID.OrgWorkspace = tfc.ParseOrgWorkspace(wsID.Raw)
}
