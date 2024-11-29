package list

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
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
	IO        *iolib.IOStreams
	TFEClient func() (*tfe.Client, error)
	Clock     *cmdutil.Clock

	Org       string
	Workspace string
	Columns   []string
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:     "list <ORG/WORKSPACE>",
		Short:   "List a workspace's variables",
		Aliases: []string{"ls"},
		Long: text.Heredoc(`
			List a workspace's variables.
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", ColumnsDefault, "Columns to show.", ColumnsAll)

	return cmd
}

func (opts *Options) Complete(_ *cobra.Command, args []string) {
	opts.Org, opts.Workspace = parse(args[0])
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	ws, err := client.Workspaces.Read(ctx, opts.Org, opts.Workspace)
	if err != nil {
		return err
	}

	vars, err := client.Variables.List(ctx, ws.ID, &tfe.VariableListOptions{})
	if err != nil {
		return err
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)
	for _, v := range vars.Items {
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

func parse(workspace string) (string, string) {
	parts := strings.Split(workspace, "/")

	return parts[0], parts[1]
}
