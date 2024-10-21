package cmdutil

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

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
