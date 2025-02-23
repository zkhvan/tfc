package tfc

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"

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
