package eval

import (
	"fmt"
	"path/filepath"
)

func retryAttemptFromResult(result LocalAgentBenchmarkResult) RetryAttempt {
	return RetryAttempt{
		RunAttempt:     result.RunAttempt,
		Status:         result.Status,
		EvidenceStatus: result.EvidenceStatus,
		ScorePath:      result.ScorePath,
		StdoutPath:     result.StdoutPath,
		StderrPath:     result.StderrPath,
		TimingPath:     result.TimingPath,
		Note:           result.Note,
	}
}

func newLocalAgentBenchmarkResult(spec localAgentSpec, task Task, attempt int, workspaceDir string, resultDir string) LocalAgentBenchmarkResult {
	return LocalAgentBenchmarkResult{
		ID:               spec.id,
		Name:             spec.name,
		TaskID:           task.ID,
		TaskTitle:        task.Title,
		Attempt:          attempt,
		Status:           localAgentStatusSkipped,
		Binary:           spec.binary,
		WorkspaceDir:     workspaceDir,
		ExitCode:         -1,
		CommandPath:      filepath.Join(resultDir, "command.json"),
		StdoutPath:       filepath.Join(resultDir, "stdout.log"),
		StderrPath:       filepath.Join(resultDir, "stderr.log"),
		ReportPath:       filepath.Join(resultDir, "report.json"),
		ScorePath:        filepath.Join(resultDir, "score.json"),
		DiffPath:         filepath.Join(resultDir, "diff.patch"),
		ChangedFilesPath: filepath.Join(resultDir, "changed-files.txt"),
		GitBeforePath:    filepath.Join(resultDir, "git-status-before.txt"),
		GitAfterPath:     filepath.Join(resultDir, "git-status-after.txt"),
		TimingPath:       filepath.Join(resultDir, "timing.txt"),
		EvidenceStatus:   localAgentEvidenceNotRun,
		SetupHint:        spec.setupHint,
		Note:             "binary not found on PATH; skipped instead of failed",
	}
}

func benchmarkResultError(task Task, result LocalAgentBenchmarkResult, note string, err error) LocalAgentBenchmarkResult {
	result.Status = localAgentStatusFail
	result.Error = err.Error()
	result.Note = note
	result.EvidenceStatus = localAgentEvidenceIncomplete
	if writeErr := writeBenchmarkTerminalEvidence(task, result, note); writeErr != nil {
		result.Error = fmt.Sprintf("%s; evidence write failed: %v", result.Error, writeErr)
	}
	return result
}
