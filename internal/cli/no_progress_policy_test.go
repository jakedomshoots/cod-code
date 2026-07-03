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

func Test_Run_stops_repeated_weak_subagent_when_no_progress_stop_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--subagent-attempts",
		"4",
		"--no-progress-stop",
		"2",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_no_progress_model",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_NO_PROGRESS_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want failed verdict\n%s", err, out.String())
	}
	var body struct {
		SubagentResults []struct {
			AgentName         string `json:"agent_name"`
			Attempts          int    `json:"attempts"`
			NoProgressStopped bool   `json:"no_progress_stopped"`
		} `json:"subagent_results"`
		VerificationSummary struct {
			SubagentNoProgressStopCount int `json:"subagent_no_progress_stop_count"`
		} `json:"verification_summary"`
		RunManifest struct {
			NoProgressStop int `json:"no_progress_stop"`
		} `json:"run_manifest"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	scanner := body.SubagentResults[0]
	if scanner.AgentName != "scanner" || scanner.Attempts != 2 || !scanner.NoProgressStopped {
		t.Fatalf("scanner result = %#v, want scanner stopped after two weak attempts", scanner)
	}
	if body.VerificationSummary.SubagentNoProgressStopCount != 1 {
		t.Fatalf("SubagentNoProgressStopCount = %d, want 1", body.VerificationSummary.SubagentNoProgressStopCount)
	}
	if body.RunManifest.NoProgressStop != 2 {
		t.Fatalf("NoProgressStop = %d, want 2", body.RunManifest.NoProgressStop)
	}
}

func Test_HelperProcess_cli_no_progress_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_NO_PROGRESS_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.HasPrefix(string(prompt), "agent: scanner\n") {
		os.Stdout.WriteString(`{"status":"fail","summary":"same weak result"}`)
		os.Exit(0)
		return
	}
	os.Stdout.WriteString(`{"summary":"ok"}`)
	os.Exit(0)
}
