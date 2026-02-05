package view

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

type Options struct {
	IO              *iolib.IOStreams
	TFEClient       func() (*tfc.Client, error)
	Clock           *cmdutil.Clock
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	Web         bool
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View workspace details",
		Long: text.Heredoc(`
			View detailed information about a Terraform Cloud workspace.

			By default, displays workspace metadata including name, organization,
			Terraform version, execution mode, VCS connection, and status information.

			Use the --web flag to open the workspace in your default browser.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		Example: text.Heredoc(`
			# View currently detected workspace
			$ tfc workspace view

			# View workspace using -W flag
			$ tfc workspace view -W myorg/myworkspace

			# Open workspace in browser
			$ tfc workspace view -W myorg/myworkspace --web

			# Short form with state.tf
			$ tfc workspace view -w
		`),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Complete(cmd)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	cmd.Flags().BoolVarP(&opts.Web, "web", "w", false, "Open workspace in web browser")

	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command) {
	cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)
}

func (opts *Options) Run(ctx context.Context) error {
	if err := opts.WorkspaceID.Validate(); err != nil {
		return fmt.Errorf("workspace required: use -W ORG/WORKSPACE or ensure state.tf exists")
	}

	client, err := opts.TFEClient()
	if err != nil {
		return fmt.Errorf("failed to initialize TFE client: %w", err)
	}

	// Use ReadWithOptions to include related resources in a single API call
	readOpts := &tfe.WorkspaceReadOptions{
		Include: []tfe.WSIncludeOpt{
			tfe.WSCurrentRun,
			tfe.WSProject,
			tfe.WSLockedBy,
		},
	}
	ws, err := client.Workspaces.ReadWithOptions(ctx, opts.WorkspaceID.Org, opts.WorkspaceID.Workspace, readOpts)
	if err != nil {
		return fmt.Errorf("failed to read workspace %s: %w", opts.WorkspaceID.String(), err)
	}

	if opts.Web {
		return opts.openWorkspaceInBrowser(ctx, ws)
	}

	return opts.displayWorkspace(ws)
}

func (opts *Options) displayWorkspace(ws *tfc.Workspace) error {
	url := buildWorkspaceURL(ws.Organization.Name, ws.Name)
	out := opts.IO.Out

	faintStyle := lipgloss.NewStyle().Faint(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))

	// Identity Section
	fmt.Fprintf(out, "%s\n", headerStyle.Render("IDENTITY"))
	fmt.Fprintf(out, "  Name:                 %s\n", ws.Name)
	fmt.Fprintf(out, "  ID:                   %s\n", faintStyle.Render(ws.ID))
	fmt.Fprintf(out, "  Organization:         %s\n", ws.Organization.Name)

	if ws.Project != nil {
		fmt.Fprintf(out, "  Project:              %s\n", ws.Project.Name)
	}

	if ws.Description != "" {
		fmt.Fprintf(out, "  Description:          %s\n", ws.Description)
	}

	if len(ws.TagNames) > 0 {
		fmt.Fprintf(out, "  Tags:                 %s\n", formatList(ws.TagNames))
	}

	// Configuration Section
	fmt.Fprintf(out, "\n%s\n", headerStyle.Render("CONFIGURATION"))
	fmt.Fprintf(out, "  Terraform Version:    %s\n", ws.TerraformVersion)
	fmt.Fprintf(out, "  Execution Mode:       %s\n", ws.ExecutionMode)

	if ws.AgentPool != nil && ws.ExecutionMode == "agent" {
		fmt.Fprintf(out, "  Agent Pool:           %s\n", ws.AgentPool.Name)
	}

	if ws.WorkingDirectory != "" {
		fmt.Fprintf(out, "  Working Directory:    %s\n", ws.WorkingDirectory)
	}

	// Automation Section
	fmt.Fprintf(out, "\n%s\n", headerStyle.Render("AUTOMATION"))
	fmt.Fprintf(out, "  Auto Apply:           %s\n", formatBool(ws.AutoApply))
	fmt.Fprintf(out, "  Auto Apply Triggers:  %s\n", formatBool(ws.AutoApplyRunTrigger))
	fmt.Fprintf(out, "  Queue All Runs:       %s\n", formatBool(ws.QueueAllRuns))
	fmt.Fprintf(out, "  Speculative Plans:    %s\n", formatBool(ws.SpeculativeEnabled))

	// VCS Connection Section (only if VCS is configured)
	if ws.VCSRepo != nil {
		fmt.Fprintf(out, "\n%s\n", headerStyle.Render("VCS CONNECTION"))
		fmt.Fprintf(out, "  Repository:           %s\n", ws.VCSRepo.DisplayIdentifier)

		if ws.VCSRepo.Branch != "" {
			fmt.Fprintf(out, "  Branch:               %s\n", ws.VCSRepo.Branch)
		} else {
			fmt.Fprintf(out, "  Branch:               (default)\n")
		}

		if ws.VCSRepo.ServiceProvider != "" {
			fmt.Fprintf(out, "  Provider:             %s\n", ws.VCSRepo.ServiceProvider)
		}

		fmt.Fprintf(out, "  File Triggers:        %s\n", formatBool(ws.FileTriggersEnabled))

		if len(ws.TriggerPrefixes) > 0 {
			fmt.Fprintf(out, "  Trigger Prefixes:     %s\n", formatList(ws.TriggerPrefixes))
		}

		if len(ws.TriggerPatterns) > 0 {
			fmt.Fprintf(out, "  Trigger Patterns:     %s\n", formatList(ws.TriggerPatterns))
		}
	}

	// State & Resources Section
	fmt.Fprintf(out, "\n%s\n", headerStyle.Render("STATE & RESOURCES"))
	fmt.Fprintf(out, "  Resource Count:       %d\n", ws.ResourceCount)

	if ws.Locked {
		lockedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
		lockedBy := "unknown"
		if ws.LockedBy != nil {
			if ws.LockedBy.User != nil {
				lockedBy = fmt.Sprintf("user %s", ws.LockedBy.User.Username)
			} else if ws.LockedBy.Run != nil {
				lockedBy = fmt.Sprintf("run %s", ws.LockedBy.Run.ID)
			} else if ws.LockedBy.Team != nil {
				lockedBy = fmt.Sprintf("team %s", ws.LockedBy.Team.Name)
			}
		}
		fmt.Fprintf(out, "  Locked:               %s (by %s)\n", lockedStyle.Render("Yes"), lockedBy)
	} else {
		fmt.Fprintf(out, "  Locked:               No\n")
	}

	if ws.GlobalRemoteState {
		fmt.Fprintf(out, "  Global Remote State:  Yes\n")
	}

	// Current Run Section (only if there's a current run)
	if ws.CurrentRun != nil {
		fmt.Fprintf(out, "\n%s\n", headerStyle.Render("CURRENT RUN"))
		statusStyle := lipgloss.NewStyle().Foreground(tfc.RunStatusColor(ws.CurrentRun.Status))
		fmt.Fprintf(out, "  Status:               %s\n", statusStyle.Render(string(ws.CurrentRun.Status)))
		fmt.Fprintf(out, "  Run ID:               %s\n", faintStyle.Render(ws.CurrentRun.ID))

		if ws.CurrentRun.Message != "" {
			message := ws.CurrentRun.Message
			if len(message) > 60 {
				message = message[:57] + "..."
			}
			fmt.Fprintf(out, "  Message:              %s\n", message)
		}
	}

	// Timestamps Section
	fmt.Fprintf(out, "\n%s\n", headerStyle.Render("TIMESTAMPS"))
	createdTime := text.RelativeTimeAgo(opts.Clock.Now(), ws.CreatedAt)
	updatedTime := text.RelativeTimeAgo(opts.Clock.Now(), ws.UpdatedAt)
	fmt.Fprintf(out, "  Created:              %s\n", createdTime)
	fmt.Fprintf(out, "  Last Updated:         %s\n", updatedTime)

	// URL
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "  URL:                  %s\n", url)

	return nil
}

// formatBool formats a boolean value as a colored yes/no string
func formatBool(v bool) string {
	if v {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("Yes")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("No")
}

// formatList formats a slice of strings as a comma-separated list
func formatList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}

func buildWorkspaceURL(org, workspace string) string {
	hostname := os.Getenv("TFE_HOSTNAME")
	if hostname == "" {
		hostname = "app.terraform.io"
	}
	return fmt.Sprintf("https://%s/app/%s/workspaces/%s", hostname, org, workspace)
}

func (opts *Options) openWorkspaceInBrowser(ctx context.Context, ws *tfc.Workspace) error {
	url := buildWorkspaceURL(ws.Organization.Name, ws.Name)

	// Print URL first
	fmt.Fprintf(opts.IO.Out, "Opening workspace in browser:\n")
	fmt.Fprintf(opts.IO.Out, "%s\n", url)

	// Determine browser command based on OS
	var browserCmd string
	switch runtime.GOOS {
	case "darwin":
		browserCmd = "open"
	case "windows":
		browserCmd = "start"
	default:
		browserCmd = "xdg-open"
	}

	// Execute browser command
	cmd := exec.CommandContext(ctx, browserCmd, url)
	if err := cmd.Start(); err != nil {
		// Graceful degradation - print error but don't fail
		fmt.Fprintf(opts.IO.ErrOut, "Warning: Could not open browser: %v\n", err)
		fmt.Fprintf(opts.IO.Out, "\nPlease open the URL manually: %s\n", url)
		return nil
	}

	return nil
}
