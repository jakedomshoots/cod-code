package ceo

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"ceoharness/internal/subagent"
)

func Test_Runtime_RunJob_summarizes_verification_when_check_retries(t *testing.T) {
	// Given
	runtime := NewRuntime()
	statePath := filepath.Join(t.TempDir(), "retry-state")

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_retry_check",
		},
		CheckEnv:      []string{"GO_WANT_CEO_HELPER_PROCESS=retry", "GO_CEO_RETRY_STATE=" + statePath},
		CheckAttempts: 2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	summary := report.VerificationSummary
	if summary.SubagentPassCount != 3 || summary.SubagentFailCount != 0 {
		t.Fatalf("subagent counts = %d pass %d fail; want 3 pass 0 fail", summary.SubagentPassCount, summary.SubagentFailCount)
	}
	if summary.CheckAttemptCount != 2 {
		t.Fatalf("CheckAttemptCount = %d, want 2", summary.CheckAttemptCount)
	}
	if summary.CheckPassCount != 1 || summary.CheckFailCount != 1 {
		t.Fatalf("check counts = %d pass %d fail; want 1 pass 1 fail", summary.CheckPassCount, summary.CheckFailCount)
	}
	if summary.CheckTotalDurationMS < 0 {
		t.Fatalf("CheckTotalDurationMS = %d, want nonnegative duration", summary.CheckTotalDurationMS)
	}
}

func Test_summarizeVerification_counts_provider_failures_from_attempt_records(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status: "pass",
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", ProviderErrorKind: "rate_limited", ProviderHTTPStatus: 429},
				{Status: "pass"},
			},
		},
		{
			Status:             "fail",
			ProviderErrorKind:  "unauthorized",
			ProviderHTTPStatus: 401,
			AttemptRecords: []subagent.AttemptRecord{
				{Status: "fail", ProviderErrorKind: "unauthorized", ProviderHTTPStatus: 401},
			},
		},
		{
			Status:             "fail",
			ProviderErrorKind:  "unavailable",
			ProviderHTTPStatus: 502,
		},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if summary.ProviderErrorCount != 3 {
		t.Fatalf("ProviderErrorCount = %d, want 3", summary.ProviderErrorCount)
	}
	if summary.ProviderRateLimitedCount != 1 {
		t.Fatalf("ProviderRateLimitedCount = %d, want 1", summary.ProviderRateLimitedCount)
	}
	if summary.ProviderUnauthorizedCount != 1 {
		t.Fatalf("ProviderUnauthorizedCount = %d, want 1", summary.ProviderUnauthorizedCount)
	}
	if summary.ProviderUnavailableCount != 1 {
		t.Fatalf("ProviderUnavailableCount = %d, want 1", summary.ProviderUnavailableCount)
	}
	if summary.ProviderErrorKindCounts["rate_limited"] != 1 ||
		summary.ProviderErrorKindCounts["unauthorized"] != 1 ||
		summary.ProviderErrorKindCounts["unavailable"] != 1 {
		t.Fatalf("ProviderErrorKindCounts = %#v, want one count for each provider error kind", summary.ProviderErrorKindCounts)
	}
}

func Test_summarizeVerification_counts_subagent_retry_pressure(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			Status:   "pass",
			Attempts: 2,
			AttemptRecords: []subagent.AttemptRecord{
				{Attempt: 1, Status: "fail"},
				{Attempt: 2, Status: "pass"},
			},
		},
		{
			Status:   "fail",
			Attempts: 3,
			AttemptRecords: []subagent.AttemptRecord{
				{Attempt: 1, Status: "fail"},
				{Attempt: 2, Status: "fail"},
				{Attempt: 3, Status: "fail"},
			},
		},
		{Status: "pass"},
	}

	// When
	summary := summarizeVerification(results, nil)

	// Then
	if summary.SubagentAttemptCount != 6 {
		t.Fatalf("SubagentAttemptCount = %d, want 6", summary.SubagentAttemptCount)
	}
	if summary.SubagentRetryCount != 3 {
		t.Fatalf("SubagentRetryCount = %d, want 3", summary.SubagentRetryCount)
	}
	if summary.SubagentRetriedCount != 2 {
		t.Fatalf("SubagentRetriedCount = %d, want 2", summary.SubagentRetriedCount)
	}
	if summary.SubagentRetryExhaustedCount != 1 {
		t.Fatalf("SubagentRetryExhaustedCount = %d, want 1", summary.SubagentRetryExhaustedCount)
	}
}
