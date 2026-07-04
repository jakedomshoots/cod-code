package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_quickstarts_workspace_with_example_adapters_and_doctor_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status     string           `json:"status"`
		Workspace  string           `json:"workspace"`
		ConfigInit configInitReport `json:"config_init"`
		Doctor     doctorReport     `json:"doctor"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "pass" || body.Workspace != root {
		t.Fatalf("quickstart status/workspace = %q/%q, want pass/%q", body.Status, body.Workspace, root)
	}
	if body.ConfigInit.ConfigPath != filepath.Join(root, ".ceo-harness.json") || !body.ConfigInit.ExampleAdapters {
		t.Fatalf("config init = %#v, want example config path", body.ConfigInit)
	}
	requireQuickstartDoctorCheck(t, body.Doctor.Checks, "model_command")
	requireQuickstartDoctorCheck(t, body.Doctor.Checks, "ceo_model_command")
	requireQuickstartDoctorCheck(t, body.Doctor.Checks, "research_command")
}

func Test_Run_quickstart_prints_text_first_run_checklist_when_format_text_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root, "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Quickstart: pass",
		"Workspace: " + root,
		"Config: " + filepath.Join(root, ".ceo-harness.json"),
		"Doctor: pass",
		"Next:",
		`ceo-packet oauth doctor --format text`,
		`ceo-packet oauth init kimi --workspace "` + root + `" --format text`,
		`ceo-packet run --workspace "` + root + `" --check go test ./... -- "Fix one real task"`,
		`ceo-packet production-status --workspace "` + root + `" --format text`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("quickstart text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_quickstart_text_includes_provider_setup_steps_when_provider_check_fails(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("CEO_MISSING_TOKEN", "")

	// When
	err := Run(context.Background(), &out, []string{
		"--quickstart",
		root,
		"--format",
		"text",
		"--http-provider",
		"fast",
		"--http-url",
		"http://127.0.0.1:1/v1/chat/completions",
		"--http-model",
		"fast-model",
		"--http-api-key-env",
		"CEO_MISSING_TOKEN",
		"--default-provider",
		"fast",
	})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want failed quickstart verdict\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Quickstart: fail",
		"export CEO_MISSING_TOKEN=...",
		`ceo-packet --workspace "` + root + `" --doctor-provider "fast" --format text`,
		`ceo-packet oauth doctor --format text`,
		`ceo-packet oauth init kimi --workspace "` + root + `" --format text`,
		`ceo-packet run --workspace "` + root + `" --check go test ./... -- "Fix one real task"`,
		`ceo-packet production-status --workspace "` + root + `" --format text`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("quickstart text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_quickstart_rejects_events_format(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root, "--format", "events"})

	// Then
	if err == nil {
		t.Fatal("expected events format error")
	}
	if !strings.Contains(err.Error(), "only available for run reports") {
		t.Fatalf("error = %q, want run report guidance", err.Error())
	}
}

func Test_Run_quickstart_enables_required_go_checks_when_go_module_is_detected(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test/quickstart\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ConfigInit configInitReport `json:"config_init"`
		Doctor     doctorReport     `json:"doctor"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ConfigInit.CheckCommandArgc != 3 || !body.ConfigInit.RequireChecks {
		t.Fatalf("config init check policy = argc %d require %v, want go test ./... required", body.ConfigInit.CheckCommandArgc, body.ConfigInit.RequireChecks)
	}
	requireQuickstartDoctorStatus(t, body.Doctor.Checks, "verification_policy")
}

func Test_Run_quickstart_enables_required_npm_checks_when_package_test_script_is_detected(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	packageJSON := `{"scripts":{"test":"vitest run"}}`
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(packageJSON), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ConfigInit configInitReport `json:"config_init"`
		Doctor     doctorReport     `json:"doctor"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ConfigInit.CheckCommandArgc != 2 || !body.ConfigInit.RequireChecks {
		t.Fatalf("config init check policy = argc %d require %v, want npm test required", body.ConfigInit.CheckCommandArgc, body.ConfigInit.RequireChecks)
	}
	requireQuickstartDoctorStatus(t, body.Doctor.Checks, "verification_policy")
}

func Test_Run_quickstart_refuses_to_overwrite_existing_config(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := Run(context.Background(), &out, []string{"--workspace", root, "--init-config", "--init-example-adapters"}); err != nil {
		t.Fatalf("init config returned error: %v\n%s", err, out.String())
	}
	out.Reset()

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root})

	// Then
	if !errors.Is(err, config.ErrConfigExists) {
		t.Fatalf("error = %v, want ErrConfigExists", err)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty output", out.String())
	}
}

func Test_Run_quickstart_rejects_task_text(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--quickstart", root, "Fix", "tests"})

	// Then
	if err == nil {
		t.Fatal("expected quickstart usage error")
	}
	if !strings.Contains(err.Error(), "--quickstart cannot be combined with task text") {
		t.Fatalf("error = %q, want quickstart task guidance", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty output", out.String())
	}
}

func requireQuickstartDoctorCheck(t *testing.T, checks []doctorCheck, name string) {
	t.Helper()
	for _, check := range checks {
		if check.Name == name && check.Status == "pass" && check.Source == "workspace" {
			return
		}
	}
	t.Fatalf("doctor checks = %#v, want passing workspace check %s", checks, name)
}

func requireQuickstartDoctorStatus(t *testing.T, checks []doctorCheck, name string) {
	t.Helper()
	for _, check := range checks {
		if check.Name == name && check.Status == "pass" {
			return
		}
	}
	t.Fatalf("doctor checks = %#v, want passing check %s", checks, name)
}
