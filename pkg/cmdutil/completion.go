package cmdutil

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
