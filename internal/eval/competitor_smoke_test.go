package eval

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_RunCompetitorSmoke_records_missing_binary_as_skip_when_tool_is_absent(t *testing.T) {
	// Given
	t.Setenv("PATH", t.TempDir())
	configPath := writeCompetitorFixture(t, completeCompetitorEntries(map[string]string{
		"codex_cli": "missing-ceo-harness-competitor",
	}))
	outputDir := filepath.Join(t.TempDir(), "smoke")

	// When
	summary, err := RunCompetitorSmoke(context.Background(), CompetitorSmokeRequest{
		CompetitorsPath: configPath,
		OutputDir:       outputDir,
		TimeoutSeconds:  5,
	})
	// Then
	if err != nil {
		t.Fatalf("RunCompetitorSmoke returned error: %v", err)
	}
	result := requireCompetitorSmokeResult(t, summary, "codex_cli")
	if result.Status != "skipped_missing_binary" {
		t.Fatalf("Status = %q, want skipped_missing_binary", result.Status)
	}
	if result.SetupHint == "" {
		t.Fatalf("SetupHint is empty, want install guidance")
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
}

func Test_RunCompetitorSmoke_runs_version_and_dry_run_when_binary_exists(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutable(t, filepath.Join(binDir, "codex"))
	t.Setenv("PATH", binDir)
	configPath := writeCompetitorFixture(t, completeCompetitorEntries(map[string]string{
		"codex_cli": "codex",
	}))
	outputDir := filepath.Join(t.TempDir(), "smoke")

	// When
	summary, err := RunCompetitorSmoke(context.Background(), CompetitorSmokeRequest{
		CompetitorsPath: configPath,
		OutputDir:       outputDir,
		TimeoutSeconds:  5,
	})
	// Then
	if err != nil {
		t.Fatalf("RunCompetitorSmoke returned error: %v", err)
	}
	result := requireCompetitorSmokeResult(t, summary, "codex_cli")
	if result.Status != "smoke_pass" {
		t.Fatalf("Status = %q, want smoke_pass; result = %+v", result.Status, result)
	}
	if len(result.EvidencePaths) != 4 {
		t.Fatalf("EvidencePaths = %#v, want stdout/stderr for version and dry-run", result.EvidencePaths)
	}
	requireFile(t, filepath.Join(outputDir, "codex_cli", "version.stdout"))
	requireFile(t, filepath.Join(outputDir, "codex_cli", "version.stderr"))
	requireFile(t, filepath.Join(outputDir, "codex_cli", "dry-run.stdout"))
	requireFile(t, filepath.Join(outputDir, "codex_cli", "dry-run.stderr"))
}

func Test_RunCompetitorSmoke_marks_provider_quota_as_setup_blocked(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "opencode"), `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf 'opencode 1.0.0\n'
  exit 0
fi
printf 'AI_APICallError: Token Plan usage limit reached\n' >&2
exit 1
`)
	t.Setenv("PATH", binDir)
	entries := completeCompetitorEntries(map[string]string{
		"opencode": "opencode",
	})
	for _, entry := range entries {
		if entry["id"] == "opencode" {
			entry["dry_run_args"] = []string{"run", "--print-logs", "--log-level", "INFO", "Reply exactly CEO_HARNESS_EVAL_OK."}
		}
	}
	configPath := writeCompetitorFixture(t, entries)
	outputDir := filepath.Join(t.TempDir(), "smoke")

	// When
	summary, err := RunCompetitorSmoke(context.Background(), CompetitorSmokeRequest{
		CompetitorsPath: configPath,
		OutputDir:       outputDir,
		TimeoutSeconds:  5,
	})
	// Then
	if err != nil {
		t.Fatalf("RunCompetitorSmoke returned error: %v", err)
	}
	result := requireCompetitorSmokeResult(t, summary, "opencode")
	if result.Status != competitorSmokeStatusBlocked || summary.SetupBlocked != 1 {
		t.Fatalf("result=%+v summary=%+v, want setup-blocked", result, summary)
	}
	stderr, err := os.ReadFile(filepath.Join(outputDir, "opencode", "dry-run.stderr"))
	if err != nil {
		t.Fatalf("read dry-run stderr: %v", err)
	}
	if !strings.Contains(string(stderr), "Token Plan usage limit reached") {
		t.Fatalf("dry-run stderr missing quota evidence:\n%s", stderr)
	}
}

func Test_RunCLI_runs_competitor_smoke_when_flag_is_set(t *testing.T) {
	// Given
	t.Setenv("PATH", t.TempDir())
	configPath := writeCompetitorFixture(t, completeCompetitorEntries(nil))
	outputDir := filepath.Join(t.TempDir(), "smoke")

	// When
	err := RunCLI(context.Background(), os.Stdout, os.Stderr, []string{
		"--comparison-smoke",
		"--competitors", configPath,
		"--output-dir", outputDir,
		"--timeout-seconds", "1",
	})
	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v", err)
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
}

func requireCompetitorSmokeResult(t *testing.T, summary CompetitorSmokeSummary, id string) CompetitorSmokeResult {
	t.Helper()
	for _, result := range summary.Results {
		if result.ID == id {
			return result
		}
	}
	t.Fatalf("missing smoke result for %s in %+v", id, summary.Results)
	return CompetitorSmokeResult{}
}
