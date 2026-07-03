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

func Test_Run_uses_check_commands_when_workspace_config_supplies_check_commands(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"check_commands":[[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_multi_check_first"` +
		`],[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_multi_check_second"` +
		`]]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_MULTI_CHECK", "pass")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Status string `json:"status"`
			Stdout string `json:"stdout"`
		} `json:"check_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(body.CheckResults))
	}
	if !strings.Contains(body.CheckResults[0].Stdout, "first check passed") {
		t.Fatalf("first stdout = %q, want first check passed", body.CheckResults[0].Stdout)
	}
	if !strings.Contains(body.CheckResults[1].Stdout, "second check passed") {
		t.Fatalf("second stdout = %q, want second check passed", body.CheckResults[1].Stdout)
	}
}

func Test_HelperProcess_cli_multi_check_first(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MULTI_CHECK") != "pass" {
		return
	}
	os.Stdout.WriteString("first check passed")
}

func Test_HelperProcess_cli_multi_check_second(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MULTI_CHECK") != "pass" {
		return
	}
	os.Stdout.WriteString("second check passed")
}
