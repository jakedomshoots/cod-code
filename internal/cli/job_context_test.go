package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ceoharness/internal/ceo"
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_Run_prints_compact_job_context_when_job_context_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:                    "Fix checkout",
		TaskKind:                "coding",
		RiskLevel:               "medium",
		Verdict:                 "needs_input",
		ChangedFiles:            []string{"app.go"},
		ExecutionPlanNextAction: "answer subagent questions",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	report := ceo.Report{
		JobID:   "job-000001",
		Verdict: "needs_input",
		JobPacket: jobpacket.Packet{
			Task:        "Fix checkout",
			TaskProfile: jobpacket.TaskProfile{Kind: "coding", RiskLevel: "medium"},
		},
		Resume: &ceo.ResumeContext{
			Questions: []string{"Which package should I change?"},
		},
		SubagentResults: []subagent.Result{
			{
				AgentName: "scanner",
				Role:      "inspect scope",
				Status:    "needs_input",
				Summary:   "Need target package",
				Questions: []string{"Which package should I change?"},
			},
		},
		ChangedFiles: []string{"app.go", "README.md"},
		RunLedger: ceo.RunLedger{
			Owner:              "scanner",
			Verdict:            "needs_input",
			NextAction:         "answer subagent questions",
			VerificationStatus: "fail",
			ChangedFileCount:   2,
			ChangedFiles:       []string{"app.go", "README.md"},
			ProviderRouteCount: 1,
			ProviderRouteReasons: []string{
				"provider_policy.default_provider",
			},
		},
		CheckResults: []checkrunner.Result{
			{
				Argv:        []string{"go", "test", "./..."},
				Status:      "fail",
				ExitCode:    1,
				CheckIndex:  2,
				Attempt:     3,
				MaxAttempts: 3,
				DurationMS:  456,
				Stderr:      "FAIL\ncheckout panic",
			},
		},
		ExecutionPlan: ceo.ExecutionPlan{NextAction: "answer subagent questions"},
		CEOReview:     &ceo.CEOReview{Summary: "Blocked on missing package"},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", payload); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--job-context", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobContext struct {
			JobID      string `json:"job_id"`
			Task       string `json:"task"`
			Verdict    string `json:"verdict"`
			TaskKind   string `json:"task_kind"`
			RiskLevel  string `json:"risk_level"`
			NextAction string `json:"next_action"`
			RunLedger  struct {
				Owner                string   `json:"owner"`
				Verdict              string   `json:"verdict"`
				NextAction           string   `json:"next_action"`
				VerificationStatus   string   `json:"verification_status"`
				ChangedFileCount     int      `json:"changed_file_count"`
				ProviderRouteReasons []string `json:"provider_route_reasons"`
			} `json:"run_ledger"`
			Questions        []string `json:"questions"`
			SuggestedCommand string   `json:"suggested_command"`
			ChangedFiles     []string `json:"changed_files"`
			CEOReviewSummary string   `json:"ceo_review_summary"`
			FailedChecks     []struct {
				Command        []string `json:"command"`
				Status         string   `json:"status"`
				ExitCode       int      `json:"exit_code"`
				CheckIndex     int      `json:"check_index"`
				Attempt        int      `json:"attempt"`
				MaxAttempts    int      `json:"max_attempts"`
				DurationMS     int64    `json:"duration_ms"`
				FailureExcerpt string   `json:"failure_excerpt"`
			} `json:"failed_checks"`
			Subagents []struct {
				Name    string `json:"name"`
				Status  string `json:"status"`
				Summary string `json:"summary"`
			} `json:"subagents"`
		} `json:"job_context"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobContext.JobID != "job-000001" || body.JobContext.Task != "Fix checkout" {
		t.Fatalf("job context = %#v, want compact job identity", body.JobContext)
	}
	if body.JobContext.Verdict != "needs_input" || body.JobContext.NextAction != "answer subagent questions" {
		t.Fatalf("job context = %#v, want needs_input next action", body.JobContext)
	}
	if body.JobContext.RunLedger.Owner != "scanner" || body.JobContext.RunLedger.VerificationStatus != "fail" {
		t.Fatalf("run ledger = %#v, want scanner failed verification", body.JobContext.RunLedger)
	}
	if body.JobContext.RunLedger.ChangedFileCount != 2 || !containsString(body.JobContext.RunLedger.ProviderRouteReasons, "provider_policy.default_provider") {
		t.Fatalf("run ledger = %#v, want compact changed files and route reason", body.JobContext.RunLedger)
	}
	if body.JobContext.TaskKind != "coding" || body.JobContext.RiskLevel != "medium" {
		t.Fatalf("profile = %q/%q, want coding/medium", body.JobContext.TaskKind, body.JobContext.RiskLevel)
	}
	if len(body.JobContext.Questions) != 1 || body.JobContext.Questions[0] != "Which package should I change?" {
		t.Fatalf("questions = %#v, want compact resume question", body.JobContext.Questions)
	}
	if !strings.Contains(body.JobContext.SuggestedCommand, "--resume job-000001 --answer") {
		t.Fatalf("suggested command = %q, want resume command", body.JobContext.SuggestedCommand)
	}
	if len(body.JobContext.ChangedFiles) != 2 || body.JobContext.ChangedFiles[1] != "README.md" {
		t.Fatalf("changed files = %#v, want snapshot changed files", body.JobContext.ChangedFiles)
	}
	if len(body.JobContext.Subagents) != 1 || body.JobContext.Subagents[0].Name != "scanner" {
		t.Fatalf("subagents = %#v, want scanner summary", body.JobContext.Subagents)
	}
	if len(body.JobContext.FailedChecks) != 1 || !strings.Contains(body.JobContext.FailedChecks[0].FailureExcerpt, "checkout panic") {
		t.Fatalf("failed checks = %#v, want compact failure excerpt", body.JobContext.FailedChecks)
	}
	if body.JobContext.FailedChecks[0].CheckIndex != 2 ||
		body.JobContext.FailedChecks[0].Attempt != 3 ||
		body.JobContext.FailedChecks[0].MaxAttempts != 3 ||
		body.JobContext.FailedChecks[0].DurationMS != 456 {
		t.Fatalf("failed check metadata = %#v, want index/attempt/duration", body.JobContext.FailedChecks[0])
	}
	if body.JobContext.CEOReviewSummary != "Blocked on missing package" {
		t.Fatalf("CEOReviewSummary = %q, want review summary", body.JobContext.CEOReviewSummary)
	}
}
