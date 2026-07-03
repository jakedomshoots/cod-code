package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_skips_check_fix_when_max_ceo_iterations_flag_is_exhausted(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace", root,
		"--apply-model-patches",
		"--check-fix-attempts", "1",
		"--max-ceo-iterations", "1",
		"--check",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_max_ceo_iterations_check",
		"--",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_max_ceo_iterations_model",
		"--",
		"Repair",
		"app",
	}
	t.Setenv("GO_WANT_CLI_MAX_CEO_ITERATIONS_MODEL", "1")
	t.Setenv("GO_WANT_CLI_MAX_CEO_ITERATIONS_CHECK", "1")
	t.Setenv("GO_CLI_MAX_CEO_ITERATIONS_TARGET", target)

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want failed verdict\n%s", err, out.String())
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(got) != "bad" {
		t.Fatalf("content = %q, want bad because check-fix was skipped", string(got))
	}
	var body struct {
		RunManifest struct {
			MaxCEOIterations      int  `json:"max_ceo_iterations"`
			CEOIterationCount     int  `json:"ceo_iteration_count"`
			CEOIterationExhausted bool `json:"ceo_iteration_exhausted"`
		} `json:"run_manifest"`
		CheckResults []struct {
			Status string `json:"status"`
		} `json:"check_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.RunManifest.MaxCEOIterations != 1 ||
		body.RunManifest.CEOIterationCount != 1 ||
		!body.RunManifest.CEOIterationExhausted {
		t.Fatalf("run manifest = %#v, want exhausted one-iteration budget", body.RunManifest)
	}
	if len(body.CheckResults) != 1 || body.CheckResults[0].Status != "fail" {
		t.Fatalf("check results = %#v, want one failed check", body.CheckResults)
	}
}

func Test_HelperProcess_cli_max_ceo_iterations_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MAX_CEO_ITERATIONS_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "Verification failed") {
		os.Stdout.WriteString(`{"summary":"fix patch","patches":[{"path":"app.txt","old":"bad","new":"good"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"summary":"ok"}`)
	os.Exit(0)
}

func Test_HelperProcess_cli_max_ceo_iterations_check(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MAX_CEO_ITERATIONS_CHECK") != "1" {
		return
	}
	content, err := os.ReadFile(os.Getenv("GO_CLI_MAX_CEO_ITERATIONS_TARGET"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if strings.TrimSpace(string(content)) == "good" {
		os.Stdout.WriteString("file fixed\n")
		os.Exit(0)
	}
	os.Stderr.WriteString("file still bad\n")
	os.Exit(4)
}
