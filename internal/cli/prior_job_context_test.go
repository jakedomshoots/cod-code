package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"ceoharness/internal/ceo"
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_Run_injects_compact_job_context_when_with_job_context_flag_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:                    "Fix checkout",
		TaskKind:                "coding",
		RiskLevel:               "medium",
		Verdict:                 "fail",
		ChangedFiles:            []string{"checkout.go"},
		ExecutionPlanNextAction: "fix failing checks",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	report := ceo.Report{
		JobID:   "job-000001",
		Verdict: "fail",
		JobPacket: jobpacket.Packet{
			Task:        "Fix checkout",
			TaskProfile: jobpacket.TaskProfile{Kind: "coding", RiskLevel: "medium"},
		},
		SubagentResults: []subagent.Result{
			{AgentName: "coder", Role: "apply bounded changes", Status: "pass", Summary: "Patched checkout handler"},
		},
		ChangedFiles: []string{"checkout.go"},
		RunLedger: ceo.RunLedger{
			Owner:              "coder",
			Verdict:            "fail",
			NextAction:         "fix failing checks",
			VerificationStatus: "fail",
			ChangedFileCount:   1,
			ChangedFiles:       []string{"checkout.go"},
		},
		CheckResults: []checkrunner.Result{
			{
				Argv:        []string{"go", "test", "./..."},
				Status:      "fail",
				ExitCode:    1,
				CheckIndex:  1,
				Attempt:     1,
				MaxAttempts: 2,
				DurationMS:  123,
				Stderr:      "FAIL\nfirst checkout panic",
			},
			{
				Argv:        []string{"go", "test", "./..."},
				Status:      "fail",
				ExitCode:    1,
				CheckIndex:  1,
				Attempt:     2,
				MaxAttempts: 2,
				DurationMS:  789,
				Stderr:      "FAIL\ncheckout panic",
			},
		},
		ExecutionPlan: ceo.ExecutionPlan{NextAction: "fix failing checks"},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", payload); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_PRIOR_CONTEXT_MODEL", "1")

	// When
	err = Run(context.Background(), &out, []string{
		"--workspace", root,
		"--with-job-context", "job-000001",
		"--model-command", os.Args[0],
		"-test.run=Test_HelperProcess_cli_prior_context_model",
		"--",
		"Continue", "checkout",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Verdict   string `json:"verdict"`
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if !strings.Contains(body.JobPacket.Task, "prior_job_context:") ||
		!strings.Contains(body.JobPacket.Task, "previous_job: job-000001") ||
		!strings.Contains(body.JobPacket.Task, "previous_run_ledger: owner=coder verdict=fail next=\"fix failing checks\" verification=fail changed=1 routes=0") ||
		!strings.Contains(body.JobPacket.Task, "go test ./... [fail] index=1 attempt=2/2 duration_ms=789") {
		t.Fatalf("task = %q, want compact previous job context", body.JobPacket.Task)
	}
	if strings.Contains(body.JobPacket.Task, "attempt=1/2") ||
		strings.Contains(body.JobPacket.Task, "first checkout panic") {
		t.Fatalf("task = %q, want only final failed retry in prior context", body.JobPacket.Task)
	}
}

func Test_HelperProcess_cli_prior_context_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_PRIOR_CONTEXT_MODEL") == "" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "prior_job_context:") &&
		strings.Contains(text, "previous_job: job-000001") &&
		strings.Contains(text, "previous_run_ledger: owner=coder verdict=fail") &&
		strings.Contains(text, "go test ./... [fail] index=1 attempt=2/2 duration_ms=789") &&
		!strings.Contains(text, "attempt=1/2") &&
		strings.Contains(text, "checkout panic") {
		os.Stdout.WriteString(`{"summary":"prior context received"}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"status":"needs_input","summary":"missing prior context","questions":["Where is the prior job context?"]}`)
	os.Exit(0)
}
