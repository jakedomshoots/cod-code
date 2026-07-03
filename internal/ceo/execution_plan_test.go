package ceo

import (
	"context"
	"os"
	"testing"
)

func Test_Runtime_RunJob_builds_execution_plan_with_ceo_final_verdict(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.JobOwner != "coder" {
		t.Fatalf("JobOwner = %q, want coder", report.JobOwner)
	}
	plan := report.ExecutionPlan
	if plan.Authority != "ceo" || plan.Mode != "delegated" {
		t.Fatalf("execution plan = %#v, want delegated CEO authority", plan)
	}
	if len(plan.Steps) != 4 {
		t.Fatalf("execution plan steps = %d, want 4: %#v", len(plan.Steps), plan.Steps)
	}
	if plan.Steps[0].Owner != "scanner" || plan.Steps[0].Status != "pass" {
		t.Fatalf("first step = %#v, want passing scanner step", plan.Steps[0])
	}
	last := plan.Steps[len(plan.Steps)-1]
	if last.Owner != "ceo" || last.Status != "pass" || plan.NextAction != "accept" {
		t.Fatalf("final plan state = %#v / %q, want CEO accept", last, plan.NextAction)
	}
}

func Test_Runtime_RunJob_reports_planner_as_owner_for_planning_task(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Plan roadmap",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.JobOwner != "planner" {
		t.Fatalf("JobOwner = %q, want planner", report.JobOwner)
	}
}

func Test_Runtime_RunJob_adds_failed_check_to_execution_plan(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_fail_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_HELPER_PROCESS=fail"},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	plan := report.ExecutionPlan
	if len(plan.Steps) != 5 {
		t.Fatalf("execution plan steps = %d, want 5: %#v", len(plan.Steps), plan.Steps)
	}
	checkStep := plan.Steps[3]
	if checkStep.Owner != "checker" || checkStep.Status != "fail" {
		t.Fatalf("check step = %#v, want failed checker step", checkStep)
	}
	if plan.NextAction != "fix failing checks" {
		t.Fatalf("NextAction = %q, want fix failing checks", plan.NextAction)
	}
}
