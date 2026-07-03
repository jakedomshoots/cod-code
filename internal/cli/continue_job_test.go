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

func Test_Run_continues_saved_job_without_rerunning_passed_subagents(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_CONTINUE_JOB_MODEL", "1")

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_continue_job_model",
		"--",
		"--continue-job",
		"job-000001",
	})

	// Then
	if err != nil {
		t.Fatalf("continue Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobID        string `json:"job_id"`
		Continuation struct {
			JobID               string `json:"job_id"`
			ReusedSubagentCount int    `json:"reused_subagent_count"`
		} `json:"continuation"`
		RunManifest struct {
			ReusedSubagentCount int `json:"reused_subagent_count"`
		} `json:"run_manifest"`
		SubagentResults []struct {
			Reused bool `json:"reused"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.JobID != "job-000002" {
		t.Fatalf("JobID = %q, want job-000002", body.JobID)
	}
	wantReused := len(body.SubagentResults)
	if body.Continuation.JobID != "job-000001" || body.Continuation.ReusedSubagentCount != wantReused {
		t.Fatalf("Continuation = %+v, want source job and %d reused", body.Continuation, wantReused)
	}
	if body.RunManifest.ReusedSubagentCount != wantReused {
		t.Fatalf("ReusedSubagentCount = %d, want %d", body.RunManifest.ReusedSubagentCount, wantReused)
	}
	for index, result := range body.SubagentResults {
		if !result.Reused {
			t.Fatalf("SubagentResults[%d].Reused = false, want true", index)
		}
	}
}

func Test_Run_rejects_continue_job_with_task_text(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--continue-job", "job-000001", "New", "task"})

	// Then
	if err == nil {
		t.Fatal("expected continue/task conflict error")
	}
}

func Test_Run_records_reused_subagent_count_in_history_when_continuing(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "--continue-job", "job-000001"}); err != nil {
		t.Fatalf("continue Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--history"})

	// Then
	if err != nil {
		t.Fatalf("history Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		History []struct {
			ID                      string `json:"id"`
			ReusedSubagentCount     int    `json:"reused_subagent_count"`
			ExecutionPlanNextAction string `json:"execution_plan_next_action"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 2 {
		t.Fatalf("history length = %d, want 2", len(body.History))
	}
	if body.History[1].ID != "job-000002" || body.History[1].ReusedSubagentCount != 3 {
		t.Fatalf("continued history row = %#v, want job-000002 with 3 reused subagents", body.History[1])
	}
}

func Test_Run_continues_saved_job_without_reasking_ceo_delegation(t *testing.T) {
	// Given
	root := t.TempDir()
	configJSON := `{"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_continue_job_ceo_model"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CONTINUE_JOB_CEO_MODEL", "seed")
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_CONTINUE_JOB_CEO_MODEL", "continue")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--continue-job", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("continue Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Continuation struct {
			JobID               string `json:"job_id"`
			ReusedSubagentCount int    `json:"reused_subagent_count"`
		} `json:"continuation"`
		JobPacket struct {
			Subagents []struct {
				Name string `json:"name"`
			} `json:"subagents"`
		} `json:"job_packet"`
		CEOReview struct {
			Summary string `json:"summary"`
		} `json:"ceo_review"`
		CEODelegation struct {
			Source            string   `json:"source"`
			SelectedSubagents []string `json:"selected_subagents"`
		} `json:"ceo_delegation"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Continuation.JobID != "job-000001" || body.Continuation.ReusedSubagentCount != 1 {
		t.Fatalf("Continuation = %+v, want one reused saved subagent", body.Continuation)
	}
	if len(body.JobPacket.Subagents) != 1 || body.JobPacket.Subagents[0].Name != "coder" {
		t.Fatalf("JobPacket subagents = %#v, want saved delegated coder only", body.JobPacket.Subagents)
	}
	if body.CEOReview.Summary != "continued saved delegation passed" {
		t.Fatalf("CEO review summary = %q, want continued saved delegation passed", body.CEOReview.Summary)
	}
	if body.CEODelegation.Source != "history" ||
		len(body.CEODelegation.SelectedSubagents) != 1 ||
		body.CEODelegation.SelectedSubagents[0] != "coder" {
		t.Fatalf("CEODelegation = %#v, want history-sourced saved coder delegation", body.CEODelegation)
	}
}

func Test_HelperProcess_cli_continue_job_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CONTINUE_JOB_MODEL") != "1" {
		return
	}
	t.Fatal("continued job should not call model for reused subagents")
}

func Test_HelperProcess_cli_continue_job_ceo_model(t *testing.T) {
	mode := os.Getenv("GO_WANT_CLI_CONTINUE_JOB_CEO_MODEL")
	if mode == "" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		if mode == "continue" {
			os.Stderr.WriteString("continue-job reasked CEO delegation")
			os.Exit(17)
		}
		os.Stdout.WriteString(`{"selected_subagents":["coder"],"summary":"Use saved coder lane."}`)
		os.Exit(0)
	}
	if strings.Contains(text, "guard_verdict: pass") {
		os.Stdout.WriteString(`{"recommended_verdict":"pass","summary":"continued saved delegation passed"}`)
		os.Exit(0)
	}
	os.Stderr.WriteString("unexpected CEO prompt")
	os.Exit(18)
}
