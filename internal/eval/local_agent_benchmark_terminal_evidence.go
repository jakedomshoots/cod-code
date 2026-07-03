package eval

import "fmt"

func writeBenchmarkTerminalEvidence(task Task, result LocalAgentBenchmarkResult, checkName string) error {
	message := result.Error
	if message == "" {
		message = result.Note
	}
	check := CheckResult{
		Name:     checkName,
		Status:   "fail",
		Evidence: result.ScorePath,
		Message:  message,
	}
	score := ScoreResult{
		TaskID:  task.ID,
		Verdict: "fail",
		Passed:  0,
		Total:   1,
		Checks:  []CheckResult{check},
	}
	report := map[string]any{
		"task_id":         task.ID,
		"agent_id":        result.ID,
		"status":          result.Status,
		"evidence_status": result.EvidenceStatus,
		"error":           result.Error,
		"note":            result.Note,
		"workspace_dir":   result.WorkspaceDir,
		"changed_files":   []string{},
		"check_results":   []commandResult{},
		"patch_results":   []patchResult{},
		"evidence_paths":  append([]string(nil), task.RequiredArtifacts...),
	}
	if err := writeJSONFile(result.CommandPath, map[string][]string{"command": result.Command}); err != nil {
		return err
	}
	if err := writeTextFile(result.StdoutPath, nonEmptyLog("")); err != nil {
		return err
	}
	if err := writeTextFile(result.StderrPath, nonEmptyLog(message)); err != nil {
		return err
	}
	if err := writeTextFile(result.DiffPath, "no diff captured for terminal benchmark result\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.ChangedFilesPath, "no changed files captured for terminal benchmark result\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitBeforePath, "no pre-run git status captured for terminal benchmark result\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitAfterPath, "no post-run git status captured for terminal benchmark result\n"); err != nil {
		return err
	}
	timing := fmt.Sprintf("duration_ms=%d\nexit_code=%d\n", result.DurationMS, result.ExitCode)
	if err := writeTextFile(result.TimingPath, timing); err != nil {
		return err
	}
	if err := writeJSONFile(result.ReportPath, report); err != nil {
		return err
	}
	return writeJSONFile(result.ScorePath, score)
}
