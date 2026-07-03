package ceo

import (
	"context"
	"fmt"
	"time"

	"ceoharness/internal/checkrunner"
)

func (r Runtime) runChecks(ctx context.Context, req JobRequest) ([]checkrunner.Result, error) {
	commands := checkCommands(req)
	if len(commands) == 0 {
		return nil, nil
	}
	results := []checkrunner.Result{}
	attempts := req.CheckAttempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(req.CheckBackoffMS) * time.Millisecond
	for commandIndex, command := range commands {
		finalResult := checkrunner.Result{}
		for attempt := 0; attempt < attempts; attempt++ {
			checkResult, err := r.checks.Run(ctx, checkrunner.Command{
				Argv:      command,
				Env:       req.CheckEnv,
				WorkDir:   req.WorkspaceDir,
				TimeoutMS: req.ToolCommandTimeoutMS,
			})
			if err != nil {
				return nil, fmt.Errorf("run check command: %w", err)
			}
			checkResult.CheckIndex = commandIndex + 1
			checkResult.Attempt = attempt + 1
			checkResult.MaxAttempts = attempts
			results = append(results, checkResult)
			finalResult = checkResult
			if checkResult.Status == "pass" {
				break
			}
			if attempt+1 < attempts {
				if err := waitForRetryBackoff(ctx, backoff); err != nil {
					return nil, fmt.Errorf("run check command: %w", err)
				}
			}
		}
		if finalResult.Status != "pass" {
			break
		}
	}
	return results, nil
}

func checkCommands(req JobRequest) [][]string {
	if len(req.CheckCommands) > 0 {
		return cloneCheckCommands(req.CheckCommands)
	}
	if len(req.CheckCommand) == 0 {
		return nil
	}
	return [][]string{append([]string(nil), req.CheckCommand...)}
}

func cloneCheckCommands(commands [][]string) [][]string {
	copied := make([][]string, 0, len(commands))
	for _, command := range commands {
		copied = append(copied, append([]string(nil), command...))
	}
	return copied
}
