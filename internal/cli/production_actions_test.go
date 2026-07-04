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
      "evidence": "`+filepath.ToSlash(filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "index.md"))+`",
      "declared_evidence_files": [
        {
          "field": "evidence",
          "path": "`+filepath.ToSlash(filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "index.md"))+`",
          "exists": true,
          "size_bytes": 27,
          "sha256": "b4da9bebb1ea70a3c232161b674b70a15ed3d1ef59fd23dbf94f8dae8c7fda42"
        }
      ],
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
  "setup_action_count": 2,
  "setup_actions_sha256": "1111111111111111111111111111111111111111111111111111111111111111",
  "setup_command_policy": "no_publish_no_secret_assignment",
  "publish_actions_performed": false,
  "secret_value_saved": false,
  "origin_remote_configured": false,
  "github_auth_status": "pass"
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "release-readiness-final", "index.md"), `# Release Readiness Evidence
`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "release-readiness-final", "setup-actions.md"), `# Release Setup Actions

- git_remote: configure an origin remote for the public repo.
- github_release_assets: push a v* tag and upload release assets.
`)
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
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "summary.json"), `{
  "schema_version": 1,
  "status": "blocked",
  "provider": "openai",
  "provider_mode": "http-provider",
  "http_preset": "openai",
  "http_model": "gpt-5",
  "api_key_env": "OPENAI_API_KEY",
  "blocked_reason": "missing_api_key_env",
  "secret_value_saved": false,
  "command_script_secret_policy": "no_secret_assignment",
  "setup_checklist_item_count": 3,
  "setup_artifacts_sha256": {
    "blocked.md": "2222222222222222222222222222222222222222222222222222222222222222",
    "commands.sh": "3333333333333333333333333333333333333333333333333333333333333333",
    "setup-checklist.md": "4444444444444444444444444444444444444444444444444444444444444444"
  },
  "artifacts": {
    "checklist": "setup-checklist.md",
    "commands": "commands.sh",
    "env_template": "env.template"
  }
}`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "index.md"), `# Provider Proof Evidence
`)
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "provider-proof-openai", "setup-checklist.md"), `# Provider Setup Checklist

1. Export `+"`OPENAI_API_KEY`"+` in the shell or local secret manager.
2. Keep the key out of git, logs, reports, and evidence folders.
3. Run `+"`commands.sh`"+` from the repo root.
`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Production actions: blocked",
		"Required actions: 5",
		"Env ready: 4",
		"Ready now: 0",
		"Runnable commands: 0",
		"Blocked commands: 4",
		"Evidence matches: declared=1 mismatched=0",
		"States: missing_env=1 setup_blocked=2 waiting=2",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
		"(missing env: OPENAI_API_KEY)",
		"Reason: missing required env: OPENAI_API_KEY",
		"Provider blocker: missing_api_key_env",
		"Provider model: gpt-5",
		"Command secret policy: no_secret_assignment",
		"Setup checklist:",
		"Setup checklist count: 3",
		"provider-proof-openai",
		"Setup checklist items:",
		"1. Export `OPENAI_API_KEY` in the shell or local secret manager.",
		"2. Keep the key out of git, logs, reports, and evidence folders.",
		"Setup artifact hashes: blocked.md=2222222222222222222222222222222222222222222222222222222222222222 commands.sh=3333333333333333333333333333333333333333333333333333333333333333 setup-checklist.md=4444444444444444444444444444444444444444444444444444444444444444",
		"Setup command file:",
		"Evidence file:",
		"sha256=",
		"declared_match=true",
		"Requires env: OPENAI_API_KEY",
		"Command: sh scripts/provider-proof.sh --provider openai",
		"release-readiness [release_proof]: Prove public release readiness",
		"Reason: release setup blocked: git_remote, github_release_assets",
		"Release readiness: blocked, public_ready=false, artifacts_verified=true, blocked=2",
		"Blocked checks: git_remote, github_release_assets",
		"Setup actions:",
		"release-readiness-final",
		"Setup action count: 2",
		"Setup actions sha256: 1111111111111111111111111111111111111111111111111111111111111111",
		"Setup command policy: no_publish_no_secret_assignment",
		"Publish actions performed: false",
		"Secret value saved: false",
		"Setup action items:",
		"git_remote: configure an origin remote for the public repo.",
		"github_release_assets: push a v* tag and upload release assets.",
		"Command: sh scripts/release-readiness.sh",
		"competitor-smoke [competitor_setup]: Fix competitor setup",
		"Reason: competitor setup blocked: opencode, aider",
		"Competitor setup: 1 pass, 1 blocked, 1 skipped, 0 failed",
		"opencode: setup_blocked - provider setup is blocked",
		"aider: skipped_missing_binary - Install Aider",
		"Setup actions:",
		"setup-actions.md",
		"all-agent-29-comparison [comparison]: Run comparison",
		"Reason: waiting on: competitor-smoke",
		"Waiting on: competitor-smoke",
		"production-readiness [final_readiness]: Run final readiness",
		"Reason: waiting on: provider-openai, release-readiness, competitor-smoke, all-agent-29-comparison",
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
	if body.RequiredActionCount != 5 || body.EnvReadyActionCount != 4 || body.ReadyActionCount != 0 || body.RunnableCommandCount != 0 || body.BlockedCommandCount != 4 || body.EvidenceDeclaredMatchCount != 1 || body.EvidenceDeclaredMismatchCount != 0 || len(body.Actions) != 5 || body.Actions[0]["id"] != "provider-openai" || body.Actions[0]["env_ready"] != false {
		t.Fatalf("body = %+v, want five actions starting with provider-openai", body)
	}
	if body.ActionStateCounts["missing_env"] != 1 || body.ActionStateCounts["setup_blocked"] != 2 || body.ActionStateCounts["waiting"] != 2 {
		t.Fatalf("ActionStateCounts = %#v, want missing/setup/waiting counts", body.ActionStateCounts)
	}
	if body.Actions[0]["action_state"] != "missing_env" || body.Actions[1]["action_state"] != "setup_blocked" || body.Actions[2]["action_state"] != "setup_blocked" || body.Actions[3]["action_state"] != "waiting" {
		t.Fatalf("action states = %#v/%#v/%#v/%#v, want missing_env/setup_blocked/setup_blocked/waiting", body.Actions[0]["action_state"], body.Actions[1]["action_state"], body.Actions[2]["action_state"], body.Actions[3]["action_state"])
	}
	if body.Actions[0]["action_reason"] != "missing required env: OPENAI_API_KEY" || body.Actions[1]["action_reason"] != "release setup blocked: git_remote, github_release_assets" || body.Actions[2]["action_reason"] != "competitor setup blocked: opencode, aider" || body.Actions[3]["action_reason"] != "waiting on: competitor-smoke" {
		t.Fatalf("action reasons = %#v/%#v/%#v/%#v, want specific blocker reasons", body.Actions[0]["action_reason"], body.Actions[1]["action_reason"], body.Actions[2]["action_reason"], body.Actions[3]["action_reason"])
	}
	providerSummary, _ := body.Actions[0]["provider_summary"].(map[string]any)
	checklistItems, _ := providerSummary["checklist_items"].([]any)
	if len(checklistItems) != 3 {
		t.Fatalf("checklist_items = %#v, want three structured provider checklist items", providerSummary["checklist_items"])
	}
	if numberValue(providerSummary["setup_checklist_item_count"]) != 3 {
		t.Fatalf("setup_checklist_item_count = %#v, want 3", providerSummary["setup_checklist_item_count"])
	}
	if hashes := stringStringMap(providerSummary["setup_artifacts_sha256"]); hashes["commands.sh"] != "3333333333333333333333333333333333333333333333333333333333333333" {
		t.Fatalf("setup_artifacts_sha256 = %#v, want commands hash", providerSummary["setup_artifacts_sha256"])
	}
	if providerSummary["command_script_secret_policy"] != "no_secret_assignment" {
		t.Fatalf("command_script_secret_policy = %#v, want no_secret_assignment", providerSummary["command_script_secret_policy"])
	}
	releaseSummary, _ := body.Actions[1]["release_summary"].(map[string]any)
	setupItems, _ := releaseSummary["setup_action_items"].([]any)
	if len(setupItems) != 2 {
		t.Fatalf("setup_action_items = %#v, want two structured release setup items", releaseSummary["setup_action_items"])
	}
	if numberValue(releaseSummary["setup_action_count"]) != 2 || releaseSummary["setup_actions_sha256"] != "1111111111111111111111111111111111111111111111111111111111111111" {
		t.Fatalf("release setup proof = %#v, want count and sha", releaseSummary)
	}
	if releaseSummary["setup_command_policy"] != "no_publish_no_secret_assignment" || releaseSummary["publish_actions_performed"] != false || releaseSummary["secret_value_saved"] != false {
		t.Fatalf("release safety proof = %#v, want no publish/no secret policy", releaseSummary)
	}
	if blockedBy := stringSlice(body.Actions[3]["blocked_by"]); len(blockedBy) != 1 || blockedBy[0] != "competitor-smoke" {
		t.Fatalf("comparison blocked_by = %#v, want competitor-smoke", body.Actions[3]["blocked_by"])
	}
	evidenceFiles, _ := body.Actions[0]["evidence_files"].([]any)
	if len(evidenceFiles) != 1 {
		t.Fatalf("provider evidence_files = %#v, want one declared evidence file", body.Actions[0]["evidence_files"])
	}
	firstEvidence, _ := evidenceFiles[0].(map[string]any)
	if firstEvidence["exists"] != true || firstEvidence["sha256"] == "" || numberValue(firstEvidence["size_bytes"]) <= 0 || firstEvidence["matches_declared"] != true {
		t.Fatalf("provider evidence metadata = %#v, want existing fingerprinted evidence", firstEvidence)
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
		"Runnable commands: 0",
		"Blocked commands: 1",
		"Evidence matches: declared=1 mismatched=0",
		"States: missing_env=1",
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
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-state", "setup_blocked"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 2",
		"Filter: state=setup_blocked",
		"States: setup_blocked=2",
		"release-readiness [release_proof]: Prove public release readiness",
		"competitor-smoke [competitor_setup]: Fix competitor setup",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("setup-blocked production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "provider-openai [provider_proof]") {
		t.Fatalf("setup-blocked production actions included provider missing-env action:\n%s", text)
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
		"Ready now: 0",
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
		"Required actions: 0",
		"Ready now: 0",
		"Filter: ready=true",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("ready-only production actions text missing %q:\n%s", want, text)
		}
	}
	for _, notWant := range []string{
		"provider-openai [provider_proof]",
		"release-readiness [release_proof]",
		"competitor-smoke [competitor_setup]",
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
		"Required actions: 0",
		"Ready now: 0",
		"Filter: next=true",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("next production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "release-readiness [release_proof]") || strings.Contains(text, "competitor-smoke [competitor_setup]") {
		t.Fatalf("next production actions included setup-blocked action:\n%s", text)
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
		"# provider-openai [provider_proof] missing env: OPENAI_API_KEY state: missing_env reason: missing required env: OPENAI_API_KEY",
		"# setup checklist:",
		"# 1. Export `OPENAI_API_KEY` in the shell or local secret manager.",
		"# 2. Keep the key out of git, logs, reports, and evidence folders.",
		"# blocked command: sh scripts/provider-proof.sh --provider openai",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("commands-only production actions text missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "secret-value") || strings.Contains(text, "Production actions:") {
		t.Fatalf("commands-only output leaked value or included normal header:\n%s", text)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--action-id", "release-readiness", "--commands-only"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"# release-readiness [release_proof] state: setup_blocked reason: release setup blocked: git_remote, github_release_assets",
		"# setup actions:",
		"# - git_remote: configure an origin remote for the public repo.",
		"# - github_release_assets: push a v* tag and upload release assets.",
		"# blocked command: sh scripts/release-readiness.sh",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("release commands-only text missing %q:\n%s", want, text)
		}
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

func Test_ParseArgs_validates_production_action_state(t *testing.T) {
	for _, state := range []string{"ready", "missing_env", "empty_env", "setup_blocked", "waiting"} {
		opts, err := parseArgs([]string{"production-actions", "--workspace", "/tmp/workspace", "--action-state", state})
		if err != nil {
			t.Fatalf("parseArgs(%s): %v", state, err)
		}
		if opts.productionActionState != state {
			t.Fatalf("productionActionState = %q, want %q", opts.productionActionState, state)
		}
	}
	if _, err := parseArgs([]string{"production-actions", "--workspace", "/tmp/workspace", "--action-state", "blocked"}); err == nil || !strings.Contains(err.Error(), "--action-state must be ready, missing_env, empty_env, setup_blocked, or waiting") {
		t.Fatalf("parseArgs invalid state err = %v, want validation error", err)
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

func TestProductionActionsTreatsEmptyEnvAsNotReady(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	actions := annotateProductionActions([]map[string]any{{
		"id":           "provider-openrouter",
		"kind":         "provider_proof",
		"required_env": "OPENROUTER_API_KEY",
		"command":      []any{"sh", "scripts/provider-proof.sh", "--provider", "openrouter"},
	}}, "")
	if len(actions) != 1 {
		t.Fatalf("actions length = %d, want 1", len(actions))
	}
	if actions[0]["required_env_set"] != true || actions[0]["env_ready"] != false || actions[0]["empty_required_env"] != "OPENROUTER_API_KEY" || actions[0]["action_state"] != "empty_env" {
		t.Fatalf("actions = %+v, want empty env blocker", actions)
	}
	report := productionActionsReport{Actions: actions, CommandsOnly: true}
	text := renderProductionActionCommandsOnly(report)
	if !strings.Contains(text, "# provider-openrouter [provider_proof] empty env: OPENROUTER_API_KEY state: empty_env reason: required env is set but empty: OPENROUTER_API_KEY") {
		t.Fatalf("commands-only output missing empty env state:\n%s", text)
	}
	if strings.Contains(text, "missing env: OPENROUTER_API_KEY") {
		t.Fatalf("commands-only output should distinguish empty env from missing env:\n%s", text)
	}
	if !strings.Contains(text, "# blocked command: sh scripts/provider-proof.sh --provider openrouter") {
		t.Fatalf("commands-only output should comment blocked command:\n%s", text)
	}
}

func TestProductionActionsShellCommandLineQuotesUnsafeArgs(t *testing.T) {
	got := shellCommandLine([]string{"sh", "scripts/run check.sh", "it's", ""})
	want := "sh 'scripts/run check.sh' 'it'\"'\"'s' ''"
	if got != want {
		t.Fatalf("shellCommandLine = %q, want %q", got, want)
	}
}
