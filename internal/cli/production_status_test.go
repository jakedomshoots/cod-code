package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "release-bootstrap-r1", "summary.json"), `{
  "status": "blocked",
  "blocked_count": 3,
  "version": "0.2.0-test",
  "artifacts": {
    "handoff": "release-handoff.md"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "release-bootstrap-r1", "release-handoff.md"), `# Public Release Handoff
`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "provider-setup-preflight", "summary.json"), `{
  "status": "blocked",
  "provider_count": 3,
  "ready_count": 1,
  "blocked_count": 2,
  "blocked_providers": ["openrouter", "moonshot"],
  "secret_value_saved": false
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
		"Release bootstrap: blocked blocked=3 version=0.2.0-test",
		"Release handoff:",
		"release-bootstrap-r1/release-handoff.md",
		"Provider setup preflight: blocked ready=1 blocked=2 providers=3 blocked_providers=openrouter,moonshot",
		"provider-setup-preflight/summary.json",
		"Next action: open ",
		"production-finalize-r1/next-actions.md",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("production status text missing %q:\n%s", want, text)
		}
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body productionStatusReport
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode production status: %v\n%s", err, out.String())
	}
	if body.ReleaseBootstrap == nil || body.ReleaseBootstrap.Status != "blocked" || body.ReleaseBootstrap.BlockedCount != 3 || !strings.Contains(body.ReleaseBootstrap.HandoffPath, "release-handoff.md") {
		t.Fatalf("release bootstrap = %+v, want surfaced handoff", body.ReleaseBootstrap)
	}
	if body.ProviderSetupPreflight == nil || body.ProviderSetupPreflight.Status != "blocked" || body.ProviderSetupPreflight.ReadyCount != 1 || body.ProviderSetupPreflight.BlockedCount != 2 || len(body.ProviderSetupPreflight.BlockedProviders) != 2 {
		t.Fatalf("provider setup preflight = %+v, want surfaced provider setup status", body.ProviderSetupPreflight)
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
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-planned", "summary.json"), `{
  "status": "planned",
  "skipped_steps": [],
  "next_actions": {
    "path": "next-actions.md",
    "required_action_count": 9
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
	if strings.Contains(text, "production-finalize-planned/next-actions.md") {
		t.Fatalf("production status used planned finalizer:\n%s", text)
	}
}

func Test_Run_production_status_ignores_skipped_readiness_summaries(t *testing.T) {
	root := t.TempDir()
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-complete", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 2,
  "blocked_checks": ["release.public_release_readiness_run", "release.public_release_ready"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "required_action_count": 2,
    "status": "pass"
  },
  "checks": [
    {"category": "release", "name": "public_release_readiness_run", "status": "blocked"},
    {"category": "release", "name": "public_release_ready", "status": "blocked"}
  ]
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-skipped", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 1,
  "blocked_checks": ["release.public_release_ready"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "required_action_count": 1,
    "status": "pass"
  },
  "checks": [
    {"category": "release", "name": "public_release_readiness_run", "status": "skipped"},
    {"category": "release", "name": "public_release_ready", "status": "blocked"}
  ]
}`)
	oldTime := time.Now().Add(-2 * time.Hour)
	newTime := time.Now()
	if err := os.Chtimes(filepath.Join(root, ".omo", "evidence", "production-readiness-complete", "summary.json"), oldTime, oldTime); err != nil {
		t.Fatalf("touch complete readiness: %v", err)
	}
	if err := os.Chtimes(filepath.Join(root, ".omo", "evidence", "production-readiness-skipped", "summary.json"), newTime, newTime); err != nil {
		t.Fatalf("touch skipped readiness: %v", err)
	}

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-status", "--workspace", root}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body productionStatusReport
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode production status: %v\n%s", err, out.String())
	}
	if !strings.Contains(body.SummaryPath, "production-readiness-complete") || body.BlockedCount != 2 {
		t.Fatalf("status used wrong readiness summary: %+v", body)
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
