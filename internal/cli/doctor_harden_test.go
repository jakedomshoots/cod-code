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
)

func Test_Run_Doctor_skips_missing_optional_provider_without_failing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"optional":{"model_command":["ceo-definitely-missing-provider-cli"]}}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	body := decodeDoctorHardenReport(t, out.Bytes())
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	check := requireDoctorHardenCheck(t, body.Checks, "provider.optional")
	if check.Status != "skipped" || check.Requirement != "optional" || check.Guidance == "" {
		t.Fatalf("provider.optional check = %+v, want optional skipped with guidance", check)
	}
}

func Test_Run_Doctor_blocks_required_http_provider_when_api_key_is_missing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"http":{"url":"http://127.0.0.1:1/v1/chat/completions","model":"gpt-5","api_key_env":"CEO_MISSING_KEY"}}},"provider_policy":{"default_provider":"main"}}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("CEO_MISSING_KEY", "")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err == nil {
		t.Fatal("Run returned nil error, want required provider block")
	}
	body := decodeDoctorHardenReport(t, out.Bytes())
	if body.Status != "fail" {
		t.Fatalf("Status = %q, want fail", body.Status)
	}
	check := requireDoctorHardenCheck(t, body.Checks, "provider.main")
	if check.Status != "blocked" || check.Requirement != "required" || !strings.Contains(check.Error, "CEO_MISSING_KEY") {
		t.Fatalf("provider.main check = %+v, want required blocked missing key", check)
	}
}

func Test_Run_Doctor_blocks_missing_workspace_path(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := filepath.Join(t.TempDir(), "missing")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err == nil {
		t.Fatal("Run returned nil error, want missing workspace block")
	}
	body := decodeDoctorHardenReport(t, out.Bytes())
	check := requireDoctorHardenCheck(t, body.Checks, "workspace")
	if check.Status != "blocked" || check.Requirement != "required" {
		t.Fatalf("workspace check = %+v, want required blocked", check)
	}
}

func Test_Run_Doctor_reports_malformed_config_as_blocked(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"model_command":[`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err == nil {
		t.Fatal("Run returned nil error, want malformed config block")
	}
	body := decodeDoctorHardenReport(t, out.Bytes())
	check := requireDoctorHardenCheck(t, body.Checks, "workspace_config")
	if check.Status != "blocked" || check.Requirement != "required" || check.Error == "" {
		t.Fatalf("workspace_config check = %+v, want blocked malformed config", check)
	}
}

func Test_DoctorToolChecks_classify_required_go_and_optional_tools(t *testing.T) {
	// Given
	lookup := func(name string) (string, error) {
		if name == "go" {
			return "/usr/local/bin/go", nil
		}
		return "", errors.New("missing")
	}

	// When
	checks := doctorToolChecks(lookup)

	// Then
	goCheck := requireDoctorCheck(t, checks, "tool.go")
	if goCheck.Status != "pass" || goCheck.Requirement != "required" {
		t.Fatalf("tool.go check = %+v, want required pass", goCheck)
	}
	strictCheck := requireDoctorCheck(t, checks, "tool.gofumpt")
	if strictCheck.Status != "skipped" || strictCheck.Requirement != "optional" {
		t.Fatalf("tool.gofumpt check = %+v, want optional skipped", strictCheck)
	}
	providerCheck := requireDoctorCheck(t, checks, "provider_cli.kimi")
	if providerCheck.Status != "skipped" || providerCheck.Requirement != "optional" {
		t.Fatalf("provider_cli.kimi check = %+v, want optional skipped", providerCheck)
	}
}

type doctorHardenReport struct {
	Status string              `json:"status"`
	Checks []doctorHardenCheck `json:"checks"`
}

type doctorHardenCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Requirement string `json:"requirement"`
	Error       string `json:"error"`
	Guidance    string `json:"guidance"`
}

func decodeDoctorHardenReport(t *testing.T, content []byte) doctorHardenReport {
	t.Helper()
	var body doctorHardenReport
	if err := json.Unmarshal(content, &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, string(content))
	}
	return body
}

func requireDoctorHardenCheck(t *testing.T, checks []doctorHardenCheck, name string) doctorHardenCheck {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("Checks = %#v, want %s", checks, name)
	return doctorHardenCheck{}
}

func requireDoctorCheck(t *testing.T, checks []doctorCheck, name string) doctorCheck {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("Checks = %#v, want %s", checks, name)
	return doctorCheck{}
}
