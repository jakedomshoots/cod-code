package ceo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_buildHistoryEntry_preserves_no_progress_stop_count(t *testing.T) {
	// Given
	report := Report{
		JobPacket: jobpacket.Packet{Task: "Fix repeated weak output"},
		SubagentResults: []subagent.Result{
			{AgentName: "scanner", NoProgressStopped: true},
			{AgentName: "coder"},
		},
		VerificationSummary: VerificationSummary{
			SubagentAttemptCount:        4,
			SubagentRetryCount:          2,
			SubagentNoProgressStopCount: 1,
		},
		Verdict: "fail",
	}

	// When
	entry := buildHistoryEntry(report)

	// Then
	if entry.SubagentNoProgressStopCount != 1 {
		t.Fatalf("SubagentNoProgressStopCount = %d, want 1", entry.SubagentNoProgressStopCount)
	}
}

func Test_Runtime_RunJob_persists_schema_version_when_workspace_report_is_saved(t *testing.T) {
	// Given
	root := t.TempDir()
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Save schema version",
		WorkspaceDir: root,
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", report.SchemaVersion)
	}
	snapshot, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json"))
	if err != nil {
		t.Fatalf("read report snapshot: %v", err)
	}
	var body struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(snapshot, &body); err != nil {
		t.Fatalf("snapshot must be JSON: %v\n%s", err, string(snapshot))
	}
	if body.SchemaVersion != 1 {
		t.Fatalf("snapshot schema_version = %d, want 1", body.SchemaVersion)
	}
}
