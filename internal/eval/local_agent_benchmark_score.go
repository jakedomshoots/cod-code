package eval

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func scoreLocalAgentBenchmark(ctx context.Context, task Task, result LocalAgentBenchmarkResult) (ScoreResult, []string, error) {
	checks := runBenchmarkRequiredCommands(ctx, result.WorkspaceDir, task, localAgentTimeout(0))
	changedFiles, statusAfter, err := benchmarkChangedFiles(ctx, result.WorkspaceDir)
	if err != nil {
		return ScoreResult{}, nil, err
	}
	if err := writeTextFile(result.GitAfterPath, statusAfter); err != nil {
		return ScoreResult{}, nil, err
	}
	if err := writeTextFile(result.ChangedFilesPath, strings.Join(changedFiles, "\n")+"\n"); err != nil {
		return ScoreResult{}, nil, err
	}
	diff, patches, err := benchmarkPatchResults(ctx, result.WorkspaceDir, task)
	if err != nil {
		return ScoreResult{}, nil, err
	}
	if err := writeTextFile(result.DiffPath, nonEmptyLog(diff)); err != nil {
		return ScoreResult{}, nil, err
	}
	report, err := localAgentBenchmarkReport(task, changedFiles, checks, patches, statusAfter)
	if err != nil {
		return ScoreResult{}, nil, err
	}
	if err := writeJSONFile(result.ReportPath, report); err != nil {
		return ScoreResult{}, nil, err
	}
	score, err := ScoreReport(ctx, ScoreRequest{
		Task:         task,
		ReportPath:   result.ReportPath,
		WorkspaceDir: result.WorkspaceDir,
	})
	if err != nil {
		return ScoreResult{}, nil, err
	}
	if err := writeJSONFile(result.ScorePath, score); err != nil {
		return ScoreResult{}, nil, err
	}
	return score, changedFiles, nil
}

func runBenchmarkRequiredCommands(ctx context.Context, workspaceDir string, task Task, timeout time.Duration) []commandResult {
	results := make([]commandResult, 0, len(task.RequiredCommands))
	for _, command := range task.RequiredCommands {
		argv := strings.Fields(command)
		run := runLocalAgentCommand(ctx, argv, workspaceDir, nil, timeout)
		status := "fail"
		if run.exitCode == 0 && run.errText == "" {
			status = "pass"
		}
		results = append(results, commandResult{
			Argv:     argv,
			Status:   status,
			ExitCode: run.exitCode,
			Stdout:   run.stdout,
			Stderr:   run.stderr,
		})
	}
	return results
}

func benchmarkPatchResults(ctx context.Context, workspaceDir string, task Task) (string, []patchResult, error) {
	diff, err := gitDiff(ctx, workspaceDir)
	if err != nil {
		return "", nil, err
	}
	results := make([]patchResult, 0, len(task.RequiredChangedFiles))
	for _, path := range task.RequiredChangedFiles {
		fileDiff, err := gitDiffForPath(ctx, workspaceDir, path)
		if err != nil {
			return "", nil, err
		}
		results = append(results, patchResult{Path: path, Diff: fileDiff})
	}
	return diff, results, nil
}

func gitDiff(ctx context.Context, workspaceDir string) (string, error) {
	return gitOutput(ctx, workspaceDir, "diff", "--no-ext-diff", "HEAD", "--")
}

func gitDiffForPath(ctx context.Context, workspaceDir string, path string) (string, error) {
	clean, ok := cleanRelativeArtifactPath(path)
	if !ok {
		return "", fmt.Errorf("invalid diff path %q", path)
	}
	return gitOutput(ctx, workspaceDir, "diff", "--no-ext-diff", "HEAD", "--", clean)
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func localAgentBenchmarkReport(task Task, changedFiles []string, checks []commandResult, patches []patchResult, status string) (map[string]any, error) {
	report := map[string]any{
		"task_id":        task.ID,
		"changed_files":  append([]string(nil), changedFiles...),
		"check_results":  checks,
		"patch_results":  patches,
		"evidence_paths": append([]string(nil), task.RequiredArtifacts...),
	}
	if task.DirtyWorktreeSensitive {
		report["worktree_status"] = worktreeStatusEvidence{
			Source:  "git",
			Command: append([]string(nil), gitStatusCommand...),
			Payload: status,
			SHA256:  sha256String(status),
		}
	}
	for _, field := range task.RequiredReportFields {
		setSyntheticReportField(report, field)
	}
	return report, nil
}

func benchmarkExtraChangedFiles(task Task, changedFiles []string) []string {
	allowed := make(map[string]struct{}, len(task.RequiredChangedFiles)+len(task.RequiredArtifacts))
	for _, path := range task.RequiredChangedFiles {
		allowed[path] = struct{}{}
	}
	for _, path := range task.RequiredArtifacts {
		allowed[path] = struct{}{}
	}
	extra := make([]string, 0)
	for _, path := range changedFiles {
		if _, ok := allowed[path]; !ok {
			extra = append(extra, path)
		}
	}
	return extra
}

func localAgentBenchmarkStatus(run localAgentRunResult, verdict string) string {
	if localAgentSetupBlocked(run) {
		return localAgentStatusSetupBlocked
	}
	if run.timedOut {
		return localAgentStatusTimeout
	}
	if run.exitCode != 0 || run.errText != "" {
		return localAgentStatusFail
	}
	switch verdict {
	case "pass":
		return localAgentStatusPass
	case "partial":
		return localAgentStatusPartial
	default:
		return localAgentStatusFail
	}
}

func localAgentBenchmarkNote(status string) string {
	switch status {
	case localAgentStatusPass:
		return "agent command exited 0 and scored pass on saved benchmark evidence"
	case localAgentStatusPartial:
		return "agent command exited 0 but only partially satisfied benchmark scoring"
	case localAgentStatusTimeout:
		return "agent command timed out; process tree was canceled"
	case localAgentStatusSetupBlocked:
		return "agent provider setup is blocked; stdout/stderr captured the auth, quota, or credential issue"
	default:
		return "agent command or benchmark score failed"
	}
}

func localAgentSetupBlocked(run localAgentRunResult) bool {
	body := strings.ToLower(run.stdout + "\n" + run.stderr + "\n" + run.errText)
	for _, marker := range []string{
		"token plan usage limit reached",
		"token refresh failed",
		"api key appears to be invalid",
		"api key appears to be invalid or may have expired",
		"authentication_error",
		"quota",
		"usage limit",
	} {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}
