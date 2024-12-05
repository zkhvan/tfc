package list

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/internal/tfc"
	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

const (
	ColumnName      = "NAME"
	ColumnEmail     = "EMAIL"
	ColumnCreatedAt = "CREATED_AT"
)

var (
	ColumnDefault = []string{
		ColumnName,
		ColumnCreatedAt,
	}

	ColumnAll = []string{
		ColumnName,
		ColumnEmail,
		ColumnCreatedAt,
	}
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfc.Client, error)
	Clock     *cmdutil.Clock

	Columns []string
	Limit   int
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Terraform organizations",
		Long: text.Heredoc(`
			List Terraform organizations.
		`),
		Aliases:           []string{"ls"},
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 20, "Limit the number of results.")
	_ = cmdutil.FlagStringEnumSliceP(cmd, &opts.Columns, "columns", "c", ColumnDefault, "Columns to show.", ColumnAll)

	_ = cmdutil.MarkAllFlagsWithNoFileCompletions(cmd)

	return cmd
}

func (opts *Options) Complete(_ *cobra.Command, _ []string) {
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	orgs, err := listOrganizations(ctx, client, &listOptions{Limit: opts.Limit})
	if err != nil {
		return err
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)
	for _, org := range orgs {
		opts.write(p, org)
	}
	p.Flush()

	return nil
}

type listOptions struct {
	Limit int
}

func listOrganizations(ctx context.Context, client *tfc.Client, opts *listOptions) ([]*tfe.Organization, error) {
	f := func(o tfe.ListOptions) ([]*tfe.Organization, *tfe.Pagination, error) {
		response, err := client.Organizations.List(ctx, &tfe.OrganizationListOptions{
			ListOptions: o,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list organizations: %w", err)
		}

		return response.Items, response.Pagination, nil
	}

	pager := tfepaging.New(f)

	var orgs []*tfe.Organization
	for i, org := range pager.All() {
		if opts.Limit <= i {
			break
		}

		orgs = append(orgs, org)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

func (opts *Options) write(p cmdutil.Printer, org *tfe.Organization) {
	v := make(map[string]string, 0)

	v[ColumnName] = org.Name
	v[ColumnEmail] = org.Email
	v[ColumnCreatedAt] = text.RelativeTimeAgo(opts.Clock.Now(), org.CreatedAt)

	p.Write(v)
}
