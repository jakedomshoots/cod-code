package ceo

import (
	"context"
	"strings"
	"sync"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type continueCountingRunner struct {
	mu    sync.Mutex
	calls []string
}

func (r *continueCountingRunner) Run(_ context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	r.mu.Lock()
	r.calls = append(r.calls, packet.AgentName)
	r.mu.Unlock()
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		ContextReceived: packet.ContextMode,
		PriorFindings:   packet.PriorFindings,
		Summary:         packet.AgentName + " saw " + packet.PriorFindings,
		Evidence:        []string{packet.AgentName + " evidence"},
	}, nil
}

func Test_Runtime_RunJob_reuses_matching_passed_subagents_when_continuing(t *testing.T) {
	// Given
	runner := &continueCountingRunner{}
	runtime := NewRuntimeWithSubagentRunner(runner)
	scanner := jobpacket.Subagent{Name: "scanner", Role: "inspect scope", AllowedActions: jobpacket.DefaultActionsForAgent("scanner")}
	coder := jobpacket.Subagent{Name: "coder", Role: "apply patch", AllowedActions: jobpacket.DefaultActionsForAgent("coder")}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix checkout",
		Subagents: []jobpacket.Subagent{
			scanner,
			coder,
		},
		Continuation: &ContinuationContext{
			JobID: "job-000001",
			ReusableResults: []subagent.Result{
				{
					AgentName:      "scanner",
					Role:           "inspect scope",
					Status:         "pass",
					Summary:        "cached scanner summary",
					Evidence:       []string{"cached evidence"},
					AllowedActions: jobpacket.ActionStrings(scanner.AllowedActions),
				},
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if got := strings.Join(runner.calls, ","); got != "coder" {
		t.Fatalf("runner calls = %q, want only coder", got)
	}
	if !report.SubagentResults[0].Reused {
		t.Fatal("scanner Reused = false, want true")
	}
	if report.RunManifest.ReusedSubagentCount != 1 {
		t.Fatalf("ReusedSubagentCount = %d, want 1", report.RunManifest.ReusedSubagentCount)
	}
	if report.Continuation == nil || report.Continuation.ReusedSubagentCount != 1 {
		t.Fatalf("Continuation = %+v, want reused count", report.Continuation)
	}
	if !strings.Contains(report.SubagentResults[1].PriorFindings, "cached scanner summary") {
		t.Fatalf("coder PriorFindings = %q, want cached scanner summary", report.SubagentResults[1].PriorFindings)
	}
	if !hasReusedScannerEvent(report.RunEvents) {
		t.Fatalf("RunEvents = %+v, want reused scanner history event", report.RunEvents)
	}
}

func Test_Runtime_RunJob_skips_ceo_delegation_when_using_saved_delegation(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"recommended_verdict":"pass","summary":"saved delegation reviewed"}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)
	coder := jobpacket.Subagent{Name: "coder", Role: "apply bounded changes", AllowedActions: jobpacket.DefaultActionsForAgent("coder")}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:      "Fix checkout",
		Subagents: []jobpacket.Subagent{coder},
		Continuation: &ContinuationContext{
			JobID:              "job-000001",
			UseSavedDelegation: true,
			SavedDelegation: &CEODelegation{
				Source:            "model",
				SelectedSubagents: []string{"coder"},
				Summary:           "original coder lane",
				PromptBytes:       123,
			},
			ReusableResults: []subagent.Result{
				{
					AgentName:      "coder",
					Role:           "apply bounded changes",
					Status:         "pass",
					Summary:        "cached coder summary",
					AllowedActions: jobpacket.ActionStrings(coder.AllowedActions),
				},
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil ||
		report.CEODelegation.Source != "history" ||
		report.CEODelegation.SelectedSubagents[0] != "coder" {
		t.Fatalf("CEODelegation = %#v, want history-sourced saved delegation", report.CEODelegation)
	}
	if report.CEOReview == nil || report.CEOReview.Summary != "saved delegation reviewed" {
		t.Fatalf("CEOReview = %#v, want final review", report.CEOReview)
	}
	if len(client.prompts) != 1 || strings.Contains(client.prompts[0], "candidate_subagents") {
		t.Fatalf("CEO prompts = %#v, want final review only", client.prompts)
	}
	if len(report.SubagentResults) != 1 || !report.SubagentResults[0].Reused {
		t.Fatalf("SubagentResults = %#v, want reused coder", report.SubagentResults)
	}
}

func hasReusedScannerEvent(events []RunEvent) bool {
	for _, event := range events {
		if event.Kind == "subagent" && event.AgentName == "scanner" && event.Source == "history" {
			return true
		}
	}
	return false
}
