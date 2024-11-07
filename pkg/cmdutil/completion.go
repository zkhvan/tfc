package cmdutil

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func MarkFlagsWithNoFileCompletions(cmd *cobra.Command, flags ...string) error {
	var errs []error
	for _, flag := range flags {
		errs = append(errs, cmd.RegisterFlagCompletionFunc(flag, cobra.NoFileCompletions))
	}
	return errors.Join(errs...)
}

func FlagStringEnumSliceP(cmd *cobra.Command, p *[]string, name, shorthand string, defaultValue []string, usage string, options []string) {
	cmd.Flags().StringSliceVarP(p, name, shorthand, defaultValue, usage)
	cmd.RegisterFlagCompletionFunc(name, GenerateOptionCompletionFunc(options))
}

func FlagStringEnumSlice(cmd *cobra.Command, p *[]string, name string, defaultValue []string, usage string, options []string) {
	cmd.Flags().StringSliceVar(p, name, defaultValue, usage)
	cmd.RegisterFlagCompletionFunc(name, GenerateOptionCompletionFunc(options))
}

func GenerateOptionCompletionFunc(opts []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
