package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_Run_prints_report_and_returns_needs_input_error_when_model_needs_input(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_needs_input_model",
		"--",
		"Fix",
		"ambiguous",
		"package",
	}
	t.Setenv("GO_WANT_CLI_NEEDS_INPUT_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictNeedsInput) {
		t.Fatalf("Run error = %v, want ErrVerdictNeedsInput", err)
	}
	var body struct {
		Verdict         string `json:"verdict"`
		SubagentResults []struct {
			Status    string   `json:"status"`
			Questions []string `json:"questions"`
		} `json:"subagent_results"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "needs_input" {
		t.Fatalf("Verdict = %q, want needs_input", body.Verdict)
	}
	if len(body.SubagentResults[0].Questions) != 1 {
		t.Fatalf("Questions = %+v, want one question", body.SubagentResults[0].Questions)
	}
}

func Test_HelperProcess_cli_needs_input_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_NEEDS_INPUT_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: scanner") {
		os.Stdout.WriteString(`{"status":"needs_input","summary":"missing target repo","questions":["Which package should I change?"]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}
