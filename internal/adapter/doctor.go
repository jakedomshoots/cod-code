package adapter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultDoctorTimeout = 2 * time.Second
	versionProbeAttempts = 4
)

func DoctorAll(ctx context.Context, opts DoctorOptions) []Report {
	tools := SupportedTools()
	reports := make([]Report, 0, len(tools))
	for _, tool := range tools {
		reports = append(reports, Doctor(ctx, tool, opts))
	}
	return reports
}

func Doctor(ctx context.Context, tool Tool, opts DoctorOptions) Report {
	command := strings.TrimSpace(os.Getenv(tool.EnvVar))
	if command == "" {
		return skipReport(tool)
	}
	version, versionOK := probeVersion(ctx, tool, command, opts)
	parsed, err := probeDryRun(ctx, tool, command, opts)
	if err != nil {
		return failedReport(tool, version, versionOK, err)
	}
	return Report{
		Tool:        string(tool.ID),
		DisplayName: tool.DisplayName,
		EnvVar:      tool.EnvVar,
		Status:      StatusPass,
		Version:     version,
		Summary:     parsed.Summary,
		PatchCount:  parsed.PatchCount,
		Capabilities: Capabilities{
			VersionCheck:   versionOK,
			DryRun:         true,
			OutputParser:   true,
			Timeout:        true,
			ProviderHealth: true,
		},
		Health: Health{Status: HealthPass},
	}
}

func skipReport(tool Tool) Report {
	return Report{
		Tool:        string(tool.ID),
		DisplayName: tool.DisplayName,
		EnvVar:      tool.EnvVar,
		Status:      StatusSkip,
		Capabilities: Capabilities{
			Timeout: true,
		},
		Health:     Health{Status: HealthSkip, ErrorKind: string(ErrorKindMissingSetup), Error: "adapter command is not configured"},
		ErrorKind:  string(ErrorKindMissingSetup),
		Error:      "adapter command is not configured",
		SetupSteps: setupSteps(tool),
		Err:        &Error{Tool: tool.ID, Kind: ErrorKindMissingSetup, Err: ErrMissingSetup},
	}
}

func failedReport(tool Tool, version string, versionOK bool, err error) Report {
	kind := errorKind(err)
	return Report{
		Tool:        string(tool.ID),
		DisplayName: tool.DisplayName,
		EnvVar:      tool.EnvVar,
		Status:      StatusFail,
		Version:     version,
		Capabilities: Capabilities{
			VersionCheck:   versionOK,
			Timeout:        true,
			ProviderHealth: true,
		},
		Health:    Health{Status: HealthFail, ErrorKind: string(kind), Error: err.Error()},
		ErrorKind: string(kind),
		Error:     err.Error(),
		Err:       err,
	}
}

func setupSteps(tool Tool) []string {
	return []string{
		"export " + tool.EnvVar + "=/path/to/" + string(tool.ID) + "-adapter",
		"read " + tool.SetupDoc,
	}
}

func probeVersion(ctx context.Context, tool Tool, command string, opts DoctorOptions) (string, bool) {
	for attempt := 0; attempt < versionProbeAttempts; attempt++ {
		stdout, err := runProbe(ctx, command, opts, "version", "")
		if err != nil {
			if errors.Is(err, ErrTimeout) {
				continue
			}
			return "", false
		}
		version := strings.TrimSpace(stdout)
		if version == "" {
			return "", false
		}
		return version, strings.Contains(strings.ToLower(version), string(tool.ID)) || version != ""
	}
	return "", false
}

func probeDryRun(ctx context.Context, tool Tool, command string, opts DoctorOptions) (ParsedOutput, error) {
	stdout, err := runProbe(ctx, command, opts, "dry-run", "role: adapter doctor\nassignment: return one structured JSON patch proposal\n")
	if err != nil {
		return ParsedOutput{}, withTool(tool, err)
	}
	parsed, err := ParseOutput(stdout)
	if err != nil {
		return ParsedOutput{}, withTool(tool, err)
	}
	return parsed, nil
}

func runProbe(ctx context.Context, command string, opts DoctorOptions, probe string, stdin string) (string, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultDoctorTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := probeCommand(runCtx, command)
	isolateProbeProcess(cmd)
	cmd.Env = probeEnv(probe)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return "", &Error{Kind: ErrorKindTimeout, Err: ErrTimeout}
		}
		return "", &Error{Kind: ErrorKindCommandFailed, Err: fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))}
	}
	return stdout.String(), nil
}

func probeEnv(probe string) []string {
	env := os.Environ()
	clean := make([]string, 0, len(env)+1)
	for _, item := range env {
		if strings.HasPrefix(item, "CEO_HARNESS_ADAPTER_PROBE=") {
			continue
		}
		clean = append(clean, item)
	}
	return append(clean, "CEO_HARNESS_ADAPTER_PROBE="+probe)
}

func probeCommand(ctx context.Context, command string) *exec.Cmd {
	fields := strings.Fields(command)
	if len(fields) == 1 {
		return exec.CommandContext(ctx, fields[0])
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
}

func errorKind(err error) ErrorKind {
	var adapterErr *Error
	if errors.As(err, &adapterErr) {
		return adapterErr.Kind
	}
	return ErrorKindCommandFailed
}

func withTool(tool Tool, err error) error {
	var adapterErr *Error
	if errors.As(err, &adapterErr) {
		return &Error{Tool: tool.ID, Kind: adapterErr.Kind, Err: adapterErr.Err}
	}
	return &Error{Tool: tool.ID, Kind: ErrorKindCommandFailed, Err: err}
}
