package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_production_status_reads_latest_readiness_summary(t *testing.T) {
	root := t.TempDir()
	writeProductionStatusSummary(t, filepath.Join(root, ".omo", "evidence", "production-readiness-r1", "summary.json"), `{
  "status": "blocked",
  "local_production_ready": true,
  "public_production_ready": false,
  "blocked_count": 2,
  "blocked_checks": ["release.public_release_ready", "provider.openai_http_provider"],
  "launch_checklist": {
    "path": "launch-checklist.md",
    "sha256": "abc123",
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
	if body.LaunchChecklist == nil || body.LaunchChecklist.RequiredActionCount != 2 || body.LaunchChecklist.SHA256 != "abc123" {
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
