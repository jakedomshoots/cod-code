package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_config_check_reports_adapter_capabilities_and_missing_setup(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	fakeCodex := writeConfigCheckFakeAdapter(t, t.TempDir())
	t.Setenv("CEO_CODEX_ADAPTER_COMMAND", fakeCodex)
	for _, envName := range []string{"CEO_CLAUDE_ADAPTER_COMMAND", "CEO_OPENCODE_ADAPTER_COMMAND", "CEO_AIDER_ADAPTER_COMMAND", "CEO_GOOSE_ADAPTER_COMMAND"} {
		t.Setenv(envName, "")
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		AdapterCapabilities []struct {
			Tool       string   `json:"tool"`
			Status     string   `json:"status"`
			ErrorKind  string   `json:"error_kind,omitempty"`
			PatchCount int      `json:"patch_count"`
			SetupSteps []string `json:"setup_steps,omitempty"`
		} `json:"adapter_capabilities"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.AdapterCapabilities) != 5 {
		t.Fatalf("adapter_capabilities length = %d, want five", len(body.AdapterCapabilities))
	}
	var codexSeen bool
	var missingSeen bool
	for _, report := range body.AdapterCapabilities {
		if report.Tool == "codex" {
			codexSeen = true
			if report.Status != "pass" || report.PatchCount != 1 {
				t.Fatalf("codex report = %+v, want passing dry-run patch parse", report)
			}
		}
		if report.Tool == "goose" && report.Status == "skip" && report.ErrorKind == "missing_setup" && len(report.SetupSteps) > 0 {
			missingSeen = true
		}
	}
	if !codexSeen || !missingSeen {
		t.Fatalf("adapter reports = %+v, want codex pass and goose missing setup", body.AdapterCapabilities)
	}
}

func writeConfigCheckFakeAdapter(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "codex-adapter")
	body := `#!/bin/sh
if [ "$CEO_HARNESS_ADAPTER_PROBE" = "version" ]; then
  echo "codex version 9.9.9"
  exit 0
fi
if [ "$CEO_HARNESS_ADAPTER_PROBE" = "dry-run" ]; then
  cat >/dev/null
  echo '{"status":"pass","summary":"codex patch ready","patches":[{"path":"app.txt","old":"old","new":"new"}]}'
  exit 0
fi
exit 2
`
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake adapter: %v", err)
	}
	return path
}
