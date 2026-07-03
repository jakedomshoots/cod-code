package ceo

import (
	"strings"
	"testing"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
)

func Test_renderCEOReviewPrompt_includes_compact_check_metadata(t *testing.T) {
	// Given
	longOutput := strings.Repeat("check output ", 80)
	input := ceoReviewInput{
		Packet: jobpacket.Packet{
			Task: "Fix failing test",
			TaskProfile: jobpacket.TaskProfile{
				Kind:      "coding",
				RiskLevel: "low",
			},
		},
		GuardVerdict: "fail",
		Checks: []checkrunner.Result{
			{
				Argv:        []string{"go", "test", "./..."},
				Status:      "fail",
				ExitCode:    1,
				CheckIndex:  1,
				Attempt:     2,
				MaxAttempts: 2,
				DurationMS:  123,
				Stdout:      longOutput,
				Stderr:      "compile failed",
			},
		},
	}

	// When
	prompt := renderCEOReviewPrompt(input)

	// Then
	for _, want := range []string{
		"argv=\"go test ./...\"",
		"index=1",
		"attempt=2/2",
		"exit_code=1",
		"duration_ms=123",
		"stderr=compile failed",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("CEO prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, longOutput) {
		t.Fatalf("CEO prompt included uncapped check output")
	}
	if !strings.Contains(prompt, "stdout=check output") || !strings.Contains(prompt, "...") {
		t.Fatalf("CEO prompt = %q, want compact stdout with truncation marker", prompt)
	}
}
