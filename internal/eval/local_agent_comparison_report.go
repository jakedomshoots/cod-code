package eval

import (
	"fmt"
	"strings"
)

func writeLocalAgentComparisonReport(path string, summary LocalAgentBenchmarkSummary) error {
	agentOrder, agentStats := localAgentComparisonStats(summary.Results)
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Market Parity Comparison Report\n\n")
	fmt.Fprintf(&builder, "Mode: `%s`\n\n", summary.Mode)
	fmt.Fprintf(&builder, "Tasks: %d\n", summary.TaskCount)
	fmt.Fprintf(&builder, "Concurrency: %d\n", summary.Concurrency)
	fmt.Fprintf(&builder, "Runs: %d\n", summary.RunCount)
	fmt.Fprintf(&builder, "Incomplete evidence: %d\n\n", summary.IncompleteEvidence)
	fmt.Fprintf(&builder, "## Aggregate Counts\n\n")
	fmt.Fprintf(&builder, "Passed: %d\n", summary.Passed)
	fmt.Fprintf(&builder, "Partial: %d\n", summary.Partial)
	fmt.Fprintf(&builder, "Failed: %d\n", summary.Failed)
	fmt.Fprintf(&builder, "Timed out: %d\n", summary.TimedOut)
	fmt.Fprintf(&builder, "Setup blocked: %d\n", summary.SetupBlocked)
	fmt.Fprintf(&builder, "Skipped: %d\n", summary.Skipped)
	fmt.Fprintf(&builder, "Incomplete evidence: %d\n\n", summary.IncompleteEvidence)
	writeLocalAgentComparisonDecision(&builder, summary, agentOrder, agentStats)
	fmt.Fprintf(&builder, "## Per-Agent Status\n\n")
	fmt.Fprintf(&builder, "| Agent | Runs | Pass | Partial | Fail | Timeout | Setup blocked | Skipped | Evidence complete | Evidence incomplete | Duration ms |\n")
	fmt.Fprintf(&builder, "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, id := range agentOrder {
		stats := agentStats[id]
		if stats == nil {
			continue
		}
		fmt.Fprintf(
			&builder,
			"| %s | %d | %d | %d | %d | %d | %d | %d | %d | %d | %d |\n",
			stats.Name,
			stats.Runs,
			stats.Passed,
			stats.Partial,
			stats.Failed,
			stats.TimedOut,
			stats.SetupBlocked,
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

func localAgentComparisonStats(results []LocalAgentBenchmarkResult) ([]string, map[string]*localAgentComparisonAgentStats) {
	agentOrder := make([]string, 0)
	agentStats := make(map[string]*localAgentComparisonAgentStats)
	for _, result := range results {
		stats, ok := agentStats[result.ID]
		if !ok {
			stats = &localAgentComparisonAgentStats{Name: result.Name}
			agentStats[result.ID] = stats
			agentOrder = append(agentOrder, result.ID)
		}
		stats.add(result)
	}
	return agentOrder, agentStats
}

func writeLocalAgentComparisonDecision(builder *strings.Builder, summary LocalAgentBenchmarkSummary, agentOrder []string, agentStats map[string]*localAgentComparisonAgentStats) {
	overallClean := summary.RunCount > 0 &&
		summary.Passed == summary.RunCount &&
		summary.Partial == 0 &&
		summary.Failed == 0 &&
		summary.TimedOut == 0 &&
		summary.SetupBlocked == 0 &&
		summary.IncompleteEvidence == 0
	ceoStats := agentStats["ceo_harness"]
	ceoClean := ceoStats != nil &&
		ceoStats.Runs > 0 &&
		ceoStats.Passed == ceoStats.Runs &&
		ceoStats.Partial == 0 &&
		ceoStats.Failed == 0 &&
		ceoStats.TimedOut == 0 &&
		ceoStats.SetupBlocked == 0 &&
		ceoStats.EvidenceIncomplete == 0

	fmt.Fprintf(builder, "## Readiness Decision\n\n")
	if overallClean {
		fmt.Fprintf(builder, "Overall comparison: pass\n")
	} else {
		fmt.Fprintf(builder, "Overall comparison: blocked\n")
	}
	if ceoClean {
		fmt.Fprintf(builder, "Cod Code result: clean\n")
	} else {
		fmt.Fprintf(builder, "Cod Code result: needs attention\n")
	}
	blockers := make([]string, 0)
	for _, id := range agentOrder {
		if id == "ceo_harness" {
			continue
		}
		stats := agentStats[id]
		if stats == nil {
			continue
		}
		if stats.Partial > 0 || stats.Failed > 0 || stats.TimedOut > 0 || stats.SetupBlocked > 0 || stats.EvidenceIncomplete > 0 {
			blockers = append(blockers, fmt.Sprintf("%s partial=%d fail=%d timeout=%d setup_blocked=%d incomplete=%d", stats.Name, stats.Partial, stats.Failed, stats.TimedOut, stats.SetupBlocked, stats.EvidenceIncomplete))
		}
	}
	if len(blockers) == 0 {
		fmt.Fprintf(builder, "External blockers: none\n\n")
		return
	}
	fmt.Fprintf(builder, "External blockers: %s\n\n", strings.Join(blockers, "; "))
}

type localAgentComparisonAgentStats struct {
	Name               string
	Runs               int
	Passed             int
	Partial            int
	Failed             int
	TimedOut           int
	SetupBlocked       int
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
	case localAgentStatusSetupBlocked:
		stats.SetupBlocked++
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
