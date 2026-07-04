package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_production_status_reads_latest_readiness_summary(t *testing.T) {
	root := t.TempDir()
	launchBody := "# Launch Checklist\n\n- one\n- two\n"
	launchSHA := sha256Text(launchBody)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "launch-checklist.md"), strings.TrimSuffix(launchBody, "\n"))
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 2,
  "blocked_checks": ["release.public_release_ready", "provider.openai_http_provider"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "sha256": "`+launchSHA+`",
    "required_action_count": 2,
    "status": "pass"
  }
}`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body productionStatusReport
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode production status: %v\n%s", err, out.String())
	}
	if body.Status != "blocked" || !body.LocalProductionReady || body.PublicProductionReady {
		t.Fatalf("body = %+v, want blocked local-ready public-not-ready", body)
	}
	if body.BlockedCount != 2 || len(body.BlockedChecks) != 2 {
		t.Fatalf("blocked = %d/%v, want two blocked checks", body.BlockedCount, body.BlockedChecks)
	}
	if body.LaunchChecklist == nil || body.LaunchChecklist.RequiredActionCount != 2 || body.LaunchChecklist.SHA256 != launchSHA || body.LaunchChecklist.CurrentSHA256 != launchSHA || body.LaunchChecklist.MatchesDeclared == nil || *body.LaunchChecklist.MatchesDeclared != true {
		t.Fatalf("launch checklist = %+v, want fingerprinted checklist", body.LaunchChecklist)
	}
	if !strings.Contains(body.NextAction, "launch-checklist.md") {
		t.Fatalf("next action = %q, want launch checklist path", body.NextAction)
	}
}

func Test_Run_production_status_text_reports_missing_evidence(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Production status: missing",
		"Local ready: false",
		"Public ready: false",
		"Next action: run sh scripts/production-readiness.sh",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("production status text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_production_status_prefers_finalizer_next_actions(t *testing.T) {
	root := t.TempDir()
	launchBody := "# Launch Checklist\n\n- action\n"
	launchSHA := sha256Text(launchBody)
	setupBody := "# Production Setup Actions\n\n- release\n- provider\n- rerun\n- verify\n"
	setupSHA := sha256Text(setupBody)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "launch-checklist.md"), strings.TrimSuffix(launchBody, "\n"))
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "setup-actions.md"), strings.TrimSuffix(setupBody, "\n"))
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 1,
  "blocked_checks": ["comparison.all_agent_29_task_comparison"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "sha256": "`+launchSHA+`",
    "required_action_count": 5,
    "status": "pass"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "summary.json"), `{
  "status": "blocked",
  "next_actions": {
    "path": "next-actions.md",
    "json_path": "next-actions.json",
    "required_action_count": 2
  },
  "setup_actions": {
    "path": "setup-actions.md",
    "required_action_count": 4,
    "sha256": "`+setupSHA+`"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "next-actions.json"), `{
  "status": "blocked",
  "actions": [
    {
      "id": "provider-openai",
      "kind": "provider_proof",
      "required_env": "OPENAI_API_KEY",
      "evidence": "`+filepath.ToSlash(filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "index.md"))+`"
    },
    {
      "id": "production-readiness",
      "kind": "final_readiness",
      "status": "blocked"
    }
  ]
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "summary.json"), `{
  "status": "blocked",
  "provider": "openai",
  "provider_mode": "http-provider",
  "http_preset": "openai",
  "http_model": "gpt-5",
  "api_key_env": "OPENAI_API_KEY",
  "blocked_reason": "missing_api_key_env",
  "secret_value_saved": false,
  "artifacts": {}
}`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Launch checklist: launch-checklist.md (5 actions) declared_match=true",
		"Finalizer next actions:",
		"production-finalize-r1/next-actions.md (2 actions)",
		"Finalizer actions JSON:",
		"production-finalize-r1/next-actions.json",
		"Finalizer action states: missing_env=1 waiting=1",
		"Finalizer commands: runnable=0 blocked=0",
		"Finalizer evidence matches: declared=0 mismatched=0",
		"Finalizer setup actions:",
		"production-finalize-r1/setup-actions.md",
		"(4 actions)",
		"sha256=" + setupSHA,
		"declared_match=true",
		"Next action: open ",
		"production-finalize-r1/next-actions.md",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("production status text missing %q:\n%s", want, text)
		}
	}
}

func sha256Text(body string) string {
	sum := sha256.Sum256([]byte(body))
	return fmt.Sprintf("%x", sum[:])
}

func Test_Run_production_status_ignores_skipped_finalizer_next_actions(t *testing.T) {
	root := t.TempDir()
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 1,
  "blocked_checks": ["provider.openai_http_provider"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "required_action_count": 5,
    "status": "pass"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-complete", "summary.json"), `{
  "status": "blocked",
  "skipped_steps": [],
  "next_actions": {
    "path": "next-actions.md",
    "json_path": "next-actions.json",
    "required_action_count": 6
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-skipped", "summary.json"), `{
  "status": "blocked",
  "skipped_steps": ["release-readiness", "provider-openai"],
  "next_actions": {
    "path": "next-actions.md",
    "required_action_count": 2
  }
}`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	if !strings.Contains(text, "production-finalize-complete/next-actions.md (6 actions)") {
		t.Fatalf("production status did not prefer complete finalizer:\n%s", text)
	}
	if strings.Contains(text, "production-finalize-skipped/next-actions.md") {
		t.Fatalf("production status used skipped finalizer:\n%s", text)
	}
}

func Test_ParseArgs_sets_production_status_from_verb(t *testing.T) {
	opts, err := parseArgs([]string{"production-status", "--workspace", "/tmp/workspace"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if !opts.showProductionStatus || opts.workspaceDir != "/tmp/workspace" {
		t.Fatalf("opts = %+v, want production status for workspace", opts)
	}
}

func writeProductionStatusSummary(t *testing.T, path string, body string) {
	t.Helper()
	writeProductionReadinessText(t, path, body+"\n")
}
