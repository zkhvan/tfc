package tfc

import (
	"context"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
	"github.com/zkhvan/tfc/pkg/term/color"
)

var runStatusGroups = map[RunStatus]RunStatusGroup{
	tfe.RunCostEstimated:      "pending",
	tfe.RunPlanned:            "pending",
	tfe.RunPolicyChecked:      "pending",
	tfe.RunPolicyOverride:     "pending",
	tfe.RunPostPlanCompleted:  "pending",
	tfe.RunPreApplyCompleted:  "pending",
	tfe.RunErrored:            "errored",
	tfe.RunApplying:           "running",
	tfe.RunConfirmed:          "running",
	tfe.RunCostEstimating:     "running",
	tfe.RunFetching:           "running",
	tfe.RunPlanning:           "running",
	tfe.RunPolicyChecking:     "running",
	tfe.RunPostPlanRunning:    "running",
	tfe.RunPreApplyRunning:    "running",
	tfe.RunPrePlanCompleted:   "running",
	tfe.RunPrePlanRunning:     "running",
	tfe.RunApplyQueued:        "holding",
	tfe.RunPending:            "holding",
	tfe.RunPlanQueued:         "holding",
	tfe.RunQueuing:            "holding",
	tfe.RunApplied:            "applied",
	tfe.RunPlannedAndFinished: "applied",
}

func RunStatusColor(status RunStatus) lipgloss.Color {
	group, ok := runStatusGroups[status]
	if !ok {
		return color.Black
	}

	switch group {
	case "pending":
		return color.Yellow
	case "errored":
		return color.Red
	case "running":
		return color.Blue
	case "holding":
		return color.LightBlack
	case "applied":
		return color.Green
	}

	return color.Black
}

type RunStatusGroup string
type RunStatus = tfe.RunStatus

const (
	RunStatusGroupUnknown RunStatusGroup = ""
	RunStatusGroupPending RunStatusGroup = "pending"
	RunStatusGroupErrored RunStatusGroup = "errored"
	RunStatusGroupRunning RunStatusGroup = "running"
	RunStatusGroupHolding RunStatusGroup = "holding"
	RunStatusGroupApplied RunStatusGroup = "applied"
)

func RunStatusesInGroup(x RunStatusGroup) []RunStatus {
	var statuses []RunStatus
	for s, g := range runStatusGroups {
		if g == x {
			statuses = append(statuses, s)
		}
	}
	return statuses
}

type RunGroup string

const (
	RunGroupNonFinal    RunGroup = "non_final"
	RunGroupFinal       RunGroup = "final"
	RunGroupDiscardable RunGroup = "discardable"
)

type RunCreateOptions = tfe.RunCreateOptions

type WorkspaceRunListOptions struct {
	ListOptions tfe.ListOptions
	Limit       int
}

// RunsService provides methods for working with Terraform runs.
type RunsService service

// Create creates a new run with the given options.
func (s *RunsService) Create(ctx context.Context, options RunCreateOptions) (*Run, error) {
	return s.tfe.Runs.Create(ctx, options)
}

// List lists all runs for a given workspace.
func (s *RunsService) List(
	ctx context.Context, workspaceID string, options *WorkspaceRunListOptions,
) ([]*Run, *Pagination, error) {
	o := *options

	if o.Limit == 0 {
		o.Limit = 20
	}

	f := func(lo tfe.ListOptions) ([]*Run, *tfe.Pagination, error) {
		o.ListOptions = lo

		result, err := s.tfe.Runs.List(ctx, workspaceID, &tfe.RunListOptions{
			ListOptions: lo,
		})
		if err != nil {
			return nil, nil, err
		}

		return result.Items, result.Pagination, nil
	}

	current := Pagination{}
	pager := tfepaging.New(f)

	var runs []*Run
	for i, run := range pager.All() {
		current.Pagination = *pager.Current()

		if o.Limit <= len(runs) {
			if i < current.TotalCount {
				current.ReachedLimit = true
			}
			break
		}

		runs = append(runs, run)
	}

	if err := pager.Err(); err != nil {
		return nil, nil, err
	}

	return runs, &current, nil
}
