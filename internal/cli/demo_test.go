package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func Test_Run_runs_golden_demo_when_demo_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--demo"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
		PatchResults []struct {
			Path string `json:"path"`
			Diff string `json:"diff"`
		} `json:"patch_results"`
		PatchAudit []struct {
			Path      string `json:"path"`
			Source    string `json:"source"`
			AgentName string `json:"agent_name"`
		} `json:"patch_audit"`
		CheckResults []struct {
			Status string `json:"status"`
		} `json:"check_results"`
		RunEvents []struct {
			Kind   string `json:"kind"`
			Status string `json:"status"`
		} `json:"run_events"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if !strings.Contains(body.JobPacket.Task, "Golden demo") || !strings.Contains(body.JobPacket.Task, "ceo-harness-demo") {
		t.Fatalf("task = %q, want golden demo workspace marker", body.JobPacket.Task)
	}
	if len(body.PatchResults) != 1 || body.PatchResults[0].Path != "app.txt" || body.PatchResults[0].Diff == "" {
		t.Fatalf("PatchResults = %+v, want app.txt diff", body.PatchResults)
	}
	if len(body.PatchAudit) != 1 || body.PatchAudit[0].Source != "model" || body.PatchAudit[0].AgentName != "coder" {
		t.Fatalf("PatchAudit = %+v, want coder model patch", body.PatchAudit)
	}
	if len(body.CheckResults) != 1 || body.CheckResults[0].Status != "pass" {
		t.Fatalf("CheckResults = %+v, want one passing check", body.CheckResults)
	}
	if !hasRunEvent(body.RunEvents, "patch", "applied") || !hasRunEvent(body.RunEvents, "check", "pass") || !hasRunEvent(body.RunEvents, "verdict", "pass") {
		t.Fatalf("RunEvents = %+v, want patch/check/verdict proof", body.RunEvents)
	}
}

func hasRunEvent(events []struct {
	Kind   string `json:"kind"`
	Status string `json:"status"`
}, kind string, status string) bool {
	for _, event := range events {
		if event.Kind == kind && event.Status == status {
			return true
		}
	}
	return false
}
