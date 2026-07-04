package ceo

import (
	"testing"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/subagent"
)

func Test_verdict_allows_repaired_run_when_latest_check_passes(t *testing.T) {
	// Given
	results := []subagent.Result{
		{AgentName: "coder", Status: "fail"},
		{AgentName: "coder", Status: "pass"},
	}
	checks := []checkrunner.Result{
		{Status: "fail"},
		{Status: "pass"},
	}

	// When
	got := verdict(results, checks, summarizeVerification(results, checks))

	// Then
	if got != "pass" {
		t.Fatalf("verdict = %q, want pass after repaired check", got)
	}
}

func Test_verdict_keeps_needs_input_even_with_checks(t *testing.T) {
	// Given
	results := []subagent.Result{
		{AgentName: "coder", Status: "needs_input"},
	}
	checks := []checkrunner.Result{{Status: "pass"}}

	// When
	got := verdict(results, checks, summarizeVerification(results, checks))

	// Then
	if got != "needs_input" {
		t.Fatalf("verdict = %q, want needs_input", got)
	}
}
