package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_LoadCompetitors_accepts_required_competitors_when_config_complete(t *testing.T) {
	// Given
	path := writeCompetitorFixture(t, completeCompetitorEntries(nil))

	// When
	config, err := LoadCompetitors(path)

	// Then
	if err != nil {
		t.Fatalf("LoadCompetitors returned error: %v", err)
	}
	if len(config.Competitors) != 5 {
		t.Fatalf("len(config.Competitors) = %d, want 5", len(config.Competitors))
	}
	if config.Competitors[0].ID != "codex_cli" {
		t.Fatalf("first competitor ID = %q, want codex_cli", config.Competitors[0].ID)
	}
}

func Test_LoadCompetitors_rejects_config_when_required_competitor_missing(t *testing.T) {
	// Given
	entries := completeCompetitorEntries(nil)
	path := writeCompetitorFixture(t, entries[:len(entries)-1])

	// When
	_, err := LoadCompetitors(path)

	// Then
	if !errors.Is(err, ErrInvalidCompetitor) {
		t.Fatalf("error = %v, want ErrInvalidCompetitor", err)
	}
}

func Test_BuildComparisonPlan_records_skipped_missing_binary_when_binary_absent(t *testing.T) {
	// Given
	t.Setenv("PATH", t.TempDir())
	path := writeCompetitorFixture(t, completeCompetitorEntries(map[string]string{
		"codex_cli": "missing-ceo-harness-competitor",
	}))
	config, err := LoadCompetitors(path)
	if err != nil {
		t.Fatalf("LoadCompetitors returned error: %v", err)
	}

	// When
	plan, err := BuildComparisonPlan(context.Background(), config)

	// Then
	if err != nil {
		t.Fatalf("BuildComparisonPlan returned error: %v", err)
	}
	result := requireComparisonResult(t, plan, "codex_cli")
	if result.Status != "skipped_missing_binary" {
		t.Fatalf("Status = %q, want skipped_missing_binary", result.Status)
	}
	if result.Status == "fail" {
		t.Fatalf("missing binary must not be recorded as fail")
	}
}

func Test_BuildComparisonPlan_emits_empty_placeholder_when_binary_exists(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "codex"))
	t.Setenv("PATH", binDir)
	path := writeCompetitorFixture(t, completeCompetitorEntries(map[string]string{
		"codex_cli": "codex",
	}))
	config, err := LoadCompetitors(path)
	if err != nil {
		t.Fatalf("LoadCompetitors returned error: %v", err)
	}

	// When
	plan, err := BuildComparisonPlan(context.Background(), config)

	// Then
	if err != nil {
		t.Fatalf("BuildComparisonPlan returned error: %v", err)
	}
	result := requireComparisonResult(t, plan, "codex_cli")
	if result.Status != "planned_no_result" {
		t.Fatalf("Status = %q, want planned_no_result", result.Status)
	}
	if len(result.EvidencePaths) != 0 {
		t.Fatalf("EvidencePaths = %#v, want empty placeholder evidence", result.EvidencePaths)
	}
	if strings.Join(result.Command, " ") != "codex --version" {
		t.Fatalf("Command = %#v, want codex version placeholder", result.Command)
	}
}

func Test_RunCLI_validates_competitor_config_when_validate_flag_is_set(t *testing.T) {
	// Given
	path := writeCompetitorFixture(t, completeCompetitorEntries(nil))
	var out bytes.Buffer
	var errOut bytes.Buffer

	// When
	err := RunCLI(context.Background(), &out, &errOut, []string{"--validate-competitors", "--competitors", path})

	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v\nstderr: %s", err, errOut.String())
	}
	if !strings.Contains(out.String(), "competitors_valid=true count=5") {
		t.Fatalf("stdout = %q, want valid competitor count", out.String())
	}
}

func requireComparisonResult(t *testing.T, plan ComparisonPlan, id string) ComparisonResult {
	t.Helper()
	for _, result := range plan.Results {
		if result.ID == id {
			return result
		}
	}
	t.Fatalf("missing result for %s in %#v", id, plan.Results)
	return ComparisonResult{}
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}
}

func writeCompetitorFixture(t *testing.T, competitors []map[string]any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "competitors.json")
	payload := map[string]any{
		"schema_version": 1,
		"competitors":    competitors,
	}
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func completeCompetitorEntries(binaryOverrides map[string]string) []map[string]any {
	ids := []string{"codex_cli", "claude_code", "aider", "opencode", "goose"}
	names := map[string]string{
		"codex_cli":   "OpenAI Codex CLI",
		"claude_code": "Claude Code",
		"aider":       "Aider",
		"opencode":    "OpenCode",
		"goose":       "Goose",
	}
	binaries := map[string]string{
		"codex_cli":   "codex",
		"claude_code": "claude",
		"aider":       "aider",
		"opencode":    "opencode",
		"goose":       "goose",
	}
	for id, binary := range binaryOverrides {
		binaries[id] = binary
	}
	entries := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, map[string]any{
			"id":                    id,
			"name":                  names[id],
			"binary":                binaries[id],
			"homepage":              "https://example.com/" + id,
			"setup_hint":            "Install " + names[id] + " before running comparison tasks.",
			"version_args":          []string{"--version"},
			"dry_run_args":          []string{"--version"},
			"timeout_seconds":       1800,
			"comparison_dimensions": requiredCompetitorDimensionsFixture(),
		})
	}
	return entries
}

func requiredCompetitorDimensionsFixture() []string {
	return []string{
		"task_success",
		"time_to_complete",
		"files_changed",
		"safety_prompts",
		"cost_provider_used",
		"evidence_quality",
	}
}
