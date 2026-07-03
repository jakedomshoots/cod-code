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

func Test_Run_applies_structured_coder_patch_and_reports_summary(t *testing.T) {
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
		"-test.run=Test_HelperProcess_cli_structured_model_output",
		"--",
		"Patch",
		"from",
		"coder",
	}
	t.Setenv("GO_WANT_CLI_STRUCTURED_MODEL_OUTPUT", "1")

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
		SubagentResults []struct {
			AgentName      string   `json:"agent_name"`
			Summary        string   `json:"summary"`
			Evidence       []string `json:"evidence"`
			PatchProposals []struct {
				Path string `json:"path"`
			} `json:"patches"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	coder := body.SubagentResults[1]
	if coder.AgentName != "coder" || coder.Summary != "structured patch ready" {
		t.Fatalf("coder result = %+v, want structured summary", coder)
	}
	if len(coder.Evidence) != 1 || coder.Evidence[0] != "old text found" {
		t.Fatalf("coder evidence = %+v, want structured evidence", coder.Evidence)
	}
	if len(coder.PatchProposals) != 1 || coder.PatchProposals[0].Path != "app.txt" {
		t.Fatalf("coder patches = %+v, want structured patch proposal", coder.PatchProposals)
	}
}

func Test_HelperProcess_cli_structured_model_output(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_STRUCTURED_MODEL_OUTPUT") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: coder") {
		os.Stdout.WriteString(`{"summary":"structured patch ready","evidence":["old text found"],"patches":[{"path":"app.txt","old":"old","new":"new"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}
