package ceo

import (
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/model"
)

func Test_parseCEODelegationResponse_parses_fenced_json(t *testing.T) {
	// Given
	response := model.Response{
		Text: "```json\n{\"selected_subagents\":[\"coder\"],\"summary\":\"Use coder.\"}\n```",
	}
	candidates := []jobpacket.Subagent{{Name: "coder", Role: "apply bounded changes"}}

	// When
	selected, delegation, err := parseCEODelegationResponse(response, candidates)

	// Then
	if err != nil {
		t.Fatalf("parseCEODelegationResponse returned error: %v", err)
	}
	if len(selected) != 1 || selected[0].Name != "coder" {
		t.Fatalf("selected = %+v, want coder", selected)
	}
	if delegation.Summary != "Use coder." {
		t.Fatalf("Summary = %q, want Use coder.", delegation.Summary)
	}
}

func Test_parseCEOReviewResponse_parses_fenced_json(t *testing.T) {
	// Given
	response := model.Response{
		Text: "```json\n{\"recommended_verdict\":\"pass\",\"summary\":\"Evidence checked.\"}\n```",
	}

	// When
	review, err := parseCEOReviewResponse(response)

	// Then
	if err != nil {
		t.Fatalf("parseCEOReviewResponse returned error: %v", err)
	}
	if review.RecommendedVerdict != "pass" || review.Summary != "Evidence checked." {
		t.Fatalf("review = %+v, want pass review", review)
	}
}
