package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/zkhvan/tfc/internal/tfc"
)

func MarkAllFlagsWithNoFileCompletions(cmd *cobra.Command) error {
	var errs []error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if _, ok := cmd.GetFlagCompletionFunc(f.Name); !ok {
			errs = append(errs, cmd.RegisterFlagCompletionFunc(f.Name, cobra.NoFileCompletions))
		}
	})
	return errors.Join(errs...)
}

func MarkFlagsWithNoFileCompletions(cmd *cobra.Command, flags ...string) error {
	var errs []error
	for _, flag := range flags {
		errs = append(errs, cmd.RegisterFlagCompletionFunc(flag, cobra.NoFileCompletions))
	}
	return errors.Join(errs...)
}

func FlagStringEnumSliceP(
	cmd *cobra.Command,
	p *[]string,
	name,
	shorthand string,
	defaultValue []string,
	usage string,
	options []string,
) error {
	cmd.Flags().StringSliceVarP(p, name, shorthand, defaultValue, usage)
	return cmd.RegisterFlagCompletionFunc(name, GenerateOptionCompletionFunc(options))
}

func FlagStringEnumSlice(
	cmd *cobra.Command,
	p *[]string,
	name string,
	defaultValue []string,
	usage string,
	options []string,
) error {
	cmd.Flags().StringSliceVar(p, name, defaultValue, usage)
	return cmd.RegisterFlagCompletionFunc(name, GenerateOptionCompletionFunc(options))
}

func GenerateOptionCompletionFunc(opts []string) func(
	*cobra.Command, []string, string,
) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		filteredOpts := make([]string, len(opts))
		copy(filteredOpts, opts)

		// Filter the options have already been used.
		tokens := strings.Split(toComplete, ",")
		for _, token := range tokens[:len(tokens)-1] {
			filteredOpts = slices.DeleteFunc(filteredOpts, func(s string) bool {
				return s == token
			})
		}

		currentToken := tokens[len(tokens)-1]
		filteredOpts = slices.DeleteFunc(filteredOpts, func(s string) bool {
			return !strings.HasPrefix(s, currentToken)
		})

		// Pre-pend the existing options that have been selected.
		if len(tokens) > 1 {
			for i := range filteredOpts {
				filteredOpts[i] = fmt.Sprintf(
					"%s,%s",
					strings.Join(tokens[:len(tokens)-1], ","),
					filteredOpts[i],
				)
			}
		}

		return filteredOpts, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	}
}

// CompletionOrgWorkspace returns a completion function that provides organization/workspace
// pairs for shell completion with server-side filtering.
func CompletionOrgWorkspace(tfeClient func() (*tfc.Client, error)) func(
	*cobra.Command, []string, string,
) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		client, err := tfeClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		parsed := tfc.ParseOrgWorkspace(toComplete)

		// Complete organizations with trailing slash (no space after)
		if !parsed.HasOrg() {
			orgOpts := &tfc.OrganizationListOptions{}
			if toComplete != "" {
				orgOpts.Query = toComplete
			}
			orgs, _, err := client.Organizations.List(ctx, orgOpts)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var completions []string
			for _, org := range orgs {
				completions = append(completions, fmt.Sprintf("%s/", org.Name))
			}
			return completions, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		}

		// Complete workspaces for the specified organization
		workspaces, _, err := client.Workspaces.List(ctx, parsed.Org, &tfc.WorkspaceListOptions{
			Search:      parsed.Workspace,
			ListOptions: tfc.ListOptions{Limit: 100},
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var completions []string
		for _, ws := range workspaces {
			completions = append(completions, fmt.Sprintf("%s/%s", parsed.Org, ws.Name))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompletionVariableNames returns a completion function that provides variable names
// for a given workspace. It requires the org/workspace to be provided as the first argument.
func CompletionVariableNames(tfeClient func() (*tfc.Client, error)) func(
	*cobra.Command, []string, string,
) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		parsed := tfc.ParseOrgWorkspace(args[0])
		if !parsed.IsComplete() {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		client, err := tfeClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Read workspace to get workspace ID
		ws, err := client.Workspaces.Read(ctx, parsed.Org, parsed.Workspace)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// List variables for the workspace
		vars, _, err := client.Variables.List(ctx, ws.ID, &tfc.VariableListOptions{
			ListOptions: tfc.ListOptions{Limit: 1000},
		})
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var completions []string
		for _, v := range vars {
			// Filter by prefix
			if strings.HasPrefix(v.Key, toComplete) {
				completions = append(completions, v.Key)
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
