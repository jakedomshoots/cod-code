package adapter

import (
	"errors"
	"fmt"
	"time"
)

type ToolID string

const (
	ToolCodex    ToolID = "codex"
	ToolClaude   ToolID = "claude"
	ToolOpenCode ToolID = "opencode"
	ToolAider    ToolID = "aider"
	ToolGoose    ToolID = "goose"
)

type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

type HealthStatus string

const (
	HealthPass HealthStatus = "pass"
	HealthFail HealthStatus = "fail"
	HealthSkip HealthStatus = "skip"
)

type ErrorKind string

const (
	ErrorKindMissingSetup  ErrorKind = "missing_setup"
	ErrorKindCommandFailed ErrorKind = "command_failed"
	ErrorKindTimeout       ErrorKind = "timeout"
	ErrorKindInvalidOutput ErrorKind = "invalid_output"
)

var (
	ErrMissingSetup  = errors.New("adapter missing setup")
	ErrInvalidOutput = errors.New("adapter invalid output")
	ErrTimeout       = errors.New("adapter timed out")
)

type Error struct {
	Tool ToolID
	Kind ErrorKind
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s adapter %s: %v", e.Tool, e.Kind, e.Err)
}

func (e *Error) Unwrap() error {
	switch e.Kind {
	case ErrorKindMissingSetup:
		return ErrMissingSetup
	case ErrorKindInvalidOutput:
		return ErrInvalidOutput
	case ErrorKindTimeout:
		return ErrTimeout
	default:
		return e.Err
	}
}

type Tool struct {
	ID          ToolID `json:"id"`
	DisplayName string `json:"display_name"`
	EnvVar      string `json:"env_var"`
	SetupDoc    string `json:"setup_doc"`
}

type Capabilities struct {
	VersionCheck   bool `json:"version_check"`
	DryRun         bool `json:"dry_run"`
	OutputParser   bool `json:"output_parser"`
	Timeout        bool `json:"timeout"`
	ProviderHealth bool `json:"provider_health"`
}

type Health struct {
	Status    HealthStatus `json:"status"`
	ErrorKind string       `json:"error_kind,omitempty"`
	Error     string       `json:"error,omitempty"`
}

type Report struct {
	Tool         string       `json:"tool"`
	DisplayName  string       `json:"display_name"`
	EnvVar       string       `json:"env_var"`
	Status       Status       `json:"status"`
	Version      string       `json:"version,omitempty"`
	Summary      string       `json:"summary,omitempty"`
	PatchCount   int          `json:"patch_count"`
	Capabilities Capabilities `json:"capabilities"`
	Health       Health       `json:"provider_health"`
	ErrorKind    string       `json:"error_kind,omitempty"`
	Error        string       `json:"error,omitempty"`
	SetupSteps   []string     `json:"setup_steps,omitempty"`
	Err          error        `json:"-"`
}

type DoctorOptions struct {
	Timeout time.Duration
}
