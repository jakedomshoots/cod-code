package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func Test_Run_times_out_check_command_when_tool_timeout_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--tool-command-timeout-ms",
		"1",
		"--check",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_tool_timeout_check",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_TOOL_TIMEOUT_CHECK", "block")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		CheckResults []struct {
			Status   string `json:"status"`
			ExitCode int    `json:"exit_code"`
			Stderr   string `json:"stderr"`
		} `json:"check_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.CheckResults) != 1 || body.CheckResults[0].Status != "fail" || body.CheckResults[0].ExitCode != -1 {
		t.Fatalf("CheckResults = %#v, want timeout failure", body.CheckResults)
	}
	if !strings.Contains(body.CheckResults[0].Stderr, "context deadline exceeded") {
		t.Fatalf("Stderr = %q, want context deadline exceeded", body.CheckResults[0].Stderr)
	}
}

func Test_HelperProcess_cli_tool_timeout_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_TOOL_TIMEOUT_CHECK") != "block" {
		return
	}
	select {}
}
