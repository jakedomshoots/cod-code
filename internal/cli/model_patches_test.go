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

func Test_Run_applies_coder_model_patch_when_enabled(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace",
		root,
		"--write-policy",
		"trusted-local",
		"--apply-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_patch",
		"--",
		"Patch",
		"from",
		"coder",
	}
	t.Setenv("GO_WANT_CLI_MODEL_PATCH", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	var body struct {
		PatchResults []struct {
			Path string `json:"path"`
		} `json:"patch_results"`
		PatchAudit []struct {
			Path      string `json:"path"`
			Source    string `json:"source"`
			AgentName string `json:"agent_name"`
		} `json:"patch_audit"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.PatchResults) != 1 || body.PatchResults[0].Path != "app.txt" {
		t.Fatalf("PatchResults = %+v, want app.txt patch", body.PatchResults)
	}
	if len(body.PatchAudit) != 1 {
		t.Fatalf("PatchAudit length = %d, want 1", len(body.PatchAudit))
	}
	if body.PatchAudit[0].Source != "model" || body.PatchAudit[0].AgentName != "coder" {
		t.Fatalf("PatchAudit[0] = %+v, want coder model source", body.PatchAudit[0])
	}
}

func Test_Run_applies_coder_model_create_file_patch_when_enabled(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--write-policy",
		"trusted-local",
		"--apply-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_create_file_patch",
		"--",
		"Create",
		"notes",
	}
	t.Setenv("GO_WANT_CLI_MODEL_CREATE_FILE_PATCH", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "docs", "notes.md"))
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	if string(got) != "# Notes\n" {
		t.Fatalf("content = %q, want created notes", string(got))
	}
	var body struct {
		PatchResults []struct {
			Path string `json:"path"`
			Diff string `json:"diff"`
		} `json:"patch_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.PatchResults) != 1 || body.PatchResults[0].Path != "docs/notes.md" || body.PatchResults[0].Diff == "" {
		t.Fatalf("PatchResults = %+v, want create file diff", body.PatchResults)
	}
}

func Test_Run_skips_coder_model_patch_when_coder_lacks_patch_action(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	config := `{"subagents":[{"name":"coder","role":"read only coding review","allowed_actions":["read_workspace"]}]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	args := []string{
		"--workspace",
		root,
		"--apply-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_patch",
		"--",
		"Patch",
		"from",
		"coder",
	}
	t.Setenv("GO_WANT_CLI_MODEL_PATCH", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged file", string(got))
	}
	var body struct {
		PatchResults []struct {
			Path string `json:"path"`
		} `json:"patch_results"`
		SubagentResults []struct {
			AllowedActions []string `json:"allowed_actions"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.PatchResults) != 0 {
		t.Fatalf("PatchResults = %+v, want no patch without permission", body.PatchResults)
	}
	if len(body.SubagentResults) != 1 || body.SubagentResults[0].AllowedActions[0] != "read_workspace" {
		t.Fatalf("SubagentResults = %+v, want read-only coder action", body.SubagentResults)
	}
}

func Test_HelperProcess_cli_model_patch(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MODEL_PATCH") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: coder") {
		os.Stdout.WriteString(`{"patches":[{"path":"app.txt","old":"old","new":"new"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}

func Test_HelperProcess_cli_model_create_file_patch(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MODEL_CREATE_FILE_PATCH") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: coder") {
		os.Stdout.WriteString(`{"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}
