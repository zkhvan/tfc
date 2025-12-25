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
// pairs for shell completion. It queries the TFE API with a 2 second timeout.
func CompletionOrgWorkspace(tfeClient func() (*tfc.Client, error)) func(
	*cobra.Command, []string, string,
) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := tfeClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// List organizations
		orgs, _, err := client.Organizations.List(ctx, &tfc.OrganizationListOptions{})
		if err != nil {
			// On timeout or error, return empty list gracefully
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var completions []string
		for _, org := range orgs {
			// List workspaces for each organization
			workspaces, _, err := client.Workspaces.List(ctx, org.Name, &tfc.WorkspaceListOptions{
				ListOptions: tfc.ListOptions{Limit: 100},
			})
			if err != nil {
				// Skip this org on error
				continue
			}

			for _, ws := range workspaces {
				completion := fmt.Sprintf("%s/%s", org.Name, ws.Name)
				// Filter by prefix
				if strings.HasPrefix(completion, toComplete) {
					completions = append(completions, completion)
				}
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// CompletionVariableNames returns a completion function that provides variable names
// for a given workspace. It requires the org/workspace to be provided as the first argument.
func CompletionVariableNames(tfeClient func() (*tfc.Client, error)) func(
	*cobra.Command, []string, string,
) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the second argument
		if len(args) != 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Parse org/workspace from first argument
		parts := strings.Split(args[0], "/")
		if len(parts) != 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		org, workspace := parts[0], parts[1]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client, err := tfeClient()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		// Read workspace to get workspace ID
		ws, err := client.Workspaces.Read(ctx, org, workspace)
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
