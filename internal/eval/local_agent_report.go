package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func accumulateLocalAgentStatus(summary *LocalAgentSuiteSummary, status string) {
	switch status {
	case localAgentStatusPass:
		summary.Passed++
	case localAgentStatusTimeout:
		summary.TimedOut++
	case localAgentStatusSkipped:
		summary.Skipped++
	default:
		summary.Failed++
	}
}

func buildLocalAgentIterations(task string, results []LocalAgentResult) []LocalAgentIteration {
	iterations := []LocalAgentIteration{
		firstLocalAgentIteration(task),
	}
	for _, result := range results {
		if result.Status == localAgentStatusPass {
			continue
		}
		iterations = append(iterations, LocalAgentIteration{
			Priority: 2,
			Area:     result.ID,
			Finding:  localAgentFinding(result),
			NextStep: localAgentNextStep(result),
			Evidence: result.CommandPath,
		})
	}
	return iterations
}

func firstLocalAgentIteration(task string) LocalAgentIteration {
	if task == localAgentTaskEditFile {
		return LocalAgentIteration{
			Priority: 1,
			Area:     "benchmark-task-runner",
			Finding:  "edit-file proves live mutation, but it is only one tiny coding task",
			NextStep: "run the first real benchmark task with repo reset, scoring, and per-agent artifacts",
			Evidence: "summary.json",
		}
	}
	return LocalAgentIteration{
		Priority: 1,
		Area:     "live-task-runner",
		Finding:  "readiness ping is not yet a full coding-task comparison",
		NextStep: "run edit-file mode, then add isolated fixture reset plus pass/fail scoring for benchmark tasks",
		Evidence: "summary.json",
	}
}

func localAgentFinding(result LocalAgentResult) string {
	switch result.Status {
	case localAgentStatusTimeout:
		return result.Name + " timed out in the non-interactive readiness probe"
	case localAgentStatusSkipped:
		return result.Name + " binary was not found"
	default:
		if !result.FileMatched {
			return result.Name + " did not produce the exact expected file content"
		}
		if !result.OutputMatched {
			return result.Name + " exited without the expected marker/output"
		}
		return result.Name + " failed the non-interactive task despite saved command evidence"
	}
}

func localAgentNextStep(result LocalAgentResult) string {
	switch result.Status {
	case localAgentStatusTimeout:
		return "add a provider/auth preflight or shorter safe-mode command before live task comparison"
	case localAgentStatusSkipped:
		return result.SetupHint
	default:
		if !result.FileMatched {
			return "capture file diffs in the benchmark runner and keep exact-content scoring strict"
		}
		return "capture auth/setup stderr and add a tool-specific setup doctor before scoring live tasks"
	}
}

func writeLocalAgentMarkdown(path string, summary LocalAgentSuiteSummary) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Local Agent Comparison Suite\n\n")
	fmt.Fprintf(&builder, "Mode: `%s`\n\n", summary.Mode)
	fmt.Fprintf(&builder, "Task: `%s`\n\n", summary.Task)
	fmt.Fprintf(&builder, "Prompt: `%s`\n\n", summary.Prompt)
	fmt.Fprintf(&builder, "| Agent | Status | Exit | Duration ms | Output matched | File matched | Evidence |\n")
	fmt.Fprintf(&builder, "| --- | --- | ---: | ---: | --- | --- | --- |\n")
	for _, result := range summary.Results {
		fmt.Fprintf(
			&builder,
			"| %s | `%s` | %d | %d | %t | %t | `%s` |\n",
			result.Name,
			result.Status,
			result.ExitCode,
			result.DurationMS,
			result.OutputMatched,
			result.FileMatched,
			filepath.Dir(result.CommandPath),
		)
	}
	fmt.Fprintf(&builder, "\nPassed: %d\n", summary.Passed)
	fmt.Fprintf(&builder, "Failed: %d\n", summary.Failed)
	fmt.Fprintf(&builder, "Timed out: %d\n", summary.TimedOut)
	fmt.Fprintf(&builder, "Skipped: %d\n", summary.Skipped)
	return writeTextFile(path, builder.String())
}

func writeLocalAgentBacklog(path string, iterations []LocalAgentIteration) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Local Agent Improvement Backlog\n\n")
	for _, item := range iterations {
		fmt.Fprintf(&builder, "## P%d %s\n\n", item.Priority, item.Area)
		fmt.Fprintf(&builder, "- Finding: %s\n", item.Finding)
		fmt.Fprintf(&builder, "- Next: %s\n", item.NextStep)
		fmt.Fprintf(&builder, "- Evidence: `%s`\n\n", item.Evidence)
	}
	return writeTextFile(path, builder.String())
}

func writeTextFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
