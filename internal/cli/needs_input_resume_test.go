package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_Run_resumes_needs_input_job_with_answer(t *testing.T) {
	// Given
	root := t.TempDir()
	seedArgs := []string{
		"--workspace", root,
		"--model-command", os.Args[0],
		"-test.run=Test_HelperProcess_cli_resume_model",
		"--",
		"Fix", "ambiguous", "package",
	}
	t.Setenv("GO_WANT_CLI_RESUME_MODEL", "seed")
	err := Run(context.Background(), &bytes.Buffer{}, seedArgs)
	if !errors.Is(err, ErrVerdictNeedsInput) {
		t.Fatalf("seed Run error = %v, want ErrVerdictNeedsInput", err)
	}
	var out bytes.Buffer
	resumeArgs := []string{
		"--workspace", root,
		"--resume", "job-000001",
		"--answer", "Use internal/cli.",
		"--model-command", os.Args[0],
		"-test.run=Test_HelperProcess_cli_resume_model",
		"--",
	}
	t.Setenv("GO_WANT_CLI_RESUME_MODEL", "resume")

	// When
	err = Run(context.Background(), &out, resumeArgs)
	// Then
	if err != nil {
		t.Fatalf("resume Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobID     string `json:"job_id"`
		Verdict   string `json:"verdict"`
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
		Resume struct {
			JobID     string   `json:"job_id"`
			Questions []string `json:"questions"`
			Answers   []string `json:"answers"`
		} `json:"resume"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobID != "job-000002" {
		t.Fatalf("JobID = %q, want job-000002", body.JobID)
	}
	if body.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", body.Verdict)
	}
	if body.Resume.JobID != "job-000001" {
		t.Fatalf("Resume.JobID = %q, want job-000001", body.Resume.JobID)
	}
	if len(body.Resume.Questions) != 1 || body.Resume.Questions[0] != "Which package should I change?" {
		t.Fatalf("Resume.Questions = %#v, want prior question", body.Resume.Questions)
	}
	if len(body.Resume.Answers) != 1 || body.Resume.Answers[0] != "Use internal/cli." {
		t.Fatalf("Resume.Answers = %#v, want supplied answer", body.Resume.Answers)
	}
	if !strings.Contains(body.JobPacket.Task, "resume_context:") || !strings.Contains(body.JobPacket.Task, "Use internal/cli.") {
		t.Fatalf("task = %q, want compact resume context", body.JobPacket.Task)
	}
}

func Test_Run_rejects_resume_without_answer(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--resume", "job-000001"})

	// Then
	if err == nil {
		t.Fatal("expected resume answer error")
	}
	if !strings.Contains(err.Error(), "--resume requires at least one --answer") {
		t.Fatalf("error = %q, want answer guidance", err.Error())
	}
}

func Test_Run_rejects_answer_without_resume(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--answer", "Use internal/cli.", "Fix", "ambiguous", "package"})

	// Then
	if err == nil {
		t.Fatal("expected answer/resume conflict error")
	}
	if !strings.Contains(err.Error(), "--answer requires --resume") {
		t.Fatalf("error = %q, want resume guidance", err.Error())
	}
}

func Test_HelperProcess_cli_resume_model(t *testing.T) {
	mode := os.Getenv("GO_WANT_CLI_RESUME_MODEL")
	if mode == "" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	switch mode {
	case "seed":
		if strings.Contains(string(prompt), "agent: scanner") {
			os.Stdout.WriteString(`{"status":"needs_input","summary":"missing target package","questions":["Which package should I change?"]}`)
			os.Exit(0)
		}
		os.Stdout.WriteString(`{"summary":"ok"}`)
		os.Exit(0)
	case "resume":
		text := string(prompt)
		if strings.Contains(text, "previous_job: job-000001") &&
			strings.Contains(text, "question: Which package should I change?") &&
			strings.Contains(text, "answer: Use internal/cli.") {
			os.Stdout.WriteString(`{"summary":"answer received"}`)
			os.Exit(0)
		}
		os.Stdout.WriteString(`{"status":"needs_input","summary":"answer missing","questions":["Which package should I change?"]}`)
		os.Exit(0)
	}
}
