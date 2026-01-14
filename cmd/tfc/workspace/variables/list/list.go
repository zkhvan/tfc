package list

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
	"github.com/zkhvan/tfc/pkg/tfconfig"
)

const (
	ColumnID          string = "ID"
	ColumnKey         string = "KEY"
	ColumnValue       string = "VALUE"
	ColumnDescription string = "DESCRIPTION"
	ColumnSensitive   string = "SENSITIVE"
	ColumnCategory    string = "CATEGORY"
	ColumnHCL         string = "HCL"
)

var (
	ColumnsDefault = []string{
		ColumnKey,
		ColumnValue,
		ColumnDescription,
	}
	ColumnsAll = []string{
		ColumnID,
		ColumnKey,
		ColumnValue,
		ColumnDescription,
		ColumnSensitive,
		ColumnCategory,
		ColumnHCL,
	}
)

type Options struct {
	IO              *iolib.IOStreams
	TFEClient       func() (*tfc.Client, error)
	Clock           *cmdutil.Clock
	TerraformConfig func() *tfconfig.TerraformConfig

	WorkspaceID cmdutil.WorkspaceIdentifier
	Columns     []string
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:              f.IOStreams,
		TFEClient:       f.TFEClient,
		Clock:           f.Clock,
		TerraformConfig: f.TerraformConfig,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List a workspace's variables",
		Aliases: []string{"ls"},
		Long: text.Heredoc(`
			List a workspace's variables.

			If -W/--workspace is not specified and state.tf is present,
			the organization and workspace will be read from state.tf.
		`),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmdutil.AddWorkspaceFlag(cmd, &opts.WorkspaceID, opts.TFEClient)

	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", ColumnsDefault, "Columns to show.", ColumnsAll)
	_ = cmdutil.MarkFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, _ []string) {
	cmdutil.CompleteWorkspaceIdentifierSilent(cmd, &opts.WorkspaceID, opts.TerraformConfig)
}

func (opts *Options) Run(ctx context.Context) error {
	if err := opts.WorkspaceID.Validate(); err != nil {
		return fmt.Errorf("workspace required: use -W ORG/WORKSPACE or ensure state.tf exists")
	}

	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	ws, err := client.Workspaces.Read(ctx, opts.WorkspaceID.Org, opts.WorkspaceID.Workspace)
	if err != nil {
		return err
	}

	vars, _, err := client.Variables.List(ctx, ws.ID, &tfc.VariableListOptions{})
	if err != nil {
		return err
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)
	for _, v := range vars {
		p.Write(opts.extractFields(v))
	}
	p.Flush()

	return nil
}

func (opts *Options) extractFields(v *tfe.Variable) map[string]string {
	out := map[string]string{
		ColumnID:          v.ID,
		ColumnKey:         v.Key,
		ColumnValue:       v.Value,
		ColumnDescription: v.Description,
		ColumnCategory:    string(v.Category),
		ColumnHCL:         strconv.FormatBool(v.HCL),
		ColumnSensitive:   strconv.FormatBool(v.Sensitive),
	}

	return out
}
