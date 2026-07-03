package eval

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ceoharness/internal/processcancel"
)

type localAgentRunResult struct {
	stdout   string
	stderr   string
	errText  string
	exitCode int
	duration time.Duration
	timedOut bool
}

func localAgentTimeout(timeoutSeconds int) time.Duration {
	if timeoutSeconds <= 0 {
		return defaultLocalAgentTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func runLocalAgentCommand(ctx context.Context, command []string, dir string, env []string, timeout time.Duration) localAgentRunResult {
	started := time.Now()
	if len(command) == 0 {
		return localAgentRunResult{
			errText:  "empty command",
			exitCode: -1,
			duration: time.Since(started),
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, command[0], command[1:]...)
	processcancel.ConfigureProcessTreeCancellation(cmd)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := localAgentRunResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: commandExitCode(err),
		duration: time.Since(started),
		timedOut: errors.Is(runCtx.Err(), context.DeadlineExceeded),
	}
	if err != nil {
		result.errText = err.Error()
		if result.timedOut {
			result.errText = "command timed out"
		}
	}
	return result
}

func localAgentStatus(run localAgentRunResult, outputMatched bool, fileMatched bool) string {
	if run.timedOut {
		return localAgentStatusTimeout
	}
	if run.exitCode == 0 && run.errText == "" && outputMatched && fileMatched {
		return localAgentStatusPass
	}
	return localAgentStatusFail
}

func localAgentNote(status string, expectedOutput string) string {
	switch status {
	case localAgentStatusPass:
		return "non-interactive command exited 0 and matched expected evidence"
	case localAgentStatusTimeout:
		return "non-interactive command timed out; process tree was canceled"
	default:
		if expectedOutput == "" {
			return "non-interactive command did not produce the expected file state"
		}
		return fmt.Sprintf("non-interactive command did not produce expected output %q", expectedOutput)
	}
}

func outputMatches(output string, expectedOutput string) bool {
	if expectedOutput == "" {
		return true
	}
	return bytes.Contains([]byte(output), []byte(expectedOutput))
}

func observedFileMatch(workspaceDir string, expectedFile string) (string, bool) {
	if expectedFile == "" {
		return "", true
	}
	content, err := os.ReadFile(filepath.Join(workspaceDir, "app.txt"))
	if err != nil {
		return "", false
	}
	observed := string(content)
	return observed, observed == expectedFile
}

func writeLocalAgentEvidence(result LocalAgentResult, run localAgentRunResult) error {
	if err := writeJSONFile(result.CommandPath, map[string][]string{"command": result.Command}); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(result.StdoutPath), 0o755); err != nil {
		return fmt.Errorf("create local agent evidence dir: %w", err)
	}
	if err := os.WriteFile(result.StdoutPath, []byte(nonEmptyLog(run.stdout)), 0o644); err != nil {
		return fmt.Errorf("write local agent stdout: %w", err)
	}
	if err := os.WriteFile(result.StderrPath, []byte(nonEmptyLog(run.stderr)), 0o644); err != nil {
		return fmt.Errorf("write local agent stderr: %w", err)
	}
	if result.AppAfterPath != "" {
		if err := os.WriteFile(result.AppAfterPath, []byte(nonEmptyLog(result.ObservedFile)), 0o644); err != nil {
			return fmt.Errorf("write local agent app state: %w", err)
		}
	}
	return nil
}

func nonEmptyLog(content string) string {
	if content == "" {
		return "(empty)\n"
	}
	return content
}
