package ceo

import (
	"context"
	"testing"
)

func Test_Runtime_RunJob_includes_compact_run_manifest(t *testing.T) {
	// Given
	runtime := NewRuntime()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                "Fix a failing test",
		CheckCommand:        []string{"go", "version"},
		SubagentConcurrency: 2,
		MaxToolRequests:     4,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	manifest := report.RunManifest
	if manifest.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", manifest.SchemaVersion)
	}
	if manifest.TaskBytes != len("Fix a failing test") {
		t.Fatalf("TaskBytes = %d, want task length", manifest.TaskBytes)
	}
	if manifest.ContextMode != "lean" || manifest.MaxContextBytes != 4096 {
		t.Fatalf("context manifest = %#v, want lean 4096", manifest)
	}
	if manifest.SubagentCount != 3 || manifest.CheckAttemptCount != 1 {
		t.Fatalf("run counts = %#v, want 3 subagents and 1 check attempt", manifest)
	}
	if manifest.SubagentConcurrency != 2 {
		t.Fatalf("SubagentConcurrency = %d, want 2", manifest.SubagentConcurrency)
	}
	if manifest.MaxToolRequests != 4 {
		t.Fatalf("MaxToolRequests = %d, want 4", manifest.MaxToolRequests)
	}
	if manifest.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", manifest.Verdict)
	}
}
