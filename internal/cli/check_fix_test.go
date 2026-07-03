package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_runs_check_fix_attempt_when_enabled(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace", root,
		"--apply-model-patches",
		"--check-fix-attempts", "1",
		"--check", os.Args[0], "-test.run=Test_HelperProcess_cli_check_fix_check", "--",
		"--model-command", os.Args[0], "-test.run=Test_HelperProcess_cli_check_fix_model", "--",
		"Repair", "app",
	}
	t.Setenv("GO_WANT_CLI_CHECK_FIX_CHECK", "1")
	t.Setenv("GO_WANT_CLI_CHECK_FIX_MODEL", "1")
	t.Setenv("GO_CLI_FIX_TARGET", target)

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
	var body struct {
		Verdict      string `json:"verdict"`
		CheckResults []struct {
			Status string `json:"status"`
		} `json:"check_results"`
		PatchAudit []struct {
			Source    string `json:"source"`
			AgentName string `json:"agent_name"`
		} `json:"patch_audit"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if len(body.CheckResults) != 2 || body.CheckResults[0].Status != "fail" || body.CheckResults[1].Status != "pass" {
		t.Fatalf("CheckResults = %+v, want fail then pass", body.CheckResults)
	}
	if len(body.PatchAudit) != 1 || body.PatchAudit[0].Source != "model" || body.PatchAudit[0].AgentName != "coder" {
		t.Fatalf("PatchAudit = %+v, want coder model patch", body.PatchAudit)
	}
}

func Test_HelperProcess_cli_check_fix_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CHECK_FIX_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: coder") && strings.Contains(string(prompt), "Verification failed") {
		os.Stdout.WriteString(`{"patches":[{"path":"app.txt","old":"bad","new":"good"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}

func Test_HelperProcess_cli_check_fix_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CHECK_FIX_CHECK") != "1" {
		return
	}
	content, err := os.ReadFile(os.Getenv("GO_CLI_FIX_TARGET"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if strings.TrimSpace(string(content)) == "good" {
		os.Stdout.WriteString("file fixed\n")
		os.Exit(0)
	}
	os.Stderr.WriteString("file still bad\n")
	os.Exit(4)
}
