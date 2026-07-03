package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_includes_check_duration_when_check_runs(t *testing.T) {
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
		CheckResults []struct {
			DurationMS *int64 `json:"duration_ms"`
		} `json:"check_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want 1", len(body.CheckResults))
	}
	if body.CheckResults[0].DurationMS == nil {
		t.Fatal("duration_ms missing from check result")
	}
	if *body.CheckResults[0].DurationMS < 0 {
		t.Fatalf("duration_ms = %d, want nonnegative duration", *body.CheckResults[0].DurationMS)
	}
}
