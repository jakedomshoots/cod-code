package ceo

import (
	"fmt"
	"strings"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type ExecutionPlan struct {
	Authority  string          `json:"authority"`
	Mode       string          `json:"mode"`
	Steps      []ExecutionStep `json:"steps"`
	NextAction string          `json:"next_action"`
}

type ExecutionStep struct {
	Index   int    `json:"index"`
	Owner   string `json:"owner"`
	Role    string `json:"role"`
	Stage   int    `json:"stage,omitempty"`
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
}

type executionPlanInput struct {
	Packet  jobpacket.Packet
	Results []subagent.Result
	Checks  []checkrunner.Result
	Verdict string
}

func buildExecutionPlan(input executionPlanInput) ExecutionPlan {
	steps := make([]ExecutionStep, 0, len(input.Results)+2)
	for _, result := range input.Results {
		steps = append(steps, ExecutionStep{
			Index:   len(steps) + 1,
			Owner:   result.AgentName,
			Role:    result.Role,
			Stage:   result.Stage,
			Status:  executionStatus(result.Status),
			Summary: strings.TrimSpace(result.Summary),
		})
	}
	if len(input.Checks) > 0 {
		steps = append(steps, checkExecutionStep(len(steps)+1, input.Checks))
	}
	steps = append(steps, ExecutionStep{
		Index:   len(steps) + 1,
		Owner:   "ceo",
		Role:    "final verdict",
		Status:  executionStatus(input.Verdict),
		Summary: fmt.Sprintf("CEO final verdict for %q", input.Packet.Task),
	})
	return ExecutionPlan{
		Authority:  "ceo",
		Mode:       "delegated",
		Steps:      steps,
		NextAction: executionNextAction(input),
	}
}

func checkExecutionStep(index int, checks []checkrunner.Result) ExecutionStep {
	last := checks[len(checks)-1]
	return ExecutionStep{
		Index:   index,
		Owner:   "checker",
		Role:    "run verification checks",
		Status:  executionStatus(last.Status),
		Summary: fmt.Sprintf("%d check attempt(s)", len(checks)),
	}
}

func executionNextAction(input executionPlanInput) string {
	if input.Verdict == "pass" {
		return "accept"
	}
	for _, result := range input.Results {
		if result.Status == "needs_input" {
			return "answer subagent questions"
		}
		if result.Status != "pass" {
			return "retry failed subagents"
		}
	}
	if len(input.Checks) > 0 && input.Checks[len(input.Checks)-1].Status != "pass" {
		return "fix failing checks"
	}
	return "revise run"
}

func executionStatus(status string) string {
	cleanStatus := strings.TrimSpace(status)
	if cleanStatus == "" {
		return "unknown"
	}
	return cleanStatus
}

func renderExecutionPlan(plan ExecutionPlan) string {
	var builder strings.Builder
	builder.WriteString("# CEO Execution Plan\n\n")
	builder.WriteString(fmt.Sprintf("Authority: %s\n", plan.Authority))
	builder.WriteString(fmt.Sprintf("Mode: %s\n", plan.Mode))
	builder.WriteString(fmt.Sprintf("Next action: %s\n\n", plan.NextAction))
	for _, step := range plan.Steps {
		builder.WriteString(fmt.Sprintf("%d. %s - %s [%s]\n", step.Index, step.Owner, step.Role, step.Status))
		if step.Summary != "" {
			builder.WriteString(fmt.Sprintf("   %s\n", step.Summary))
		}
	}
	return builder.String()
}
