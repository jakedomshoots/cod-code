package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test_LocalAgentStatus_credit_balance_marker_classifies_setup_blocked pins the
// contract that a non-timeout run whose stdout/stderr/errText contains the
// "Credit balance is too low" provider marker is classified as setup_blocked
// even when the agent exits with a nonzero code. This is the exact phrasing
// surfaced by providers when usage/billing is exhausted; misclassifying it as
// a generic fail would hide setup regressions behind red CI.
func Test_LocalAgentStatus_credit_balance_marker_classifies_setup_blocked(t *testing.T) {
	// Given
	run := localAgentRunResult{
		stdout:   "some startup banner\n",
		stderr:   "Error: Credit balance is too low. Visit https://example.test/billing to top up.\n",
		errText:  "exit status 1",
		exitCode: 1,
	}

	// When
	status := localAgentStatus(run, false, false)

	// Then
	if status != localAgentStatusSetupBlocked {
		t.Fatalf("localAgentStatus = %q, want %q (run=%+v)", status, localAgentStatusSetupBlocked, run)
	}
}

// Test_LocalAgentStatus_setup_blocked_marker_via_errText covers the same
// classification when the marker surfaces in errText (Go's exit error message)
// rather than captured stdout/stderr. Mirrors what shell exec surfaces when a
// provider CLI aborts before printing to stderr.
func Test_LocalAgentStatus_setup_blocked_marker_via_errText(t *testing.T) {
	// Given
	run := localAgentRunResult{
		stdout:   "",
		stderr:   "",
		errText:  "exec: provider returned Credit balance is too low",
		exitCode: 2,
	}

	// When
	status := localAgentStatus(run, false, false)

	// Then
	if status != localAgentStatusSetupBlocked {
		t.Fatalf("localAgentStatus = %q, want %q (run=%+v)", status, localAgentStatusSetupBlocked, run)
	}
}

// Test_LocalAgentStatus_timeout_takes_precedence_over_setup_blocked guards
// against a regression where timeout classification is checked after the
// setup-blocked marker check. A hung provider CLI that also dumps the
// balance marker must still be reported as a timeout, not as setup-blocked,
// otherwise the harness loses the signal that the process was canceled.
func Test_LocalAgentStatus_timeout_takes_precedence_over_setup_blocked(t *testing.T) {
	// Given
	run := localAgentRunResult{
		stdout:   "",
		stderr:   "Credit balance is too low\n",
		errText:  "command timed out",
		exitCode: -1,
		timedOut: true,
	}

	// When
	status := localAgentStatus(run, false, false)

	// Then
	if status != localAgentStatusTimeout {
		t.Fatalf("localAgentStatus = %q, want %q (run=%+v)", status, localAgentStatusTimeout, run)
	}
}

// Test_AccumulateLocalAgentStatus_setup_blocked_increments_separate_counter
// is the load-bearing assertion for the suite accounting contract: when a
// result is classified setup_blocked, the suite summary must count it under
// SetupBlocked and MUST NOT also count it as Failed. A regression that drops
// the setup_blocked case (falling through to the default) would double-count
// the same run as both setup-blocked and failed in downstream reporting.
func Test_AccumulateLocalAgentStatus_setup_blocked_increments_separate_counter(t *testing.T) {
	// Given
	summary := LocalAgentSuiteSummary{}

	// When
	accumulateLocalAgentStatus(&summary, localAgentStatusSetupBlocked)

	// Then
	if summary.SetupBlocked != 1 {
		t.Fatalf("SetupBlocked = %d, want 1 (summary=%+v)", summary.SetupBlocked, summary)
	}
	if summary.Failed != 0 {
		t.Fatalf("Failed = %d, want 0 (summary=%+v)", summary.Failed, summary)
	}
	if summary.Passed != 0 || summary.TimedOut != 0 || summary.Skipped != 0 {
		t.Fatalf("unrelated counters changed: %+v", summary)
	}
}

// Test_AccumulateLocalAgentStatus_other_statuses_still_routes correctly
// guards the routing table so the new setup_blocked case did not displace an
// existing case (e.g. the timeout arm silently merging into setup_blocked).
// This walks every status string the suite uses today.
func Test_AccumulateLocalAgentStatus_other_statuses_still_routes(t *testing.T) {
	cases := []struct {
		name           string
		status         string
		wantPassed     int
		wantFailed     int
		wantTimedOut   int
		wantSetupBlock int
		wantSkipped    int
	}{
		{name: "pass", status: localAgentStatusPass, wantPassed: 1},
		{name: "fail", status: localAgentStatusFail, wantFailed: 1},
		{name: "timeout", status: localAgentStatusTimeout, wantTimedOut: 1},
		{name: "setup_blocked", status: localAgentStatusSetupBlocked, wantSetupBlock: 1},
		{name: "skipped", status: localAgentStatusSkipped, wantSkipped: 1},
		{name: "unknown_falls_to_failed", status: "mystery", wantFailed: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary := LocalAgentSuiteSummary{}
			accumulateLocalAgentStatus(&summary, tc.status)
			if summary.Passed != tc.wantPassed ||
				summary.Failed != tc.wantFailed ||
				summary.TimedOut != tc.wantTimedOut ||
				summary.SetupBlocked != tc.wantSetupBlock ||
				summary.Skipped != tc.wantSkipped {
				t.Fatalf("status=%q summary=%+v, want passed=%d failed=%d timed_out=%d setup_blocked=%d skipped=%d",
					tc.status, summary,
					tc.wantPassed, tc.wantFailed, tc.wantTimedOut, tc.wantSetupBlock, tc.wantSkipped)
			}
		})
	}
}

// Test_AccumulateLocalAgentStatus_mixed_batch_keeps_each_counter_independent
// is the integration view of the same contract: feeding a realistic mix of
// statuses must produce the per-bucket totals without any bucket spilling into
// another. This is what the suite relies on when reporting the breakdown of
// provider-auth failures vs ordinary task failures vs timeouts.
func Test_AccumulateLocalAgentStatus_mixed_batch_keeps_each_counter_independent(t *testing.T) {
	// Given
	summary := LocalAgentSuiteSummary{}
	feed := []string{
		localAgentStatusPass,
		localAgentStatusSetupBlocked,
		localAgentStatusFail,
		localAgentStatusTimeout,
		localAgentStatusSetupBlocked,
		localAgentStatusSkipped,
		localAgentStatusPass,
	}

	// When
	for _, status := range feed {
		accumulateLocalAgentStatus(&summary, status)
	}

	// Then
	wantPassed, wantFailed, wantTimedOut, wantSetupBlock, wantSkipped := 2, 1, 1, 2, 1
	if summary.Passed != wantPassed ||
		summary.Failed != wantFailed ||
		summary.TimedOut != wantTimedOut ||
		summary.SetupBlocked != wantSetupBlock ||
		summary.Skipped != wantSkipped {
		t.Fatalf("summary=%+v, want passed=%d failed=%d timed_out=%d setup_blocked=%d skipped=%d",
			summary, wantPassed, wantFailed, wantTimedOut, wantSetupBlock, wantSkipped)
	}
}

// Test_WriteLocalAgentMarkdown_includes_setup_blocked_count pins the surface
// contract for the human-readable suite report: when a summary records any
// setup-blocked agent, the markdown must surface that count, otherwise the
// report hides provider-side failures behind the ordinary Failed line.
func Test_WriteLocalAgentMarkdown_includes_setup_blocked_count(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "summary.md")
	summary := LocalAgentSuiteSummary{
		Mode:         localAgentSuiteMode,
		Task:         localAgentTaskReadiness,
		Prompt:       "readiness ping",
		AgentCount:   3,
		Passed:       1,
		Failed:       1,
		SetupBlocked: 1,
		Results: []LocalAgentResult{
			{ID: "ceo_harness", Name: "CEO Harness", Status: localAgentStatusPass, ExitCode: 0},
			{ID: "codex_cli", Name: "Codex CLI", Status: localAgentStatusFail, ExitCode: 1},
			{ID: "opencode", Name: "OpenCode", Status: localAgentStatusSetupBlocked, ExitCode: 1},
		},
	}

	// When
	if err := writeLocalAgentMarkdown(path, summary); err != nil {
		t.Fatalf("writeLocalAgentMarkdown returned error: %v", err)
	}

	// Then
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(content)
	wantLine := "Setup blocked: 1"
	if !strings.Contains(text, wantLine) {
		t.Fatalf("markdown missing %q:\n%s", wantLine, text)
	}
	wantPassed := "Passed: 1"
	if !strings.Contains(text, wantPassed) {
		t.Fatalf("markdown missing %q:\n%s", wantPassed, text)
	}
}
