package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_prints_ceo_job_packet_when_task_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobPacket struct {
			Task          string `json:"task"`
			MaxSubagents  int    `json:"max_subagents"`
			ContextPolicy struct {
				Mode string `json:"mode"`
			} `json:"context_policy"`
		} `json:"job_packet"`
		VerificationContract struct {
			Status             string `json:"status"`
			RequiredCheckCount int    `json:"required_check_count"`
		} `json:"verification_contract"`
		RunLedger struct {
			VerificationStatus string `json:"verification_status"`
		} `json:"run_ledger"`
		SubagentResults []struct {
			AgentName       string   `json:"agent_name"`
			Status          string   `json:"status"`
			ContextReceived string   `json:"context_received"`
			Evidence        []string `json:"evidence"`
		} `json:"subagent_results"`
		Verdict string `json:"verdict"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobPacket.Task != "Fix a failing test" {
		t.Fatalf("Task = %q, want joined args", body.JobPacket.Task)
	}
	if body.JobPacket.MaxSubagents != 3 {
		t.Fatalf("MaxSubagents = %d, want 3", body.JobPacket.MaxSubagents)
	}
	if body.JobPacket.ContextPolicy.Mode != "lean" {
		t.Fatalf("context mode = %q, want lean", body.JobPacket.ContextPolicy.Mode)
	}
	if body.VerificationContract.Status != "unverified" || body.VerificationContract.RequiredCheckCount != 0 {
		t.Fatalf("verification contract = %#v, want unverified with no required checks", body.VerificationContract)
	}
	if body.RunLedger.VerificationStatus != "unverified" {
		t.Fatalf("RunLedger.VerificationStatus = %q, want unverified", body.RunLedger.VerificationStatus)
	}
	if len(body.SubagentResults) != 3 {
		t.Fatalf("SubagentResults length = %d, want 3", len(body.SubagentResults))
	}
	if body.SubagentResults[0].ContextReceived != "lean" {
		t.Fatalf("ContextReceived = %q, want lean", body.SubagentResults[0].ContextReceived)
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
}

func Test_Run_includes_check_result_when_check_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--check", "go", "version", "--", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		CheckResults []struct {
			Status string `json:"status"`
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

func Test_Run_returns_error_after_report_when_check_fails(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--check", os.Args[0], "-test.run=Test_HelperProcess_cli_fail_check", "--", "Fix", "a", "failing", "test"}
	t.Setenv("GO_WANT_CLI_HELPER_PROCESS", "fail")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil {
		t.Fatal("expected failed verdict error")
	}
	var body struct {
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", body.Verdict)
	}
}

func Test_Run_applies_text_patch_when_replace_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{"--workspace", root, "--write-policy", "trusted-local", "--replace", "app.txt", "old", "new", "Patch", "app", "text"}

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
		RunLedger struct {
			ChangedFileCount int      `json:"changed_file_count"`
			ChangedFiles     []string `json:"changed_files"`
		} `json:"run_ledger"`
		PatchResults []struct {
			Path string `json:"path"`
			Diff string `json:"diff"`
		} `json:"patch_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.PatchResults) != 1 {
		t.Fatalf("PatchResults length = %d, want 1", len(body.PatchResults))
	}
	if body.PatchResults[0].Diff == "" {
		t.Fatal("expected patch diff")
	}
	if body.RunLedger.ChangedFileCount != len(body.RunLedger.ChangedFiles) {
		t.Fatalf("RunLedger changed count = %d files %+v, want matching count", body.RunLedger.ChangedFileCount, body.RunLedger.ChangedFiles)
	}
	if !containsString(body.RunLedger.ChangedFiles, "app.txt") {
		t.Fatalf("RunLedger.ChangedFiles = %+v, want app.txt", body.RunLedger.ChangedFiles)
	}
}

func Test_Run_accepts_separator_before_task_text_after_replace(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{"--workspace", root, "--write-policy", "trusted-local", "--replace", "app.txt", "old", "new", "--", "Patch demo app"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
}

func Test_HelperProcess_cli_fail_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_HELPER_PROCESS") != "fail" {
		return
	}
	os.Stderr.WriteString("cli check failed\n")
	os.Exit(8)
}

func Test_Run_opens_cod_chat_when_no_args_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, nil)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if body := out.String(); !strings.Contains(body, "╭─ Cod Code") || !strings.Contains(body, "Composer") {
		t.Fatalf("cod chat output missing expected TUI markers:\n%s", body)
	}
}

func Test_Run_writes_evidence_file_when_workspace_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{"--workspace", root, "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	wantFiles := []string{"scanner.md", "coder.md", "reviewer.md"}
	for _, wantFile := range wantFiles {
		got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", wantFile))
		if err != nil {
			t.Fatalf("read evidence file %s: %v", wantFile, err)
		}
		agentName := strings.TrimSuffix(wantFile, ".md")
		if !strings.Contains(string(got), agentName) {
			t.Fatalf("evidence file %s = %q, want agent evidence", wantFile, string(got))
		}
	}
}

func Test_Run_writes_job_history_when_workspace_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{"--workspace", root, "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		HistoryPath string `json:"history_path"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.HistoryPath != "ceo-artifacts/jobs.jsonl" {
		t.Fatalf("HistoryPath = %q, want ceo-artifacts/jobs.jsonl", body.HistoryPath)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"))
	if err != nil {
		t.Fatalf("read history file: %v", err)
	}
	if !strings.Contains(string(got), `"verdict":"pass"`) {
		t.Fatalf("history = %q, want passing verdict", string(got))
	}
}
