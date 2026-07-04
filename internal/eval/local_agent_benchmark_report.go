package eval

import (
	"fmt"
	"sort"
	"strings"
)

func accumulateLocalAgentBenchmarkStatus(summary *LocalAgentBenchmarkSummary, status string) {
	switch status {
	case localAgentStatusPass:
		summary.Passed++
	case localAgentStatusPartial:
		summary.Partial++
	case localAgentStatusTimeout:
		summary.TimedOut++
	case localAgentStatusSetupBlocked:
		summary.SetupBlocked++
	case localAgentStatusSkipped:
		summary.Skipped++
	default:
		summary.Failed++
	}
}

func accumulateLocalAgentBenchmarkEvidence(summary *LocalAgentBenchmarkSummary, status string) {
	if status == localAgentEvidenceIncomplete {
		summary.IncompleteEvidence++
	}
}

func localAgentBenchmarkEvidenceStatus(status string, failedChecks []CheckResult) string {
	if status == localAgentStatusSkipped {
		return localAgentEvidenceNotRun
	}
	if status == localAgentStatusSetupBlocked {
		return localAgentEvidenceComplete
	}
	for _, check := range failedChecks {
		if isEvidenceCheck(check.Name) {
			return localAgentEvidenceIncomplete
		}
	}
	return localAgentEvidenceComplete
}

func isEvidenceCheck(name string) bool {
	return strings.HasPrefix(name, "artifact:") ||
		strings.HasPrefix(name, "report_field:") ||
		strings.HasPrefix(name, "dirty_worktree_")
}

func buildLocalAgentBenchmarkIterations(results []LocalAgentBenchmarkResult) []LocalAgentIteration {
	iterations := []LocalAgentIteration{
		{
			Priority: 1,
			Area:     "benchmark-task-runner",
			Finding:  "benchmark comparison scores saved task runs from actual file changes",
			NextStep: "add richer task fixtures, per-agent aggregate thresholds, and harder code-edit assertions",
			Evidence: "summary.json",
		},
	}
	for _, result := range results {
		if len(result.ExtraChangedFiles) > 0 {
			iterations = append(iterations, LocalAgentIteration{
				Priority: 2,
				Area:     result.ID + "-file-footprint",
				Finding:  fmt.Sprintf("%s changed %d extra file(s) beyond task requirements", result.Name, len(result.ExtraChangedFiles)),
				NextStep: "isolate harness/runtime artifacts outside scored workspaces or mark them as ignored evidence",
				Evidence: result.ChangedFilesPath,
			})
		}
		if result.Status != localAgentStatusPass {
			iterations = append(iterations, LocalAgentIteration{
				Priority: 2,
				Area:     result.ID,
				Finding:  localAgentBenchmarkFinding(result),
				NextStep: localAgentBenchmarkNextStep(result),
				Evidence: result.ScorePath,
			})
		}
	}
	return iterations
}

func localAgentBenchmarkFinding(result LocalAgentBenchmarkResult) string {
	switch result.Status {
	case localAgentStatusTimeout:
		return result.Name + " timed out on the benchmark task"
	case localAgentStatusSetupBlocked:
		return result.Name + " setup is blocked by provider auth, quota, or credential state"
	case localAgentStatusSkipped:
		return result.Name + " binary was not found"
	case localAgentStatusPartial:
		return fmt.Sprintf("%s only passed %d/%d benchmark checks", result.Name, result.PassedChecks, result.TotalChecks)
	default:
		return result.Name + " failed the benchmark task or command"
	}
}

func localAgentBenchmarkNextStep(result LocalAgentBenchmarkResult) string {
	switch result.Status {
	case localAgentStatusTimeout:
		return "shorten the task prompt or add a tool-specific timeout/auth preflight"
	case localAgentStatusSetupBlocked:
		return "repair the provider login, quota, or selected model, then rerun with saved stderr/stdout evidence"
	case localAgentStatusSkipped:
		return result.SetupHint
	case localAgentStatusPartial:
		return "inspect score.json and improve prompt or harness task setup for missing checks"
	default:
		return "inspect stdout, stderr, diff, and score artifacts before adding the next benchmark"
	}
}

func writeLocalAgentBenchmarkMarkdown(path string, summary LocalAgentBenchmarkSummary) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Local Agent Benchmark\n\n")
	fmt.Fprintf(&builder, "Mode: `%s`\n\n", summary.Mode)
	fmt.Fprintf(&builder, "Tasks: %d\n", summary.TaskCount)
	fmt.Fprintf(&builder, "Repeats: %d\n", summary.RepeatCount)
	fmt.Fprintf(&builder, "Concurrency: %d\n", summary.Concurrency)
	fmt.Fprintf(&builder, "Timeout retries: %d\n", summary.TimeoutRetries)
	if len(summary.AgentTimeouts) > 0 {
		fmt.Fprintf(&builder, "Agent timeouts: %s\n", formatAgentTimeouts(summary.AgentTimeouts))
	}
	if len(summary.AgentModels) > 0 {
		fmt.Fprintf(&builder, "Agent models: %s\n", formatAgentModels(summary.AgentModels))
	}
	fmt.Fprintf(&builder, "Runs: %d\n\n", summary.RunCount)
	fmt.Fprintf(&builder, "| Task | Run | Retry | Agent | Status | Score | Exit | Duration ms | Extra files | Changed files | Evidence |\n")
	fmt.Fprintf(&builder, "| --- | ---: | ---: | --- | --- | ---: | ---: | ---: | ---: | --- | --- |\n")
	for _, result := range summary.Results {
		fmt.Fprintf(
			&builder,
			"| `%s` | %d | %d | %s | `%s` | %d/%d | %d | %d | %d | `%s` | `%s` |\n",
			result.TaskID,
			result.Attempt,
			result.RunAttempt,
			result.Name,
			result.Status,
			result.PassedChecks,
			result.TotalChecks,
			result.ExitCode,
			result.DurationMS,
			len(result.ExtraChangedFiles),
			strings.Join(result.ChangedFiles, ", "),
			result.ScorePath,
		)
	}
	fmt.Fprintf(&builder, "\nPassed: %d\n", summary.Passed)
	fmt.Fprintf(&builder, "Partial: %d\n", summary.Partial)
	fmt.Fprintf(&builder, "Failed: %d\n", summary.Failed)
	fmt.Fprintf(&builder, "Timed out: %d\n", summary.TimedOut)
	fmt.Fprintf(&builder, "Setup blocked: %d\n", summary.SetupBlocked)
	fmt.Fprintf(&builder, "Skipped: %d\n", summary.Skipped)
	fmt.Fprintf(&builder, "Incomplete evidence: %d\n", summary.IncompleteEvidence)
	return writeTextFile(path, builder.String())
}

func formatAgentTimeouts(timeouts map[string]int) string {
	if len(timeouts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(timeouts))
	for agent, seconds := range timeouts {
		parts = append(parts, fmt.Sprintf("%s=%ds", agent, seconds))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func formatAgentModels(models map[string]string) string {
	if len(models) == 0 {
		return ""
	}
	parts := make([]string, 0, len(models))
	for agent, model := range models {
		parts = append(parts, fmt.Sprintf("%s=%s", agent, model))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func writeBenchmarkRunEvidence(result LocalAgentBenchmarkResult, run localAgentRunResult) error {
	if err := writeJSONFile(result.CommandPath, map[string][]string{"command": result.Command}); err != nil {
		return err
	}
	if err := writeTextFile(result.StdoutPath, nonEmptyLog(run.stdout)); err != nil {
		return err
	}
	if err := writeTextFile(result.StderrPath, nonEmptyLog(run.stderr)); err != nil {
		return err
	}
	timing := fmt.Sprintf("duration_ms=%d\nexit_code=%d\ntimed_out=%t\n", run.duration.Milliseconds(), run.exitCode, run.timedOut)
	return writeTextFile(result.TimingPath, timing)
}
