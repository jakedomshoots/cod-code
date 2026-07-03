package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"ceoharness/internal/ceo"
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
)

func Test_Run_prints_compact_text_report_when_text_format_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--format", "text", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	body := out.String()
	for _, want := range []string{
		"CEO verdict: pass",
		"Task: Fix a failing test",
		"Owner: coder",
		"Next: accept",
		"Progress: owner=coder next=accept verification=unverified changed=0 provider-routes=0",
		"Verification: unverified (0 required, 0 attempts)",
		"- scanner [pass]",
		"- coder [pass]",
		"- reviewer [pass]",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("text report missing %q:\n%s", want, body)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(body), "{") {
		t.Fatalf("text report should not be JSON:\n%s", body)
	}
}

func Test_RenderTextReport_prints_operator_summary_when_changes_and_checks_exist(t *testing.T) {
	// Given
	report := ceo.Report{
		Verdict:  "pass",
		JobID:    "job-000001",
		JobOwner: "coder",
		JobPacket: jobpacket.Packet{
			Task: "Fix checkout retry",
		},
		ExecutionPlan: ceo.ExecutionPlan{
			NextAction: "accept",
		},
		RunLedger: ceo.RunLedger{
			Owner:              "coder",
			NextAction:         "accept",
			VerificationStatus: "checked",
			ChangedFileCount:   2,
			ProviderRouteCount: 1,
		},
		CheckResults: []checkrunner.Result{
			{Argv: []string{"go", "test", "./internal/cli"}, Status: "pass", ExitCode: 0},
		},
		ChangedFiles: []string{"internal/cli/help.go", "internal/cli/verbs.go"},
	}

	// When
	body := renderTextReport(reportOutputRequest{Report: report})

	// Then
	for _, want := range []string{
		"Progress: owner=coder next=accept verification=checked changed=2 provider-routes=1",
		"Checks: 1 run, last pass",
		"Changed: internal/cli/help.go, internal/cli/verbs.go",
		"Next action: accept",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("text report missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "\nChanged files:\n") {
		t.Fatalf("text report should use compact changed file summary:\n%s", body)
	}
}

func Test_Run_prints_questions_and_resume_hint_when_text_report_needs_input(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace", root,
		"--format", "text",
		"--model-command", os.Args[0],
		"-test.run=Test_HelperProcess_cli_text_needs_input_model",
		"--",
		"Fix", "ambiguous", "package",
	}
	t.Setenv("GO_WANT_CLI_TEXT_NEEDS_INPUT_MODEL", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictNeedsInput) {
		t.Fatalf("Run error = %v, want ErrVerdictNeedsInput", err)
	}
	body := out.String()
	for _, want := range []string{
		"CEO verdict: needs_input",
		"Job: job-000001",
		"Next: answer subagent questions",
		"Questions:",
		"- Which package should I change?",
		"--resume job-000001 --answer",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("text report missing %q:\n%s", want, body)
		}
	}
}

func Test_RenderTextReport_hides_prior_job_context_from_task_line(t *testing.T) {
	// Given
	report := ceo.Report{
		Verdict: "pass",
		JobPacket: jobpacket.Packet{
			Task: "Continue checkout\n\nprior_job_context:\nprevious_job: job-000001\nprevious_failed_check: noisy output",
		},
		ExecutionPlan: ceo.ExecutionPlan{
			NextAction: "accept",
		},
	}

	// When
	body := renderTextReport(reportOutputRequest{Report: report})

	// Then
	if !strings.Contains(body, "Task: Continue checkout") {
		t.Fatalf("text report = %q, want base task", body)
	}
	if strings.Contains(body, "prior_job_context") || strings.Contains(body, "previous_failed_check") {
		t.Fatalf("text report leaked prior context:\n%s", body)
	}
}

func Test_RenderTextReport_prints_patch_preview_digest(t *testing.T) {
	// Given
	report := ceo.Report{
		Verdict: "pass",
		JobPacket: jobpacket.Packet{
			Task: "Patch app text",
		},
		ExecutionPlan: ceo.ExecutionPlan{
			NextAction: "accept",
		},
		PatchApproval: &ceo.PatchApproval{
			Status:        "previewed",
			PreviewDigest: "abc123",
			PreviewCount:  1,
		},
	}

	// When
	body := renderTextReport(reportOutputRequest{Report: report})

	// Then
	if !strings.Contains(body, "Patch approval: previewed digest=abc123 previews=1") {
		t.Fatalf("text report missing patch approval digest:\n%s", body)
	}
}

func Test_RenderTextReport_prints_ceo_delegation_source_when_present(t *testing.T) {
	// Given
	report := ceo.Report{
		Verdict:  "pass",
		JobOwner: "coder",
		JobPacket: jobpacket.Packet{
			Task: "Fix checkout",
		},
		ExecutionPlan: ceo.ExecutionPlan{
			NextAction: "accept",
		},
		CEODelegation: &ceo.CEODelegation{
			Source:            "history",
			SelectedSubagents: []string{"coder"},
		},
	}

	// When
	body := renderTextReport(reportOutputRequest{Report: report})

	// Then
	if !strings.Contains(body, "Delegation: history selected=coder") {
		t.Fatalf("text report missing delegation provenance:\n%s", body)
	}
}

func Test_Run_rejects_unknown_report_format(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--format", "markdown", "Fix", "a", "test"})

	// Then
	if err == nil {
		t.Fatal("expected report format error")
	}
	if !strings.Contains(err.Error(), "--format must be json, text, or events") {
		t.Fatalf("error = %q, want format guidance", err.Error())
	}
}

func Test_HelperProcess_cli_text_needs_input_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_TEXT_NEEDS_INPUT_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: scanner") {
		os.Stdout.WriteString(`{"status":"needs_input","summary":"missing target package","questions":["Which package should I change?"]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString(`{"summary":"ok"}`)
	os.Exit(0)
}
