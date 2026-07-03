package eval

import "path/filepath"

func writeLocalAgentBenchmarkSummaryArtifacts(outputDir string, summary LocalAgentBenchmarkSummary) error {
	if err := writeJSONFile(filepath.Join(outputDir, "summary.json"), summary); err != nil {
		return err
	}
	if err := writeLocalAgentBenchmarkMarkdown(filepath.Join(outputDir, "summary.md"), summary); err != nil {
		return err
	}
	if err := writeLocalAgentComparisonReport(filepath.Join(outputDir, "comparison-report.md"), summary); err != nil {
		return err
	}
	return writeLocalAgentBacklog(filepath.Join(outputDir, "iteration-backlog.md"), summary.IterationBacklog)
}
