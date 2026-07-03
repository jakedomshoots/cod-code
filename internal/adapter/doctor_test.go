package adapter

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	contractProbeTimeout     = 10 * time.Second
	timeoutRetryProbeTimeout = 2 * time.Second
)

func Test_DoctorAll_reports_skip_setup_when_adapter_commands_are_missing(t *testing.T) {
	// Given
	for _, tool := range SupportedTools() {
		t.Setenv(tool.EnvVar, "")
	}
	t.Setenv("PATH", t.TempDir())

	// When
	reports := DoctorAll(context.Background(), DoctorOptions{Timeout: 50 * time.Millisecond})

	// Then
	if len(reports) != 5 {
		t.Fatalf("reports length = %d, want five adapter reports", len(reports))
	}
	for _, report := range reports {
		if report.Status != StatusSkip || report.Health.Status != HealthSkip {
			t.Fatalf("%s status = %s/%s, want skip/skip", report.Tool, report.Status, report.Health.Status)
		}
		if report.ErrorKind != string(ErrorKindMissingSetup) {
			t.Fatalf("%s error kind = %q, want missing_setup", report.Tool, report.ErrorKind)
		}
		if len(report.SetupSteps) == 0 || !strings.Contains(report.SetupSteps[0], report.EnvVar) {
			t.Fatalf("%s setup steps = %#v, want env var guidance", report.Tool, report.SetupSteps)
		}
	}
}

func Test_DoctorAll_reports_capabilities_when_fake_adapters_follow_contract(t *testing.T) {
	// Given
	dir := t.TempDir()
	for _, tool := range SupportedTools() {
		path := writeFakeAdapter(t, dir, string(tool.ID), "valid")
		t.Setenv(tool.EnvVar, path)
	}

	// When
	reports := DoctorAll(context.Background(), DoctorOptions{Timeout: contractProbeTimeout})

	// Then
	if len(reports) != 5 {
		t.Fatalf("reports length = %d, want five adapter reports", len(reports))
	}
	for _, report := range reports {
		if report.Status != StatusPass || report.Health.Status != HealthPass {
			t.Fatalf("%s status = %s/%s error_kind=%s error=%s, want pass/pass", report.Tool, report.Status, report.Health.Status, report.ErrorKind, report.Error)
		}
		if !report.Capabilities.VersionCheck || !report.Capabilities.DryRun || !report.Capabilities.OutputParser || !report.Capabilities.Timeout || !report.Capabilities.ProviderHealth {
			t.Fatalf("%s capabilities = %+v, want all contract capabilities", report.Tool, report.Capabilities)
		}
		if report.Version == "" || report.PatchCount != 1 {
			t.Fatalf("%s version/patch count = %q/%d, want version and one patch", report.Tool, report.Version, report.PatchCount)
		}
	}
}

func Test_Doctor_reports_valid_codex_patch_proposal_when_fake_adapter_returns_structured_output(t *testing.T) {
	// Given
	dir := t.TempDir()
	tool, ok := ToolByID("codex")
	if !ok {
		t.Fatal("codex tool must be registered")
	}
	t.Setenv(tool.EnvVar, writeFakeAdapter(t, dir, "codex", "valid"))

	// When
	report := Doctor(context.Background(), tool, DoctorOptions{Timeout: contractProbeTimeout})

	// Then
	if report.Status != StatusPass {
		t.Fatalf("status = %s error=%s, want pass", report.Status, report.Error)
	}
	if report.Summary != "codex patch ready" || report.PatchCount != 1 {
		t.Fatalf("summary/patches = %q/%d, want parsed codex patch proposal", report.Summary, report.PatchCount)
	}
}

func Test_Doctor_records_typed_health_error_when_claude_fake_adapter_emits_invalid_output(t *testing.T) {
	// Given
	dir := t.TempDir()
	tool, ok := ToolByID("claude")
	if !ok {
		t.Fatal("claude tool must be registered")
	}
	t.Setenv(tool.EnvVar, writeFakeAdapter(t, dir, "claude", "invalid"))

	// When
	report := Doctor(context.Background(), tool, DoctorOptions{Timeout: contractProbeTimeout})

	// Then
	if report.Status != StatusFail || report.Health.Status != HealthFail {
		t.Fatalf("status = %s/%s, want fail/fail", report.Status, report.Health.Status)
	}
	if report.ErrorKind != string(ErrorKindInvalidOutput) || report.Health.ErrorKind != string(ErrorKindInvalidOutput) {
		t.Fatalf("error kind = %q/%q, want invalid_output", report.ErrorKind, report.Health.ErrorKind)
	}
	var adapterErr *Error
	if !errors.As(report.Err, &adapterErr) {
		t.Fatalf("Err = %T, want *Error", report.Err)
	}
	if adapterErr.Kind != ErrorKindInvalidOutput {
		t.Fatalf("typed error kind = %q, want invalid_output", adapterErr.Kind)
	}
}

func Test_Doctor_retries_transient_version_timeout_before_reporting_capability(t *testing.T) {
	// Given
	dir := t.TempDir()
	tool, ok := ToolByID("codex")
	if !ok {
		t.Fatal("codex tool must be registered")
	}
	t.Setenv(tool.EnvVar, writeTransientVersionAdapter(t, dir, "codex"))

	// When
	report := Doctor(context.Background(), tool, DoctorOptions{Timeout: timeoutRetryProbeTimeout})

	// Then
	if report.Status != StatusPass {
		t.Fatalf("status = %s error=%s, want pass", report.Status, report.Error)
	}
	if !report.Capabilities.VersionCheck || report.Version == "" {
		t.Fatalf("version capability/version = %v/%q, want retry-backed version check", report.Capabilities.VersionCheck, report.Version)
	}
}

func Test_Doctor_retries_version_timeout_without_waiting_for_child_process(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("shell fake uses POSIX process behavior")
	}
	dir := t.TempDir()
	childPIDPath := filepath.Join(dir, "version-child.pid")
	tool, ok := ToolByID("codex")
	if !ok {
		t.Fatal("codex tool must be registered")
	}
	t.Setenv(tool.EnvVar, writeVersionTimeoutChildAdapter(t, dir, "codex", childPIDPath))

	// When
	report := Doctor(context.Background(), tool, DoctorOptions{Timeout: timeoutRetryProbeTimeout})

	// Then
	if report.Status != StatusPass {
		t.Fatalf("status = %s error=%s, want pass", report.Status, report.Error)
	}
	if !report.Capabilities.VersionCheck || report.Version == "" {
		t.Fatalf("version capability/version = %v/%q, want retry-backed version check", report.Capabilities.VersionCheck, report.Version)
	}
	childPID := readChildPID(t, childPIDPath)
	requireProcessExited(t, childPID, 5*time.Second)
}

func Test_Doctor_records_timeout_when_fake_adapter_hangs(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("shell fake uses POSIX sh")
	}
	dir := t.TempDir()
	tool, ok := ToolByID("goose")
	if !ok {
		t.Fatal("goose tool must be registered")
	}
	t.Setenv(tool.EnvVar, writeFakeAdapter(t, dir, "goose", "hang"))

	// When
	report := Doctor(context.Background(), tool, DoctorOptions{Timeout: 10 * time.Millisecond})

	// Then
	if report.ErrorKind != string(ErrorKindTimeout) {
		t.Fatalf("error kind = %q, want timeout", report.ErrorKind)
	}
	if report.Health.Status != HealthFail {
		t.Fatalf("health status = %q, want fail", report.Health.Status)
	}
}

func readChildPID(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read child pid marker: %v", err)
	}
	pid := strings.TrimSpace(string(body))
	if pid == "" {
		t.Fatal("child pid marker was empty")
	}
	return pid
}

func requireProcessExited(t *testing.T, pid string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := exec.Command("kill", "-0", pid).Run(); err != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("child process pid %s still exists after adapter timeout cleanup", pid)
}

func Test_ParseOutput_rejects_malformed_and_misleading_success_output(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "malformed json", text: `{"summary":`},
		{name: "misleading success text", text: "success: patch ready"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			_, err := ParseOutput(tt.text)

			// Then
			if !errors.Is(err, ErrInvalidOutput) {
				t.Fatalf("ParseOutput error = %v, want ErrInvalidOutput", err)
			}
		})
	}
}
