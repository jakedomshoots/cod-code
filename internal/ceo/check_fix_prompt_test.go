package ceo

import (
	"strings"
	"testing"

	"ceoharness/internal/checkrunner"
)

func Test_buildCheckFixTask_includes_compact_failed_check_metadata(t *testing.T) {
	// Given
	longOutput := strings.Repeat("stderr line ", 100)
	checks := []checkrunner.Result{
		{
			Argv:        []string{"go", "test", "./..."},
			Status:      "fail",
			ExitCode:    1,
			CheckIndex:  2,
			Attempt:     3,
			MaxAttempts: 3,
			DurationMS:  456,
			Stderr:      longOutput,
		},
	}

	// When
	task := buildCheckFixTask("Repair app", checks, 1)

	// Then
	for _, want := range []string{
		"Check index: 2",
		"Check attempt: 3/3",
		"Duration ms: 456",
		"Status: fail",
		"Command: go test ./...",
		"Exit code: 1",
	} {
		if !strings.Contains(task, want) {
			t.Fatalf("check-fix task missing %q:\n%s", want, task)
		}
	}
	if strings.Contains(task, longOutput) {
		t.Fatalf("check-fix task included uncapped stderr")
	}
	if !strings.Contains(task, "[truncated]") {
		t.Fatalf("check-fix task = %q, want truncation marker", task)
	}
}

func Test_buildCheckFixTask_includes_failed_scorer_metadata_when_available(t *testing.T) {
	// Given
	checks := []checkrunner.Result{{
		Argv:     []string{"go", "test", "./internal/ceo"},
		Status:   "fail",
		ExitCode: 1,
		Stderr:   "unit failed",
	}}
	scorerChecks := []RepairFailureDetail{{
		Name:    "diff_term:retry_history",
		Status:  "fail",
		Message: "missing required diff term",
	}}

	// When
	task := buildCheckFixTask("Repair app", checks, 2, scorerChecks)

	// Then
	for _, want := range []string{
		"Failed scorer checks:",
		"diff_term:retry_history",
		"missing required diff term",
	} {
		if !strings.Contains(task, want) {
			t.Fatalf("check-fix task missing scorer detail %q:\n%s", want, task)
		}
	}
}
