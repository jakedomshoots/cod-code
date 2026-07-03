package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_uses_auto_check_set_when_task_matches_keyword(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := fmt.Sprintf(
		`{"default_check_set":"quick","check_sets":{"quick":[[%q,"-test.run=Test_HelperProcess_cli_check_set_quick"]],"full":[[%q,"-test.run=Test_HelperProcess_cli_check_set_full"],[%q,"-test.run=Test_HelperProcess_cli_check_set_review"]]},"auto_check_sets":[{"check_set":"full","keywords":["auth","security"]}]}`,
		os.Args[0],
		os.Args[0],
		os.Args[0],
	)
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CHECK_SET", "pass")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "auth", "callback"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Stdout string `json:"stdout"`
		} `json:"check_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want 2", len(body.CheckResults))
	}
	if !strings.Contains(body.CheckResults[0].Stdout, "full check passed") {
		t.Fatalf("first stdout = %q, want full check passed", body.CheckResults[0].Stdout)
	}
	if !strings.Contains(body.CheckResults[1].Stdout, "review check passed") {
		t.Fatalf("second stdout = %q, want review check passed", body.CheckResults[1].Stdout)
	}
}

func Test_Run_uses_check_set_flag_over_auto_check_set_when_both_match(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := fmt.Sprintf(
		`{"default_check_set":"quick","check_sets":{"quick":[[%q,"-test.run=Test_HelperProcess_cli_check_set_quick"]],"full":[[%q,"-test.run=Test_HelperProcess_cli_check_set_full"]]},"auto_check_sets":[{"check_set":"full","keywords":["auth"]}]}`,
		os.Args[0],
		os.Args[0],
	)
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CHECK_SET", "pass")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--check-set", "quick", "Fix", "auth", "callback"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Stdout string `json:"stdout"`
		} `json:"check_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.CheckResults) != 1 {
		t.Fatalf("CheckResults length = %d, want 1", len(body.CheckResults))
	}
	if !strings.Contains(body.CheckResults[0].Stdout, "quick check passed") {
		t.Fatalf("stdout = %q, want quick check passed", body.CheckResults[0].Stdout)
	}
}
