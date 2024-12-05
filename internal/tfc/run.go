package tfc

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/pkg/term/color"
)

var runStatusGroups = map[tfe.RunStatus]string{
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

func RunStatusColor(status tfe.RunStatus) lipgloss.Color {
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
