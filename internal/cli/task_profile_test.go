package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_prints_task_profile_and_history_metadata_when_task_is_mixed_risk(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Research", "auth", "bug", "and", "implement", "fix"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var report struct {
		JobPacket struct {
			TaskProfile struct {
				Kind      string   `json:"kind"`
				RiskLevel string   `json:"risk_level"`
				RiskAreas []string `json:"risk_areas"`
			} `json:"task_profile"`
		} `json:"job_packet"`
		RunManifest struct {
			TaskKind  string   `json:"task_kind"`
			RiskLevel string   `json:"risk_level"`
			RiskAreas []string `json:"risk_areas"`
		} `json:"run_manifest"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &report); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if report.JobPacket.TaskProfile.Kind != "mixed" || report.JobPacket.TaskProfile.RiskLevel != "high" {
		t.Fatalf("task profile = %#v, want mixed/high", report.JobPacket.TaskProfile)
	}
	if report.RunManifest.TaskKind != "mixed" || report.RunManifest.RiskLevel != "high" {
		t.Fatalf("run manifest = %#v, want mixed/high", report.RunManifest)
	}
	assertStringSlice(t, report.JobPacket.TaskProfile.RiskAreas, []string{"security"})
	assertStringSlice(t, report.RunManifest.RiskAreas, []string{"security"})

	out.Reset()
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history"})
	if err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var historyBody struct {
		History []struct {
			TaskKind  string   `json:"task_kind"`
			RiskLevel string   `json:"risk_level"`
			RiskAreas []string `json:"risk_areas"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &historyBody); jsonErr != nil {
		t.Fatalf("history output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(historyBody.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(historyBody.History))
	}
	if historyBody.History[0].TaskKind != "mixed" || historyBody.History[0].RiskLevel != "high" {
		t.Fatalf("history profile = %#v, want mixed/high", historyBody.History[0])
	}
	assertStringSlice(t, historyBody.History[0].RiskAreas, []string{"security"})
}

func Test_Run_prints_risk_specific_subagents_for_database_billing_release_task(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--max-subagents", "7", "Research", "payment", "database", "migration", "and", "deploy", "the", "fix"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var report struct {
		JobPacket struct {
			TaskProfile struct {
				RiskAreas []string `json:"risk_areas"`
			} `json:"task_profile"`
			Subagents []struct {
				Name string `json:"name"`
			} `json:"subagents"`
		} `json:"job_packet"`
		JobOwner string `json:"job_owner"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &report); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	assertStringSlice(t, report.JobPacket.TaskProfile.RiskAreas, []string{"billing", "database", "release"})
	assertSubagentNames(t, report.JobPacket.Subagents, []string{"planner", "researcher", "coder", "billing", "database", "release", "reviewer"})
	if report.JobOwner != "coder" {
		t.Fatalf("JobOwner = %q, want coder", report.JobOwner)
	}
}

func Test_Run_prints_task_profile_counts_when_history_summary_is_requested(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	tasks := [][]string{
		{"--workspace", root, "Plan", "the", "roadmap"},
		{"--workspace", root, "Research", "auth", "bug", "and", "implement", "fix"},
	}
	for _, args := range tasks {
		if err := Run(context.Background(), &bytes.Buffer{}, args); err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--history", "--summary-only"})
	// Then
	if err != nil {
		t.Fatalf("Run history returned error: %v", err)
	}
	var body struct {
		Summary struct {
			TaskKindCounts  map[string]int `json:"task_kind_counts"`
			RiskLevelCounts map[string]int `json:"risk_level_counts"`
		} `json:"summary"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Summary.TaskKindCounts["planning"] != 1 || body.Summary.TaskKindCounts["mixed"] != 1 {
		t.Fatalf("task kind counts = %#v, want planning and mixed", body.Summary.TaskKindCounts)
	}
	if body.Summary.RiskLevelCounts["low"] != 1 || body.Summary.RiskLevelCounts["high"] != 1 {
		t.Fatalf("risk counts = %#v, want low and high", body.Summary.RiskLevelCounts)
	}
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %#v, want %#v", got, want)
	}
	for index, wantValue := range want {
		if got[index] != wantValue {
			t.Fatalf("strings[%d] = %q, want %q", index, got[index], wantValue)
		}
	}
}

func assertSubagentNames(t *testing.T, got []struct {
	Name string `json:"name"`
}, want []string,
) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("subagents = %#v, want %v", got, want)
	}
	for index, wantName := range want {
		if got[index].Name != wantName {
			t.Fatalf("subagents[%d].Name = %q, want %q", index, got[index].Name, wantName)
		}
	}
}
