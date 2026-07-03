package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"ceoharness/internal/processcancel"
)

var (
	ErrCommandRequired       = errors.New("model command is required")
	ErrCommandOutputTooLarge = errors.New("model command output too large")
)

const defaultCommandMaxOutputBytes = 1 << 20

type CommandSpec struct {
	Argv           []string
	Env            []string
	TimeoutMS      int
	MaxOutputBytes int
}

type CommandClient struct {
	argv           []string
	env            []string
	timeout        time.Duration
	maxOutputBytes int
}

func NewCommandClient(spec CommandSpec) (CommandClient, error) {
	if len(spec.Argv) == 0 {
		return CommandClient{}, ErrCommandRequired
	}
	return CommandClient{
		argv:           append([]string(nil), spec.Argv...),
		env:            append([]string(nil), spec.Env...),
		timeout:        commandTimeout(spec.TimeoutMS),
		maxOutputBytes: commandMaxOutputBytes(spec.MaxOutputBytes),
	}, nil
}

func (c CommandClient) Complete(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	if strings.TrimSpace(req.Prompt) == "" {
		return Response{}, ErrPromptRequired
	}
	runCtx := ctx
	cancel := func() {}
	if c.timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, c.timeout)
	}
	defer cancel()

	stdout := newLimitedBuffer(c.maxOutputBytes)
	stderr := newLimitedBuffer(c.maxOutputBytes)
	cmd := exec.CommandContext(runCtx, c.argv[0], c.argv[1:]...)
	processcancel.ConfigureProcessTreeCancellation(cmd)
	env := append([]string(nil), c.env...)
	env = append(env, metadataEnvPairs(req.Metadata)...)
	cmd.Env = append(cmd.Environ(), env...)
	cmd.Stdin = strings.NewReader(req.Prompt)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		if runCtx.Err() != nil {
			return Response{}, fmt.Errorf("run model command: %w", &CommandError{Kind: CommandErrorKindTimeout})
		}
		return Response{}, fmt.Errorf("run model command: %w", commandFailedError(err, stderr.String()))
	}
	if stdout.Truncated() {
		return Response{}, fmt.Errorf("run model command: %w", &CommandError{Kind: CommandErrorKindOutputTooLarge})
	}
	return Response{
		Text:        stdout.String(),
		PromptBytes: len(req.Prompt),
	}, nil
}

func commandFailedError(err error, stderr string) *CommandError {
	exitCode := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	return &CommandError{
		Kind:     CommandErrorKindFailed,
		ExitCode: exitCode,
		Stderr:   stderr,
	}
}

func commandTimeout(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		return 0
	}
	return time.Duration(timeoutMS) * time.Millisecond
}

func commandMaxOutputBytes(maxOutputBytes int) int {
	if maxOutputBytes <= 0 {
		return defaultCommandMaxOutputBytes
	}
	return maxOutputBytes
}

type limitedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		_, _ = b.buffer.Write(p)
		return len(p), nil
	}
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buffer.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	text := b.buffer.String()
	if b.truncated {
		return text + "\n[truncated]"
	}
	return text
}

func (b *limitedBuffer) Truncated() bool {
	return b.truncated
}
