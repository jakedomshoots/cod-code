package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type CommandErrorKind string

const (
	CommandErrorKindFailed         CommandErrorKind = "command_failed"
	CommandErrorKindTimeout        CommandErrorKind = "command_timeout"
	CommandErrorKindOutputTooLarge CommandErrorKind = "command_output_too_large"
)

var ErrCommandFailed = errors.New("model command failed")

type CommandError struct {
	Kind     CommandErrorKind
	ExitCode int
	Stderr   string
}

func (e *CommandError) Error() string {
	switch e.Kind {
	case CommandErrorKindTimeout:
		return "model command timed out: " + e.Unwrap().Error()
	case CommandErrorKindOutputTooLarge:
		return "model command output too large"
	case CommandErrorKindFailed:
		return e.failedError()
	default:
		return "model command error"
	}
}

func (e *CommandError) Unwrap() error {
	switch e.Kind {
	case CommandErrorKindTimeout:
		return context.DeadlineExceeded
	case CommandErrorKindOutputTooLarge:
		return ErrCommandOutputTooLarge
	default:
		return ErrCommandFailed
	}
}

func (e *CommandError) failedError() string {
	stderr := strings.TrimSpace(e.Stderr)
	if stderr == "" {
		return fmt.Sprintf("model command failed with exit code %d", e.ExitCode)
	}
	return fmt.Sprintf("model command failed with exit code %d: %s", e.ExitCode, stderr)
}
