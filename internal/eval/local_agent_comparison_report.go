package eval

import (
	"fmt"
	"strings"
)

func writeLocalAgentComparisonReport(path string, summary LocalAgentBenchmarkSummary) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Market Parity Comparison Report\n\n")
	fmt.Fprintf(&builder, "Mode: `%s`\n\n", summary.Mode)
	fmt.Fprintf(&builder, "Tasks: %d\n", summary.TaskCount)
	fmt.Fprintf(&builder, "Runs: %d\n", summary.RunCount)
	fmt.Fprintf(&builder, "Incomplete evidence: %d\n\n", summary.IncompleteEvidence)
	fmt.Fprintf(&builder, "## Aggregate Counts\n\n")
	fmt.Fprintf(&builder, "Passed: %d\n", summary.Passed)
	fmt.Fprintf(&builder, "Partial: %d\n", summary.Partial)
	fmt.Fprintf(&builder, "Failed: %d\n", summary.Failed)
	fmt.Fprintf(&builder, "Timed out: %d\n", summary.TimedOut)
	fmt.Fprintf(&builder, "Skipped: %d\n", summary.Skipped)
	fmt.Fprintf(&builder, "Incomplete evidence: %d\n\n", summary.IncompleteEvidence)
	agentOrder := make([]string, 0)
	agentStats := make(map[string]*localAgentComparisonAgentStats)
	for _, result := range summary.Results {
		stats, ok := agentStats[result.ID]
		if !ok {
			stats = &localAgentComparisonAgentStats{Name: result.Name}
			agentStats[result.ID] = stats
			agentOrder = append(agentOrder, result.ID)
		}
		stats.add(result)
	}
	fmt.Fprintf(&builder, "## Per-Agent Status\n\n")
	fmt.Fprintf(&builder, "| Agent | Runs | Pass | Partial | Fail | Timeout | Skipped | Evidence complete | Evidence incomplete | Duration ms |\n")
	fmt.Fprintf(&builder, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, id := range agentOrder {
		stats := agentStats[id]
		if stats == nil {
			continue
		}
		fmt.Fprintf(
			&builder,
			"| %s | %d | %d | %d | %d | %d | %d | %d | %d | %d |\n",
			stats.Name,
			stats.Runs,
			stats.Passed,
			stats.Partial,
			stats.Failed,
			stats.TimedOut,
			stats.Skipped,
			stats.EvidenceComplete,
			stats.EvidenceIncomplete,
			stats.DurationMS,
		)
	}
	fmt.Fprintf(&builder, "\n## Run Artifact Detail\n\n")
	fmt.Fprintf(&builder, "| Task | Run | Agent | Result | Evidence | Incomplete evidence reasons | Score | Changed files | Extra files | Duration ms | Timing | Report | Score artifact |\n")
	fmt.Fprintf(&builder, "| --- | ---: | --- | --- | --- | --- | ---: | --- | --- | ---: | --- | --- | --- |\n")
	for _, result := range summary.Results {
		fmt.Fprintf(
			&builder,
			"| `%s` | %d | %s | `%s` | `%s` | %s | %d/%d | `%s` | `%s` | %d | `%s` | `%s` | `%s` |\n",
			result.TaskID,
			result.Attempt,
			result.Name,
			result.Status,
			result.EvidenceStatus,
			comparisonEvidenceReasons(result.FailedScoreChecks),
			result.PassedChecks,
			result.TotalChecks,
			comparisonList(result.ChangedFiles),
			comparisonList(result.ExtraChangedFiles),
			result.DurationMS,
			result.TimingPath,
			result.ReportPath,
			result.ScorePath,
		)
	}
	return writeTextFile(path, builder.String())
}

type localAgentComparisonAgentStats struct {
	Name               string
	Runs               int
	Passed             int
	Partial            int
	Failed             int
	TimedOut           int
	Skipped            int
	EvidenceComplete   int
	EvidenceIncomplete int
	DurationMS         int64
}

func (stats *localAgentComparisonAgentStats) add(result LocalAgentBenchmarkResult) {
	stats.Runs++
	stats.DurationMS += result.DurationMS
	switch result.Status {
	case localAgentStatusPass:
		stats.Passed++
	case localAgentStatusPartial:
		stats.Partial++
	case localAgentStatusTimeout:
		stats.TimedOut++
	case localAgentStatusSkipped:
		stats.Skipped++
	default:
		stats.Failed++
	}
	switch result.EvidenceStatus {
	case localAgentEvidenceComplete:
		stats.EvidenceComplete++
	case localAgentEvidenceIncomplete:
		stats.EvidenceIncomplete++
	}
}

func comparisonEvidenceReasons(checks []CheckResult) string {
	reasons := make([]string, 0)
	for _, check := range checks {
		if isEvidenceCheck(check.Name) {
			reasons = append(reasons, check.Name)
		}
	}
	return comparisonList(reasons)
}

func comparisonList(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
