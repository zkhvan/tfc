package cmdutil_test

import (
	"testing"

	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/pkg/cmdutil"
)

func TestCompleteColumns(t *testing.T) {
	options := []string{
		"ID",
		"Name",
		"Org",
		"Org1",
		"Org2",
	}

	complete := cmdutil.GenerateOptionCompletionFunc(options)

	t.Run("with no characters should return all options",
		func(t *testing.T) {
			toComplete := ""

			opts, _ := complete(nil, nil, toComplete)

			test.StringSlice(t, opts, options)
		},
	)

	t.Run("with one character",
		func(t *testing.T) {
			t.Run("should return the single filtered option",
				func(t *testing.T) {
					toComplete := "I"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"ID"})
				},
			)

			t.Run("should return multiple filtered options",
				func(t *testing.T) {
					toComplete := "O"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"Org", "Org1", "Org2"})
				},
			)
		},
	)

	t.Run("with multiple characters",
		func(t *testing.T) {
			t.Run("should return the single filtered option",
				func(t *testing.T) {
					toComplete := "ID"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"ID"})
				},
			)

			t.Run("should return the multiple filtered options",
				func(t *testing.T) {
					toComplete := "Org"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{"Org", "Org1", "Org2"})
				},
			)
		},
	)

	t.Run("with one item and a comma",
		func(t *testing.T) {
			t.Run("should return the correctly filtered options and remove the completed option (ID)",
				func(t *testing.T) {
					toComplete := "ID,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name",
						"ID,Org",
						"ID,Org1",
						"ID,Org2",
					})
				},
			)

			t.Run("should return the correctly filtered options and remove the completed option (Org)",
				func(t *testing.T) {
					toComplete := "Org,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"Org,ID",
						"Org,Name",
						"Org,Org1",
						"Org,Org2",
					})
				},
			)
		},
	)

	t.Run("with one item, a comma, and some partial text",
		func(t *testing.T) {
			t.Run("should filter the options using the partial text",
				func(t *testing.T) {
					toComplete := "ID,N"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name",
					})
				},
			)
		},
	)

	t.Run("with multiple items and a comma",
		func(t *testing.T) {
			t.Run("should return the correctly filtered options and remove the completed option (ID)",
				func(t *testing.T) {
					toComplete := "ID,Name,"

					opts, _ := complete(nil, nil, toComplete)

					test.StringSlice(t, opts, []string{
						"ID,Name,Org",
						"ID,Name,Org1",
						"ID,Name,Org2",
					})
				},
			)
		},
	)
}
