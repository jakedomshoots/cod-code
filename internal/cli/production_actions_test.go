package cli

import (
	"bytes"
	"context"
	"encoding/json"
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
  "required_action_count": 2,
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
      "command": ["sh", "scripts/release-readiness.sh"]
    }
  ]
}`)

	var out bytes.Buffer
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Production actions: blocked",
		"Required actions: 2",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
		"release-readiness [release_proof]: Prove public release readiness",
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
	if body.RequiredActionCount != 2 || len(body.Actions) != 2 || body.Actions[0]["id"] != "provider-openai" {
		t.Fatalf("body = %+v, want two actions starting with provider-openai", body)
	}

	out.Reset()
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--format", "text", "--action-kind", "provider_proof"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text = out.String()
	for _, want := range []string{
		"Required actions: 1",
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
	if err := Run(context.Background(), &out, []string{"production-actions", "--workspace", root, "--action-provider", "openai"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("decode filtered production actions: %v\n%s", err, out.String())
	}
	if body.RequiredActionCount != 1 || len(body.Actions) != 1 || body.Actions[0]["provider"] != "openai" || body.Filter["provider"] != "openai" {
		t.Fatalf("filtered body = %+v, want one openai action", body)
	}
}

func Test_ParseArgs_sets_production_actions_from_verb(t *testing.T) {
	opts, err := parseArgs([]string{"production-actions", "--workspace", "/tmp/workspace", "--action-kind", "provider_proof", "--action-provider", "openai"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if !opts.showProductionActions || opts.workspaceDir != "/tmp/workspace" || opts.productionActionKind != "provider_proof" || opts.productionActionProvider != "openai" {
		t.Fatalf("opts = %+v, want production actions for workspace", opts)
	}
}
