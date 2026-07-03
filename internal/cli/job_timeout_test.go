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

	"ceoharness/internal/config"
)

func Test_Run_times_out_job_when_job_timeout_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	t.Setenv("GO_WANT_CLI_JOB_TIMEOUT_MODEL", "block")

	// When
	err := Run(context.Background(), &out, []string{
		"--job-timeout-ms",
		"1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_job_timeout_model",
		"--",
		"Fix",
		"a",
		"bug",
	})

	// Then
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want context deadline exceeded", err)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want no partial report", out.String())
	}
}

func Test_Run_uses_workspace_job_timeout_default(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("GO_WANT_CLI_JOB_TIMEOUT_MODEL", "block")
	configBody := `{"job_timeout_ms":1,"model_command":["` + os.Args[0] + `","-test.run=Test_HelperProcess_cli_job_timeout_model"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "bug"})

	// Then
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want context deadline exceeded", err)
	}
}

func Test_Run_writes_job_timeout_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--init-config",
		"--job-timeout-ms",
		"2500",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.JobTimeoutMS != 2500 {
		t.Fatalf("JobTimeoutMS = %d, want 2500", cfg.JobTimeoutMS)
	}
	var body struct {
		JobTimeoutMS int `json:"job_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.JobTimeoutMS != 2500 {
		t.Fatalf("JobTimeoutMS report = %d, want 2500", body.JobTimeoutMS)
	}
}

func Test_Run_prints_job_timeout_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"job_timeout_ms":3000}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		JobTimeoutMS int `json:"job_timeout_ms"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.JobTimeoutMS != 3000 {
		t.Fatalf("JobTimeoutMS = %d, want 3000", body.JobTimeoutMS)
	}
}

func Test_HelperProcess_cli_job_timeout_model(t *testing.T) {
	switch os.Getenv("GO_WANT_CLI_JOB_TIMEOUT_MODEL") {
	case "block":
		select {}
	case "echo":
	default:
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if !strings.Contains(string(prompt), "role:") {
		os.Stderr.WriteString("missing role in prompt")
		os.Exit(2)
	}
	os.Stdout.WriteString(`{"summary":"job timeout helper response"}`)
	os.Exit(0)
}
