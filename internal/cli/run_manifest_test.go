package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_prints_compact_run_manifest(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--check", "go", "version", "--", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RunManifest struct {
			SchemaVersion     int    `json:"schema_version"`
			ContextMode       string `json:"context_mode"`
			SubagentCount     int    `json:"subagent_count"`
			CheckAttemptCount int    `json:"check_attempt_count"`
			Verdict           string `json:"verdict"`
		} `json:"run_manifest"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.RunManifest.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", body.RunManifest.SchemaVersion)
	}
	if body.RunManifest.ContextMode != "lean" {
		t.Fatalf("ContextMode = %q, want lean", body.RunManifest.ContextMode)
	}
	if body.RunManifest.SubagentCount != 3 || body.RunManifest.CheckAttemptCount != 1 {
		t.Fatalf("run manifest = %#v, want 3 subagents and 1 check attempt", body.RunManifest)
	}
	if body.RunManifest.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.RunManifest.Verdict)
	}
}
