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

func Test_Run_prints_failing_verification_policy_doctor_check_when_required_checks_are_missing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"require_checks":true}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "fail" {
		t.Fatalf("Status = %q, want fail", body.Status)
	}
	for _, check := range body.Checks {
		if check.Name == "verification_policy" && check.Status == "fail" && strings.Contains(check.Error, "--require-checks requires at least one verification command") {
			return
		}
	}
	t.Fatalf("Checks = %#v, want failing verification_policy check", body.Checks)
}

func Test_Run_prints_passing_verification_policy_doctor_check_when_required_checks_are_configured(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"require_checks":true,"check_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_doctor_verification_check"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	for _, check := range body.Checks {
		if check.Name == "verification_policy" && check.Status == "pass" && check.Error == "" {
			return
		}
	}
	t.Fatalf("Checks = %#v, want passing verification_policy check", body.Checks)
}

func Test_HelperProcess_cli_doctor_verification_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_DOCTOR_VERIFICATION_CHECK") != "1" {
		return
	}
	os.Stdout.WriteString("doctor verification passed")
}
