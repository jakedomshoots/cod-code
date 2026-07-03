package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Test_Run_prints_verification_contract_when_check_runs(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--check",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_verification_contract_check",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_VERIFICATION_CONTRACT_CHECK", "pass")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		VerificationContract struct {
			Status             string   `json:"status"`
			RequiredCheckCount int      `json:"required_check_count"`
			RequiredChecks     []string `json:"required_checks"`
			CheckAttemptCount  int      `json:"check_attempt_count"`
			PassedCheckCount   int      `json:"passed_check_count"`
			FailedCheckCount   int      `json:"failed_check_count"`
		} `json:"verification_contract"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	contract := body.VerificationContract
	if contract.Status != "pass" || contract.RequiredCheckCount != 1 || contract.CheckAttemptCount != 1 {
		t.Fatalf("verification contract = %#v, want one passed check", contract)
	}
	if contract.PassedCheckCount != 1 || contract.FailedCheckCount != 0 {
		t.Fatalf("verification contract counts = %#v, want pass count only", contract)
	}
	if len(contract.RequiredChecks) != 1 || !strings.Contains(contract.RequiredChecks[0], "verification_contract_check") {
		t.Fatalf("required checks = %#v, want helper command", contract.RequiredChecks)
	}
}

func Test_Run_prints_verification_contract_in_text_report(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--format",
		"text",
		"--check",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_verification_contract_check",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_VERIFICATION_CONTRACT_CHECK", "pass")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	body := out.String()
	if !strings.Contains(body, "Verification: pass (1 required, 1 attempt)") {
		t.Fatalf("text report missing verification contract:\n%s", body)
	}
}

func Test_HelperProcess_cli_verification_contract_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_VERIFICATION_CONTRACT_CHECK") != "pass" {
		return
	}
	os.Stdout.WriteString("verification contract passed")
}
