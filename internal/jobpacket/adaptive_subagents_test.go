package jobpacket

import (
	"errors"
	"testing"
)

func Test_Build_uses_planning_subagents_when_task_is_planning(t *testing.T) {
	// Given
	task := "Plan the roadmap for the CLI harness"

	// When
	packet, err := Build(task)

	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	wantNames := []string{"planner", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_Build_uses_research_subagents_when_task_is_research(t *testing.T) {
	// Given
	task := "Research provider docs"

	// When
	packet, err := Build(task)

	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	wantNames := []string{"researcher", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_Build_keeps_mixed_high_risk_subagents_to_three_when_task_needs_research_and_code(t *testing.T) {
	// Given
	task := "Research auth bug and implement fix"

	// When
	packet, err := Build(task)

	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	wantNames := []string{"coder", "security", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_Build_keeps_data_specialist_when_default_budget_allows_one_risk_subagent(t *testing.T) {
	// Given
	task := "Implement database migration fix"

	// When
	packet, err := Build(task)

	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	wantNames := []string{"coder", "database", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_Build_keeps_only_one_risk_specialist_when_mixed_task_has_multiple_risk_areas(t *testing.T) {
	// Given
	task := "Research payment database migration and deploy the fix"

	// When
	packet, err := Build(task)

	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	wantNames := []string{"coder", "billing", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_BuildWithOptions_uses_explicit_subagent_budget_when_task_has_multiple_risk_areas(t *testing.T) {
	// Given
	task := "Research payment database migration and deploy the fix"

	// When
	packet, err := BuildWithOptions(BuildOptions{Task: task, MaxSubagents: 7})

	// Then
	if err != nil {
		t.Fatalf("BuildWithOptions returned error: %v", err)
	}
	wantNames := []string{"planner", "researcher", "coder", "billing", "database", "release", "reviewer"}
	assertSubagentNames(t, packet.Subagents, wantNames)
}

func Test_BuildWithOptions_rejects_subagent_budget_above_hard_cap(t *testing.T) {
	// Given
	task := "Fix a failing test"

	// When
	_, err := BuildWithOptions(BuildOptions{Task: task, MaxSubagents: MaxDelegatedSubagents + 1})

	// Then
	if !errors.Is(err, ErrInvalidSubagent) {
		t.Fatalf("error = %v, want ErrInvalidSubagent", err)
	}
}

func assertSubagentNames(t *testing.T, subagents []Subagent, wantNames []string) {
	t.Helper()
	if len(subagents) != len(wantNames) {
		t.Fatalf("subagent count = %d, want %d: %#v", len(subagents), len(wantNames), subagents)
	}
	for index, wantName := range wantNames {
		if subagents[index].Name != wantName {
			t.Fatalf("subagents[%d].Name = %q, want %q", index, subagents[index].Name, wantName)
		}
	}
}
