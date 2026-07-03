package eval

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func prepareBenchmarkWorkspace(ctx context.Context, taskDir string, task Task) (string, worktreeStatusEvidence, error) {
	if !task.DirtyWorktreeSensitive {
		return taskDir, worktreeStatusEvidence{}, nil
	}
	workspaceDir := filepath.Join(taskDir, "workspace")
	if err := os.RemoveAll(workspaceDir); err != nil {
		return "", worktreeStatusEvidence{}, fmt.Errorf("reset workspace: %w", err)
	}
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return "", worktreeStatusEvidence{}, fmt.Errorf("create workspace: %w", err)
	}
	if err := runGitCommand(ctx, workspaceDir, "init"); err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	if err := runGitCommand(ctx, workspaceDir, "config", "user.email", "eval@example.com"); err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	if err := runGitCommand(ctx, workspaceDir, "config", "user.name", "Eval Fixture"); err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	for _, path := range task.RequiredChangedFiles {
		if err := writeRelativeFile(workspaceDir, path, "baseline\n"); err != nil {
			return "", worktreeStatusEvidence{}, err
		}
	}
	if err := runGitCommand(ctx, workspaceDir, "add", "."); err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	if err := runGitCommand(ctx, workspaceDir, "commit", "-m", "baseline"); err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	for _, path := range task.RequiredChangedFiles {
		if err := writeRelativeFile(workspaceDir, path, strings.Join(task.RequiredDiffTerms, "\n")+"\n"); err != nil {
			return "", worktreeStatusEvidence{}, err
		}
	}
	payload, err := captureGitStatus(ctx, workspaceDir)
	if err != nil {
		return "", worktreeStatusEvidence{}, err
	}
	return workspaceDir, worktreeStatusEvidence{
		Source:  "git",
		Command: append([]string(nil), gitStatusCommand...),
		Payload: payload,
		SHA256:  sha256String(payload),
	}, nil
}

func runGitCommand(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}
