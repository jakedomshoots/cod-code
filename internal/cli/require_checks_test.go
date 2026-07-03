package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_rejects_strict_run_when_no_verification_check_is_configured(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--require-checks", "Fix", "checkout", "bug"})

	// Then
	if err == nil {
		t.Fatal("expected missing verification error")
	}
	if !strings.Contains(err.Error(), "--require-checks requires at least one verification command") {
		t.Fatalf("error = %q, want require-checks guidance", err.Error())
	}
}

func Test_Run_uses_workspace_require_checks_when_check_command_is_configured(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"require_checks":true,"check_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_require_checks_check"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_REQUIRE_CHECKS_CHECK", "1")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "checkout", "bug"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		VerificationContract struct {
			Status             string `json:"status"`
			RequiredCheckCount int    `json:"required_check_count"`
		} `json:"verification_contract"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.VerificationContract.Status != "pass" || body.VerificationContract.RequiredCheckCount != 1 {
		t.Fatalf("verification contract = %#v, want one required passing check", body.VerificationContract)
	}
}

func Test_Run_rejects_strict_plan_only_when_no_verification_check_is_configured(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--plan-only", "--require-checks", "Fix", "checkout", "bug"})

	// Then
	if err == nil {
		t.Fatal("expected missing verification error")
	}
	if !strings.Contains(err.Error(), "--require-checks requires at least one verification command") {
		t.Fatalf("error = %q, want require-checks guidance", err.Error())
	}
}

func Test_Run_writes_require_checks_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--init-config", "--require-checks"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RequireChecks bool `json:"require_checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if !body.RequireChecks {
		t.Fatalf("RequireChecks = false, want true")
	}
	content, readErr := os.ReadFile(filepath.Join(root, ".ceo-harness.json"))
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}
	if !strings.Contains(string(content), `"require_checks": true`) {
		t.Fatalf("config = %s, want require_checks true", string(content))
	}
}

func Test_Run_prints_require_checks_in_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"require_checks":true}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RequireChecks bool `json:"require_checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if !body.RequireChecks {
		t.Fatalf("RequireChecks = false, want true")
	}
}

func Test_HelperProcess_cli_require_checks_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_REQUIRE_CHECKS_CHECK") != "1" {
		return
	}
	os.Stdout.WriteString("strict verification passed")
}
