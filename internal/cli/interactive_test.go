package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_RunWithIO_interactively_answers_needs_input_and_resumes(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace", root,
		"--interactive",
		"--model-command", os.Args[0],
		"-test.run=Test_HelperProcess_cli_interactive_model",
		"--",
		"Fix", "ambiguous", "package",
	}
	t.Setenv("GO_WANT_CLI_INTERACTIVE_MODEL", "1")

	// When
	err := RunWithIO(context.Background(), strings.NewReader("Use internal/cli.\n"), &out, args)
	// Then
	if err != nil {
		t.Fatalf("RunWithIO returned error: %v\n%s", err, out.String())
	}
	body := out.String()
	for _, want := range []string{
		"CEO verdict: needs_input",
		"Which package should I change?",
		"> ",
		"CEO verdict: pass",
		"Job: job-000002",
		"answer received",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("interactive output missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "CEO verdict failed") {
		t.Fatalf("interactive output includes failure noise:\n%s", body)
	}
	if strings.Contains(body, "Task: Fix ambiguous package resume_context") {
		t.Fatalf("interactive text task leaked resume context:\n%s", body)
	}
}

func Test_RunWithIO_rejects_interactive_json_format(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := RunWithIO(context.Background(), strings.NewReader("answer\n"), &out, []string{
		"--interactive",
		"--format", "json",
		"Fix", "ambiguous", "package",
	})

	// Then
	if err == nil {
		t.Fatal("expected interactive format error")
	}
	if !strings.Contains(err.Error(), "--interactive requires --format text") {
		t.Fatalf("error = %q, want interactive text guidance", err.Error())
	}
}

func Test_HelperProcess_cli_interactive_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_INTERACTIVE_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "answer: Use internal/cli.") {
		os.Stdout.WriteString(`{"summary":"answer received"}`)
		os.Exit(0)
	}
	if strings.Contains(text, "agent: scanner") {
		os.Stdout.WriteString(`{"status":"needs_input","summary":"missing target package","questions":["Which package should I change?"]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"summary":"ok"}`)
	os.Exit(0)
}
