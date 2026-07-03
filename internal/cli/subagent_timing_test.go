package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func Test_Run_prints_subagent_duration(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_slow_model",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_SLOW_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			DurationMS int64 `json:"duration_ms"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SubagentResults[0].DurationMS <= 0 {
		t.Fatalf("DurationMS = %d, want positive duration", body.SubagentResults[0].DurationMS)
	}
}

func Test_HelperProcess_cli_slow_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_SLOW_MODEL") != "1" {
		return
	}
	time.Sleep(5 * time.Millisecond)
	os.Stdout.WriteString("ok")
}
