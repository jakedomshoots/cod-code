package eval

import "testing"

func Test_LocalAgentBenchmarkStatus_scoresPassFromSavedEvidenceWhenCommandNeedsInput(t *testing.T) {
	// Given: the CLI returned non-zero after writing complete task evidence.
	run := localAgentRunResult{exitCode: 1}

	// When
	status := localAgentBenchmarkStatus(run, "pass")

	// Then
	if status != localAgentStatusPass {
		t.Fatalf("status = %q, want %q", status, localAgentStatusPass)
	}
}

func Test_LocalAgentBenchmarkStatus_keepsSetupAndTimeoutAheadOfScore(t *testing.T) {
	setupBlocked := localAgentRunResult{exitCode: 1, stdout: "provider quota exhausted"}
	timedOut := localAgentRunResult{timedOut: true, exitCode: -1}

	// When / Then
	if status := localAgentBenchmarkStatus(setupBlocked, "pass"); status != localAgentStatusSetupBlocked {
		t.Fatalf("setup status = %q, want %q", status, localAgentStatusSetupBlocked)
	}
	if status := localAgentBenchmarkStatus(timedOut, "pass"); status != localAgentStatusTimeout {
		t.Fatalf("timeout status = %q, want %q", status, localAgentStatusTimeout)
	}
}

func Test_LocalAgentBenchmarkNote_doesNotClaimZeroExitForScorePass(t *testing.T) {
	// When
	note := localAgentBenchmarkNote(localAgentStatusPass)

	// Then
	if note != "agent scored pass on saved benchmark evidence" {
		t.Fatalf("note = %q", note)
	}
}
