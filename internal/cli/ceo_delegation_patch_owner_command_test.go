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

func Test_Run_applies_ceo_created_patch_owner_model_patch(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	config := map[string][]string{
		"ceo_model_command": {os.Args[0], "-test.run=Test_HelperProcess_cli_ceo_patch_owner"},
		"model_command":     {os.Args[0], "-test.run=Test_HelperProcess_cli_ceo_patch_owner"},
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), configJSON, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_PATCH_OWNER", "1")
	args := []string{
		"--workspace",
		root,
		"--write-policy",
		"trusted-local",
		"--apply-model-patches",
		"Patch",
		"checkout",
		"UX",
	}

	// When
	err = Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(content) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(content))
	}
	var body struct {
		PatchAudit []struct {
			Path      string `json:"path"`
			Source    string `json:"source"`
			AgentName string `json:"agent_name"`
		} `json:"patch_audit"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.PatchAudit) != 1 || body.PatchAudit[0].Path != "app.txt" {
		t.Fatalf("PatchAudit = %#v, want one app patch", body.PatchAudit)
	}
	if body.PatchAudit[0].Source != "model" || body.PatchAudit[0].AgentName != "ux_coder" {
		t.Fatalf("PatchAudit[0] = %#v, want ux_coder model source", body.PatchAudit[0])
	}
}

func Test_HelperProcess_cli_ceo_patch_owner(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_PATCH_OWNER") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["ux_coder"],"new_subagents":[{"name":"ux_coder","role":"patch checkout UX","stage":2,"allowed_actions":["read_workspace","propose_patch"]}],"summary":"UX patch needs one owner."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "agent: ux_coder") {
		os.Stdout.WriteString(`{"patches":[{"path":"app.txt","old":"old","new":"new"}]}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"Patch owner passed."}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}
