package checkrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

var ErrEmptyCommand = errors.New("check command is required")

type Command struct {
	Argv      []string
	Env       []string
	WorkDir   string
	TimeoutMS int
}

type Result struct {
	Argv        []string `json:"argv"`
	Status      string   `json:"status"`
	ExitCode    int      `json:"exit_code"`
	CheckIndex  int      `json:"check_index,omitempty"`
	Attempt     int      `json:"attempt,omitempty"`
	MaxAttempts int      `json:"max_attempts,omitempty"`
	DurationMS  int64    `json:"duration_ms"`
	Stdout      string   `json:"stdout"`
	Stderr      string   `json:"stderr"`
}

type Runner struct{}

func NewRunner() Runner {
	return Runner{}
}

func (r Runner) Run(ctx context.Context, command Command) (Result, error) {
	if len(command.Argv) == 0 {
		return Result{}, ErrEmptyCommand
	}
	runCtx := ctx
	cancel := func() {}
	if command.TimeoutMS > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(command.TimeoutMS)*time.Millisecond)
	}
	defer cancel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(runCtx, command.Argv[0], command.Argv[1:]...)
	configureCommandCancellation(cmd)
	cmd.Env = append(cmd.Environ(), command.Env...)
	cmd.Dir = command.WorkDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startedAt := time.Now()
	err := cmd.Run()
	result := Result{
		Argv:       append([]string(nil), command.Argv...),
		DurationMS: time.Since(startedAt).Milliseconds(),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
	}
	if runCtx.Err() != nil {
		if ctx.Err() != nil {
			return Result{}, fmt.Errorf("run check command: %w", runCtx.Err())
		}
		result.Status = "fail"
		result.ExitCode = -1
		result.Stderr = runCtx.Err().Error()
		return result, nil
	}
	if err == nil {
		result.Status = "pass"
		return result, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.Status = "fail"
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}
	return Result{}, fmt.Errorf("run check command: %w", err)
}
