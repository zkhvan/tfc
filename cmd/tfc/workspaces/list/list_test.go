package list_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/cmd/tfc/workspaces/list"
	"github.com/zkhvan/tfc/internal/test"
	"github.com/zkhvan/tfc/internal/tfe/tfetest"
	"github.com/zkhvan/tfc/pkg/clock"
	"github.com/zkhvan/tfc/pkg/cmdutil"
	"github.com/zkhvan/tfc/pkg/iolib"
)

var (
	referenceTime = time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)
)

func TestList(t *testing.T) {
	client, mux, teardown := tfetest.Setup()
	defer teardown()

	mux.HandleFunc(
		"GET /api/v2/organizations",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":[{"id":"o","type":"organizations","attributes":{"name":"o"}}]}`)
		},
	)

	mux.HandleFunc(
		"GET /api/v2/organizations/{organization}/workspaces",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":[
				{"id":"ws-1","type":"workspaces","attributes":{
					"name":"workspace-1",
					"updated-at":"1999-12-31T12:00:00Z"
				},"relationships":{
					"organization":{"data":{"id":"o","type":"organizations"}}
				}}
			]}`)
		},
	)

	result, err := runCommand(t, client, "o")
	if err != nil {
		t.Fatal(err)
	}

	test.Buffer(t, result.OutBuf, heredoc.Doc(`
		ORG  NAME         UPDATED_AT
		o    workspace-1  about 1 day ago
	`))
}

func runCommand(t *testing.T, client *tfe.Client, args ...string) (*tfetest.CmdOut, error) {
	t.Helper()

	ios, _, stdout, stderr := iolib.Test()

	f := &cmdutil.Factory{
		IOStreams: ios,
		TFEClient: func() (*tfe.Client, error) { return client, nil },
		Clock:     cmdutil.NewClock(clock.FrozenClock(referenceTime)),
	}

	cmd := list.NewCmdList(f)
	cmd.SetArgs(args)

	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true

	_, err := cmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	return &tfetest.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}
}
