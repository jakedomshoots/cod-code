package ceo

import (
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

// Test_renderCEODelegationPrompt_mentions_alpha_cod_and_keeps_json_contract pins the
// lore ("Alpha Cod") in the CEO delegation planner prompt while preserving the JSON-only
// contract and the "candidate_subagents" / "selected_subagents" markers the rest of the
// suite already depends on.
func Test_renderCEODelegationPrompt_mentions_alpha_cod_and_keeps_json_contract(t *testing.T) {
	// Given
	packet := jobpacket.Packet{
		Task: "Audit migration risk",
		TaskProfile: jobpacket.TaskProfile{
			Kind:      "review",
			RiskLevel: "medium",
		},
		Subagents: []jobpacket.Subagent{
			{Name: "scanner", Role: "scan repo"},
			{Name: "coder", Role: "apply bounded patches"},
		},
		MaxSubagents: 3,
	}

	// When
	prompt := renderCEODelegationPrompt(packet)

	// Then: the new lore identity line is present.
	if !strings.Contains(prompt, "You are the Alpha Cod") {
		t.Fatalf("CEO delegation prompt missing Alpha Cod identity line:\n%s", prompt)
	}
	if !strings.Contains(prompt, "CEO delegation planner") {
		t.Fatalf("CEO delegation prompt missing planner role marker:\n%s", prompt)
	}

	// Then: the JSON-only contract and structural markers remain intact.
	for _, want := range []string{
		"Return JSON only",
		`"selected_subagents"`,
		"candidate_subagents:",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("CEO delegation prompt missing marker %q:\n%s", want, prompt)
		}
	}
}

// Test_renderCEOReviewPrompt_mentions_alpha_cod_and_keeps_subagents_marker pins the
// "Alpha Cod" identity in the CEO final review prompt while preserving the JSON-only
// contract and the "subagents:" header the rest of the suite already depends on.
func Test_renderCEOReviewPrompt_mentions_alpha_cod_and_keeps_subagents_marker(t *testing.T) {
	// Given
	input := ceoReviewInput{
		Packet: jobpacket.Packet{
			Task: "Finalize migration review",
			TaskProfile: jobpacket.TaskProfile{
				Kind:      "review",
				RiskLevel: "low",
			},
		},
		Results: []subagent.Result{
			{AgentName: "scanner", Status: "pass", Summary: "scanned"},
			{AgentName: "coder", Status: "pass", Summary: "patched"},
		},
		GuardVerdict: "pass",
	}

	// When
	prompt := renderCEOReviewPrompt(input)

	// Then: the new lore identity line is present.
	if !strings.Contains(prompt, "You are the Alpha Cod") {
		t.Fatalf("CEO review prompt missing Alpha Cod identity line:\n%s", prompt)
	}
	if !strings.Contains(prompt, "CEO final reviewer") {
		t.Fatalf("CEO review prompt missing final reviewer role marker:\n%s", prompt)
	}

	// Then: the JSON-only contract and structural markers remain intact.
	for _, want := range []string{
		"Return JSON only",
		`"recommended_verdict"`,
		"subagents:",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("CEO review prompt missing marker %q:\n%s", want, prompt)
		}
	}

	// Then: the reviewed subagent names are surfaced via the marker so the model can
	// ground its verdict on real evidence.
	for _, want := range []string{"- scanner", "- coder"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("CEO review prompt missing reviewed subagent %q:\n%s", want, prompt)
		}
	}
}
