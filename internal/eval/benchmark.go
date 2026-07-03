package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const benchmarkSchemaVersion = 1

func RunBenchmarkFixtures(ctx context.Context, req BenchmarkFixtureRequest) (BenchmarkSummary, error) {
	tasks, err := LoadTasks(ctx, req.TasksDir)
	if err != nil {
		return BenchmarkSummary{}, err
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return BenchmarkSummary{}, fmt.Errorf("create benchmark output dir: %w", err)
	}
	summary := BenchmarkSummary{
		SchemaVersion: benchmarkSchemaVersion,
		Mode:          req.ReportMode,
		TaskCount:     len(tasks),
		Results:       make([]BenchmarkTaskResult, 0, len(tasks)),
	}
	for _, task := range tasks {
		result := runBenchmarkFixtureTask(ctx, req.OutputDir, task)
		summary.Results = append(summary.Results, result)
		switch result.Verdict {
		case "pass":
			summary.Passed++
		case "partial":
			summary.Partial++
		case "fail":
			summary.Failed++
		default:
			summary.Skipped++
		}
	}
	if err := writeJSONFile(filepath.Join(req.OutputDir, "summary.json"), summary); err != nil {
		return BenchmarkSummary{}, err
	}
	if err := writeBenchmarkMarkdown(filepath.Join(req.OutputDir, "summary.md"), summary); err != nil {
		return BenchmarkSummary{}, err
	}
	return summary, nil
}

func runBenchmarkFixtureTask(ctx context.Context, outputDir string, task Task) BenchmarkTaskResult {
	taskDir := filepath.Join(outputDir, task.ID)
	reportPath := filepath.Join(taskDir, "report.json")
	scorePath := filepath.Join(taskDir, "score.json")
	logPath := filepath.Join(taskDir, "score.log")
	result := BenchmarkTaskResult{
		TaskID:     task.ID,
		Verdict:    "skipped_unscored",
		ReportPath: reportPath,
		ScorePath:  scorePath,
		LogPath:    logPath,
	}
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		result.Reason = err.Error()
		return result
	}
	workspaceDir, status, err := prepareBenchmarkWorkspace(ctx, taskDir, task)
	if err != nil {
		return benchmarkTaskError(result, err)
	}
	if err := writeBenchmarkArtifacts(taskDir, task); err != nil {
		return benchmarkTaskError(result, err)
	}
	payload, err := benchmarkReportPayload(task, status)
	if err != nil {
		return benchmarkTaskError(result, err)
	}
	if err := writeJSONFile(reportPath, payload); err != nil {
		return benchmarkTaskError(result, err)
	}
	score, err := ScoreReport(ctx, ScoreRequest{
		Task:         task,
		ReportPath:   reportPath,
		WorkspaceDir: workspaceDir,
	})
	if err != nil {
		return benchmarkTaskError(result, err)
	}
	if err := writeJSONFile(scorePath, score); err != nil {
		return benchmarkTaskError(result, err)
	}
	if err := os.WriteFile(logPath, []byte(benchmarkLog(task, score)), 0o644); err != nil {
		return benchmarkTaskError(result, err)
	}
	result.Verdict = score.Verdict
	result.Passed = score.Passed
	result.Total = score.Total
	return result
}

func benchmarkTaskError(result BenchmarkTaskResult, err error) BenchmarkTaskResult {
	result.Reason = err.Error()
	return result
}

func benchmarkLog(task Task, score ScoreResult) string {
	return fmt.Sprintf("task=%s verdict=%s passed=%d total=%d mode=deterministic_fixture_scoring\n", task.ID, score.Verdict, score.Passed, score.Total)
}

func writeBenchmarkMarkdown(path string, summary BenchmarkSummary) error {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Benchmark Fixture Summary\n\n")
	fmt.Fprintf(&builder, "Mode: `%s`\n\n", summary.Mode)
	fmt.Fprintf(&builder, "Tasks: %d\n", summary.TaskCount)
	fmt.Fprintf(&builder, "Passed: %d\n", summary.Passed)
	fmt.Fprintf(&builder, "Partial: %d\n", summary.Partial)
	fmt.Fprintf(&builder, "Failed: %d\n", summary.Failed)
	fmt.Fprintf(&builder, "Skipped: %d\n\n", summary.Skipped)
	for _, result := range summary.Results {
		fmt.Fprintf(&builder, "- `%s`: %s (%d/%d)\n", result.TaskID, result.Verdict, result.Passed, result.Total)
		if result.Reason != "" {
			fmt.Fprintf(&builder, "  Reason: %s\n", result.Reason)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
