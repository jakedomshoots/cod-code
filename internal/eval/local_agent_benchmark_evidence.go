package eval

func writeBenchmarkSetupFailureEvidence(result LocalAgentBenchmarkResult) error {
	check := CheckResult{
		Name:     "setup:prepare_workspace",
		Status:   "fail",
		Evidence: result.WorkspaceDir,
		Message:  result.Error,
	}
	score := ScoreResult{
		TaskID:  result.TaskID,
		Verdict: "fail",
		Passed:  0,
		Total:   1,
		Checks:  []CheckResult{check},
	}
	report := map[string]any{
		"task_id":         result.TaskID,
		"agent_id":        result.ID,
		"status":          result.Status,
		"evidence_status": result.EvidenceStatus,
		"error":           result.Error,
		"note":            result.Note,
		"workspace_dir":   result.WorkspaceDir,
	}
	if err := writeJSONFile(result.CommandPath, result.Command); err != nil {
		return err
	}
	if err := writeTextFile(result.StdoutPath, "workspace setup failed before agent command\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.StderrPath, result.Error+"\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.DiffPath, "workspace setup failed before diff\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.ChangedFilesPath, "workspace setup failed before changed-file scan\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitBeforePath, "workspace setup failed before git status\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitAfterPath, "workspace setup failed before git status\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.TimingPath, "duration_ms=0\n"); err != nil {
		return err
	}
	if err := writeJSONFile(result.ReportPath, report); err != nil {
		return err
	}
	return writeJSONFile(result.ScorePath, score)
}

func failedScoreChecks(score ScoreResult) []CheckResult {
	failed := make([]CheckResult, 0)
	for _, check := range score.Checks {
		if check.Status != "pass" {
			failed = append(failed, check)
		}
	}
	return failed
}
