package eval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

var gitStatusCommand = []string{"git", "status", "--porcelain=v1", "--untracked-files=all"}

type worktreeStatusEvidence struct {
	Source  string   `json:"source"`
	Command []string `json:"command"`
	Payload string   `json:"payload"`
	SHA256  string   `json:"sha256"`
}

func appendDirtyWorktreeChecks(ctx context.Context, checks []CheckResult, req ScoreRequest, report savedReport) []CheckResult {
	if !req.Task.DirtyWorktreeSensitive {
		return checks
	}

	evidenceCheck := validateDirtyWorktreeEvidence(report.WorktreeStatus)
	checks = append(checks, evidenceCheck)
	if evidenceCheck.Status != "pass" {
		checks = append(checks, CheckResult{Name: "dirty_worktree_status", Status: "fail", Message: "valid git status evidence is required"})
		checks = append(checks, CheckResult{Name: "dirty_worktree_reported_files", Status: "fail", Message: "valid git status evidence is required"})
		return checks
	}

	actualPayload, err := captureGitStatus(ctx, req.WorkspaceDir)
	if err != nil {
		message := fmt.Sprintf("capture git status: %v", err)
		checks = append(checks, CheckResult{Name: "dirty_worktree_status", Status: "fail", Message: message})
		checks = append(checks, CheckResult{Name: "dirty_worktree_reported_files", Status: "fail", Message: message})
		return checks
	}
	if actualPayload != report.WorktreeStatus.Payload {
		checks = append(checks, CheckResult{Name: "dirty_worktree_status", Status: "fail", Message: "captured git status does not match current workspace"})
		checks = append(checks, CheckResult{Name: "dirty_worktree_reported_files", Status: "fail", Message: "current workspace status must match evidence before file coverage is scored"})
		return checks
	}

	checks = append(checks, CheckResult{Name: "dirty_worktree_status", Status: "pass", Evidence: report.WorktreeStatus.SHA256})
	checks = append(checks, dirtyWorktreeReportedFilesCheck(report))
	return checks
}

func validateDirtyWorktreeEvidence(evidence worktreeStatusEvidence) CheckResult {
	if evidence.Source == "" && len(evidence.Command) == 0 && evidence.Payload == "" && evidence.SHA256 == "" {
		return CheckResult{Name: "dirty_worktree_evidence", Status: "fail", Message: "missing git status evidence"}
	}
	if evidence.Source != "git" {
		return CheckResult{Name: "dirty_worktree_evidence", Status: "fail", Message: "source must be git"}
	}
	if !sameStrings(evidence.Command, gitStatusCommand) {
		return CheckResult{Name: "dirty_worktree_evidence", Status: "fail", Message: "command must capture porcelain git status"}
	}
	if evidence.SHA256 != sha256String(evidence.Payload) {
		return CheckResult{Name: "dirty_worktree_evidence", Status: "fail", Message: "payload hash mismatch"}
	}
	return CheckResult{Name: "dirty_worktree_evidence", Status: "pass", Evidence: evidence.SHA256}
}

func captureGitStatus(ctx context.Context, workspaceDir string) (string, error) {
	if strings.TrimSpace(workspaceDir) == "" {
		return "", fmt.Errorf("workspace is required")
	}
	cmd := exec.CommandContext(ctx, gitStatusCommand[0], gitStatusCommand[1:]...)
	cmd.Dir = workspaceDir
	output, err := cmd.Output()
	if err != nil {
		var detail string
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			detail = strings.TrimSpace(string(exitErr.Stderr))
		}
		if detail != "" {
			return "", fmt.Errorf("%w: %s", err, detail)
		}
		return "", err
	}
	return string(output), nil
}

func dirtyWorktreeReportedFilesCheck(report savedReport) CheckResult {
	missing := unreportedDirtyPaths(report.WorktreeStatus.Payload, report.ChangedFiles)
	if len(missing) > 0 {
		return CheckResult{
			Name:     "dirty_worktree_reported_files",
			Status:   "fail",
			Evidence: strings.Join(missing, "\n"),
			Message:  "dirty files missing from changed_files",
		}
	}
	return CheckResult{Name: "dirty_worktree_reported_files", Status: "pass"}
}

func unreportedDirtyPaths(payload string, changedFiles []string) []string {
	paths := dirtyPathsFromPorcelain(payload)
	missing := make([]string, 0)
	for _, path := range paths {
		if !stringInSlice(path, changedFiles) {
			missing = append(missing, path)
		}
	}
	sort.Strings(missing)
	return missing
}

func dirtyPathsFromPorcelain(payload string) []string {
	lines := strings.Split(payload, "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) < 4 {
			paths = append(paths, strings.TrimSpace(line))
			continue
		}
		path := strings.TrimSpace(line[3:])
		if renamed, ok := strings.CutPrefix(path, "-> "); ok {
			path = strings.TrimSpace(renamed)
		}
		if _, renamed, ok := strings.Cut(path, " -> "); ok {
			path = strings.TrimSpace(renamed)
		}
		paths = append(paths, strings.Trim(path, `"`))
	}
	return paths
}

func sameStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sha256String(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
