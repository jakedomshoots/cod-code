package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_production_actions_reads_finalizer_action_json(t *testing.T) {
	root := t.TempDir()
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 1,
  "blocked_checks": ["provider.openai_http_provider"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "required_action_count": 1,
    "status": "pass"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "summary.json"), `{
  "status": "blocked",
  "skipped_steps": [],
  "next_actions": {
    "path": "next-actions.md",
    "json_path": "next-actions.json",
    "required_action_count": 1
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "next-actions.json"), `{
  "schema_version": 1,
  "status": "blocked",
  "required_action_count": 5,
  "actions": [
    {
      "id": "provider-openai",
      "kind": "provider_proof",
      "provider": "openai",
      "required_env": "OPENAI_API_KEY",
      "text": "Prove OpenAI HTTP provider",
      "command": ["sh", "scripts/provider-proof.sh", "--provider", "openai"]
    },
    {
      "id": "release-readiness",
      "kind": "release_proof",
      "text": "Prove public release readiness",
      "evidence": "`+filepath.ToSlash(filepath.Join(root, ".omo", "evidence", "release-readiness-final", "index.md"))+`",
      "command": ["sh", "scripts/release-readiness.sh"]
    },
    {
      "id": "competitor-smoke",
      "kind": "competitor_setup",
      "inspect": "competitor-smoke/summary.json",
      "text": "Fix competitor setup"
    },
    {
      "id": "all-agent-29-comparison",
      "kind": "comparison",
      "status": "planned",
      "text": "Run comparison",
      "command": ["ceo-packet", "production-finalize", "--workspace", ".", "--run-comparison"]
    },
    {
      "id": "production-readiness",
      "kind": "final_readiness",
      "status": "blocked",
      "text": "Run final readiness",
      "command": ["sh", "scripts/production-readiness.sh"]
    }
  ]
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "release-readiness-final", "summary.json"), `{
  "status": "blocked",
  "public_release_ready": false,
  "release_artifacts_verified": true,
  "preflight_status": "blocked",
  "blocked_count": 2,
  "blocked_checks": ["git_remote", "github_release_assets"],
  "setup_actions": "setup-actions.md",
  "origin_remote_configured": false,
  "github_auth_status": "pass"
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-finalize-r1", "competitor-smoke", "summary.json"), `{
  "competitors": 3,
  "smoke_passed": 1,
  "smoke_failed": 0,
  "setup_blocked": 1,
  "skipped": 1,
  "setup_actions": "setup-actions.md",
  "results": [
    {"id": "codex_cli", "name": "Codex CLI", "status": "smoke_pass"},
    {"id": "opencode", "name": "OpenCode", "status": "setup_blocked", "note": "provider setup is blocked"},
    {"id": "aider", "name": "Aider", "status": "skipped_missing_binary", "setup_hint": "Install Aider"}
  ]
}`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Production actions: blocked",
		"Required actions: 5",
		"Env ready: 4",
		"Ready now: 2",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
		"(missing env: OPENAI_API_KEY)",
		"Requires env: OPENAI_API_KEY",
		"Command: sh scripts/provider-proof.sh --provider openai",
		"release-readiness [release_proof]: Prove public release readiness",
		"Release readiness: blocked, public_ready=false, artifacts_verified=true, blocked=2",
		"Blocked checks: git_remote, github_release_assets",
		"Setup actions:",
		"release-readiness-final",
		"Command: sh scripts/release-readiness.sh",
		"competitor-smoke [competitor_setup]: Fix competitor setup",
		"Competitor setup: 1 pass, 1 blocked, 1 skipped, 0 failed",
		"opencode: setup_blocked - provider setup is blocked",
		"aider: skipped_missing_binary - Install Aider",
		"Setup actions:",
		"setup-actions.md",
		"all-agent-29-comparison [comparison]: Run comparison",
		"Waiting on: competitor-smoke",
		"production-readiness [final_readiness]: Run final readiness",
		"Waiting on: provider-openai, release-readiness, competitor-smoke, all-agent-29-comparison",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("production actions text missing %q:\n%s", want, text)
		}
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body productionActionsReport
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode production actions: %v\n%s", err, out.String())
	}
	if body.RequiredActionCount != 5 || body.EnvReadyActionCount != 4 || body.ReadyActionCount != 2 || len(body.Actions) != 5 || body.Actions[0]["id"] != "provider-openai" || body.Actions[0]["env_ready"] != false {
		t.Fatalf("body = %+v, want five actions starting with provider-openai", body)
	}
	if body.Actions[0]["action_state"] != "missing_env" || body.Actions[1]["action_state"] != "ready" || body.Actions[3]["action_state"] != "waiting" {
		t.Fatalf("action states = %#v/%#v/%#v, want missing_env/ready/waiting", body.Actions[0]["action_state"], body.Actions[1]["action_state"], body.Actions[3]["action_state"])
	}
	if blockedBy := stringSlice(body.Actions[3]["blocked_by"]); len(blockedBy) != 1 || blockedBy[0] != "competitor-smoke" {
		t.Fatalf("comparison blocked_by = %#v, want competitor-smoke", body.Actions[3]["blocked_by"])
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-kind", "provider_proof"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 1",
		"Env ready: 0",
		"Ready now: 0",
		"Filter: kind=provider_proof",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("filtered production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "release-readiness") {
		t.Fatalf("filtered production actions included release action:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-id", "all-agent-29-comparison"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 1",
		"Filter: id=all-agent-29-comparison",
		"all-agent-29-comparison [comparison]: Run comparison",
		"Waiting on: competitor-smoke",
		"Command: ",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("action-id production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "provider-openai [provider_proof]") {
		t.Fatalf("action-id production actions included provider action:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--action-provider", "openai"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode filtered production actions: %v\n%s", err, out.String())
	}
	if body.RequiredActionCount != 1 || len(body.Actions) != 1 || body.Actions[0]["provider"] != "openai" || body.Filter["provider"] != "openai" {
		t.Fatalf("filtered body = %+v, want one openai action", body)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-state", "missing_env"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 1",
		"Filter: state=missing_env",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("state-filtered production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "release-readiness [release_proof]") {
		t.Fatalf("state-filtered production actions included ready action:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-state", "waiting"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 2",
		"Filter: state=waiting",
		"all-agent-29-comparison [comparison]: Run comparison",
		"production-readiness [final_readiness]: Run final readiness",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("waiting-state production actions text missing %q:\n%s", want, text)
		}
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--env-ready-only"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 4",
		"Env ready: 4",
		"Ready now: 2",
		"Filter: env_ready=true",
		"release-readiness [release_proof]: Prove public release readiness",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("env-ready production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "provider-openai [provider_proof]") {
		t.Fatalf("env-ready production actions included missing-env provider:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--ready-only"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 2",
		"Ready now: 2",
		"Filter: ready=true",
		"release-readiness [release_proof]: Prove public release readiness",
		"competitor-smoke [competitor_setup]: Fix competitor setup",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("ready-only production actions text missing %q:\n%s", want, text)
		}
	}
	for _, notWant := range []string{
		"provider-openai [provider_proof]",
		"all-agent-29-comparison [comparison]",
		"production-readiness [final_readiness]",
	} {
		if strings.Contains(text, notWant) {
			t.Fatalf("ready-only production actions included %q:\n%s", notWant, text)
		}
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--next"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 1",
		"Ready now: 1",
		"Filter: next=true",
		"release-readiness [release_proof]: Prove public release readiness",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("next production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "competitor-smoke [competitor_setup]") {
		t.Fatalf("next production actions included second ready action:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-kind", "provider_proof", "--next"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	if !strings.Contains(text, "Required actions: 0") || strings.Contains(text, "provider-openai [provider_proof]") {
		t.Fatalf("next provider queue should be empty while env is missing:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--action-id", "provider-openai", "--commands-only"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"# provider-openai [provider_proof] missing env: OPENAI_API_KEY state: missing_env",
		"sh scripts/provider-proof.sh --provider openai",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("commands-only production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "secret-value") || strings.Contains(text, "Production actions:") {
		t.Fatalf("commands-only output leaked value or included normal header:\n%s", text)
	}

	t.Setenv("OPENAI_API_KEY", "test-key")
	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-provider", "openai", "--env-ready-only"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	if !strings.Contains(text, "Env ready: 1") || !strings.Contains(text, "provider-openai") || strings.Contains(text, "test-key") {
		t.Fatalf("env-ready production actions text wrong or leaked value:\n%s", text)
	}
}

func Test_ParseArgs_sets_production_actions_from_verb(t *testing.T) {
	opts, err := parseArgs([]string{"production-actions", "--workspace", "/tmp/workspace", "--action-id", "provider-openai", "--action-kind", "provider_proof", "--action-provider", "openai", "--action-state", "missing_env", "--env-ready-only", "--ready-only", "--next", "--commands-only"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if !opts.showProductionActions || opts.workspaceDir != "/tmp/workspace" || opts.productionActionID != "provider-openai" || opts.productionActionKind != "provider_proof" || opts.productionActionProvider != "openai" || opts.productionActionState != "missing_env" || !opts.productionActionsEnvReadyOnly || !opts.productionActionsReadyOnly || !opts.productionActionsNextOnly || !opts.productionActionsCommandsOnly {
		t.Fatalf("opts = %+v, want production actions for workspace", opts)
	}
}

func TestProductionActionsDoesNotReadSecretValuesIntoReport(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "secret-value")
	actions := annotateProductionActions([]map[string]any{{
		"id":           "provider-openai",
		"required_env": "OPENAI_API_KEY",
	}}, "")
	encoded, err := json.Marshal(actions)
	if err != nil {
		t.Fatalf("marshal actions: %v", err)
	}
	if strings.Contains(string(encoded), os.Getenv("OPENAI_API_KEY")) {
		t.Fatalf("annotated action leaked env value: %s", string(encoded))
	}
	if len(actions) != 1 || actions[0]["required_env_set"] != true || actions[0]["env_ready"] != true {
		t.Fatalf("actions = %+v, want env presence only", actions)
	}
}

func TestProductionActionsShellCommandLineQuotesUnsafeArgs(t *testing.T) {
	got := shellCommandLine([]string{"sh", "scripts/run check.sh", "it's", ""})
	want := "sh 'scripts/run check.sh' 'it'\"'\"'s' ''"
	if got != want {
		t.Fatalf("shellCommandLine = %q, want %q", got, want)
	}
}
