package researchrunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const DefaultMaxOutputBytes = 8000

var (
	ErrCommandRequired = errors.New("research command is required")
	ErrQueryRequired   = errors.New("research query is required")
)

type Command struct {
	Argv      []string
	Env       []string
	Query     string
	MaxBytes  int
	TimeoutMS int
}

type Result struct {
	Status     string
	ExitCode   int
	DurationMS int64
	Output     string
	Error      string
	Bytes      int
	Truncated  bool
}

type Runner struct{}

func NewRunner() Runner {
	return Runner{}
}

func (r Runner) Run(ctx context.Context, command Command) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(command.Argv) == 0 {
		return Result{}, ErrCommandRequired
	}
	query := strings.TrimSpace(command.Query)
	if query == "" {
		return Result{}, ErrQueryRequired
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
	cmd.Env = append(cmd.Environ(), append(command.Env, "CEO_RESEARCH_QUERY="+query)...)
	cmd.Stdin = strings.NewReader(query)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startedAt := time.Now()
	err := cmd.Run()
	output, outputBytes, outputTruncated := boundedText(stdout.String(), command.MaxBytes)
	errorText, _, _ := boundedText(stderr.String(), command.MaxBytes)
	result := Result{
		DurationMS: time.Since(startedAt).Milliseconds(),
		Output:     output,
		Error:      errorText,
		Bytes:      outputBytes,
		Truncated:  outputTruncated,
	}
	if runCtx.Err() != nil {
		if ctx.Err() != nil {
			return Result{}, fmt.Errorf("run research command: %w", runCtx.Err())
		}
		result.Status = "fail"
		result.ExitCode = -1
		result.Error = runCtx.Err().Error()
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
	return Result{}, fmt.Errorf("run research command: %w", err)
}

func boundedText(text string, maxBytes int) (string, int, bool) {
	limit := maxBytes
	if limit < 1 {
		limit = DefaultMaxOutputBytes
	}
	if len(text) <= limit {
		return text, len(text), false
	}
	end := 0
	for index := range text {
		if index > limit {
			break
		}
		end = index
	}
	if end == 0 {
		return "", len(text), true
	}
	return text[:end], len(text), true
}
