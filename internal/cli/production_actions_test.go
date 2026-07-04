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
  "required_action_count": 1,
  "actions": [
    {
      "id": "provider-openai",
      "kind": "provider_proof",
      "provider": "openai",
      "required_env": "OPENAI_API_KEY",
      "text": "Prove OpenAI HTTP provider",
      "command": ["sh", "scripts/provider-proof.sh", "--provider", "openai"]
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
		"Required actions: 1",
		"provider-openai [provider_proof]: Prove OpenAI HTTP provider",
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
	if body.RequiredActionCount != 1 || len(body.Actions) != 1 || body.Actions[0]["id"] != "provider-openai" {
		t.Fatalf("body = %+v, want one provider-openai action", body)
	}
}

func Test_ParseArgs_sets_production_actions_from_verb(t *testing.T) {
	opts, err := parseArgs([]string{"production-actions", "--workspace", "/tmp/workspace"})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if !opts.showProductionActions || opts.workspaceDir != "/tmp/workspace" {
		t.Fatalf("opts = %+v, want production actions for workspace", opts)
	}
}
