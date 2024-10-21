package list

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/hashicorp/go-tfe"
	"github.com/spf13/cobra"

	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
	"github.com/zkhvan/tfc/pkg/text"
)

const MAX_PAGE_SIZE = 100

type Column string

var (
	ColumnID        Column = "ID"
	ColumnName      Column = "NAME"
	ColumnOrg       Column = "ORG"
	ColumnUpdatedAt Column = "UPDATED_AT"
)

type Options struct {
	IO        *iolib.IOStreams
	TFEClient func() (*tfe.Client, error)
	Clock     *cmdutil.Clock
	Printer   cmdutil.Printer

	Organization string
	Name         string
	Limit        int
	Columns      []string
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		IO:        f.IOStreams,
		TFEClient: f.TFEClient,
		Clock:     f.Clock,
	}

	cmd := &cobra.Command{
		Use:   "list organization",
		Short: "List Terraform workspaces",
		Long: heredoc.Doc(`
			List Terraform workspaces.

			Workspaces always belong to a single organization.
		`),
		Example: heredoc.Doc(`
			tfc workspaces list <organization>
		`),
		Aliases:           []string{"ls"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Complete(cmd, args)
			return opts.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Search by the workspace name.")
	cmd.RegisterFlagCompletionFunc("name", cobra.NoFileCompletions)

	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Limit the number of results.")
	cmd.RegisterFlagCompletionFunc("limit", cobra.NoFileCompletions)

	defaultColumns := []string{
		string(ColumnName),
		string(ColumnOrg),
		string(ColumnUpdatedAt),
	}
	cmd.Flags().StringSliceVarP(&opts.Columns, "columns", "c", defaultColumns, "Columns to show.")
	cmd.RegisterFlagCompletionFunc("columns", cmdutil.GenerateOptionCompletionFunc([]string{
		string(ColumnID),
		string(ColumnName),
		string(ColumnOrg),
		string(ColumnUpdatedAt),
	}))

	return cmd
}

func (opts *Options) Complete(cmd *cobra.Command, args []string) {
	opts.Organization = args[0]
}

func (opts *Options) Run(ctx context.Context) error {
	client, err := opts.TFEClient()
	if err != nil {
		return err
	}

	list, err := listWorkspaces(ctx, client, opts.Organization, &ListOptions{
		Name:  opts.Name,
		Limit: opts.Limit,
	})
	if err != nil {
		return err
	}

	if len(list.Items) < list.TotalCount {
		fmt.Fprintf(opts.IO.Out, "Showing %d of %d results\n", opts.Limit, list.TotalCount)
	}

	p := cmdutil.FieldPrinter(opts.IO, opts.Columns...)
	for _, item := range list.Items {
		v := map[string]string{
			"ID":         item.ID,
			"NAME":       item.Name,
			"UPDATED_AT": text.RelativeTimeAgo(opts.Clock.Now(), item.UpdatedAt),
		}

		if item.Organization != nil {
			v["ORG"] = item.Organization.Name
		}

		p.Write(v)
	}
	p.Flush()

	return nil
}

type ListOptions struct {
	Name  string
	Limit int
}

func listWorkspaces(ctx context.Context, client *tfe.Client, org string, opts *ListOptions) (*tfe.WorkspaceList, error) {
	listOpts := &tfe.WorkspaceListOptions{
		Search: opts.Name,
	}

	if opts.Limit < MAX_PAGE_SIZE {
		listOpts.PageSize = opts.Limit
		return client.Workspaces.List(ctx, org, listOpts)
	} else {
		listOpts.PageSize = MAX_PAGE_SIZE
	}

	list := &tfe.WorkspaceList{
		Pagination: &tfe.Pagination{},
		Items:      make([]*tfe.Workspace, 0, opts.Limit),
	}
	for count := 0; count < opts.Limit; {
		response, err := client.Workspaces.List(ctx, org, listOpts)
		if err != nil {
			return nil, err
		}
		list.Pagination = response.Pagination

		// Take the necessary amount of items to reach the limit.
		n := min(
			opts.Limit-count,
			len(response.Items),
		)
		count += n
		list.Items = append(list.Items, response.Items[:n]...)

		// Check if there's a next page.
		if list.NextPage == 0 {
			break
		}
		listOpts.PageNumber = list.NextPage
	}

	return list, nil
}
