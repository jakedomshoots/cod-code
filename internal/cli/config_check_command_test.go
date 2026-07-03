package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_uses_check_command_when_workspace_config_supplies_check_command(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"check_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_config_check_command"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CONFIG_CHECK_COMMAND", "pass")

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
	if len(body.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want 1", len(body.CheckResults))
	}
	if body.CheckResults[0].Status != "pass" {
		t.Fatalf("Check status = %q, want pass", body.CheckResults[0].Status)
	}
}

func Test_HelperProcess_cli_config_check_command(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CONFIG_CHECK_COMMAND") != "pass" {
		return
	}
	os.Stdout.WriteString("config check passed")
}
