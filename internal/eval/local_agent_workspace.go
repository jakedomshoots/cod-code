package eval

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func prepareLocalAgentWorkspace(workspaceDir string) error {
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return fmt.Errorf("create local agent workspace: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceDir, "README.md"), []byte("# Local Agent Eval\n"), 0o644); err != nil {
		return fmt.Errorf("write local agent README: %w", err)
	}
	if err := os.WriteFile(filepath.Join(workspaceDir, "app.txt"), []byte("hello old\n"), 0o644); err != nil {
		return fmt.Errorf("write local agent app fixture: %w", err)
	}
	return nil
}

func resolveLocalAgentBinary(binary string) (string, error) {
	if strings.Contains(binary, string(os.PathSeparator)) {
		absolute, absErr := filepath.Abs(binary)
		if absErr != nil {
			return "", fmt.Errorf("resolve local binary %s: %w", binary, absErr)
		}
		info, err := os.Stat(absolute)
		if err != nil || info.IsDir() {
			return "", fmt.Errorf("local binary unavailable: %s", binary)
		}
		return absolute, nil
	}
	return exec.LookPath(binary)
}

func localAgentCommand(binary string, args []string, workspaceDir string) []string {
	if filepath.Base(binary) == "ceo-packet" {
		command := []string{binary, "--workspace", workspaceDir}
		return append(command, args...)
	}
	if filepath.Base(binary) == "opencode" {
		command := []string{binary}
		command = append(command, args[:1]...)
		command = append(command, "--dir", workspaceDir)
		return append(command, args[1:]...)
	}
	if filepath.Base(binary) == "codex" {
		command := []string{binary}
		command = append(command, args[:1]...)
		command = append(command, "-C", workspaceDir)
		return append(command, args[1:]...)
	}
	if filepath.Base(binary) == "omp" {
		return append([]string{binary, "--cwd", workspaceDir}, args...)
	}
	return append([]string{binary}, args...)
}
