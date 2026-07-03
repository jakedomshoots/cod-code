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

func Test_Run_prints_failed_report_when_subagent_retries_exhaust(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--subagent-attempts",
		"2",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_always_fails_scanner",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_MODEL_ALWAYS_FAILS_SCANNER", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		SubagentResults []struct {
			Status         string   `json:"status"`
			Attempts       int      `json:"attempts"`
			AttemptErrors  []string `json:"attempt_errors"`
			AttemptRecords []struct {
				Attempt int    `json:"attempt"`
				Status  string `json:"status"`
				Error   string `json:"error"`
			} `json:"attempt_records"`
		} `json:"subagent_results"`
		RunManifest struct {
			Verdict string `json:"verdict"`
		} `json:"run_manifest"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Verdict != "fail" || body.RunManifest.Verdict != "fail" {
		t.Fatalf("verdicts = report %q manifest %q, want fail", body.Verdict, body.RunManifest.Verdict)
	}
	scanner := body.SubagentResults[0]
	if scanner.Status != "fail" || scanner.Attempts != 2 {
		t.Fatalf("scanner result = %#v, want failed two-attempt result", scanner)
	}
	if len(scanner.AttemptErrors) != 2 || !strings.Contains(scanner.AttemptErrors[0], "scanner model unavailable") {
		t.Fatalf("scanner attempt errors = %#v, want two scanner model errors", scanner.AttemptErrors)
	}
	if len(scanner.AttemptRecords) != 2 || scanner.AttemptRecords[1].Status != "fail" {
		t.Fatalf("scanner attempt records = %#v, want two failed records", scanner.AttemptRecords)
	}
}

func Test_HelperProcess_cli_model_always_fails_scanner(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MODEL_ALWAYS_FAILS_SCANNER") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: scanner") {
		os.Stderr.WriteString("scanner model unavailable")
		os.Exit(7)
	}
	os.Stdout.WriteString("ok")
}
