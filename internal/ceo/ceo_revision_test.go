package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type ceoRevisionRunner struct{}

func (r ceoRevisionRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	summary := "initial pass"
	var patches []subagent.PatchProposal
	if packet.AgentName == "coder" && strings.Contains(packet.Task, "CEO review failed") {
		summary = "CEO revision patch ready"
		patches = []subagent.PatchProposal{
			{Path: "app.txt", Old: "bad", New: "good"},
		}
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         summary,
		PatchProposals:  patches,
		Evidence:        []string{"ran"},
	}, nil
}

func Test_Runtime_RunJob_runs_bounded_ceo_revision_after_model_veto(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["coder","reviewer"],"summary":"Use coder and reviewer."}`,
			`{"recommended_verdict":"fail","summary":"Patch app.txt before accepting."}`,
			`{"recommended_verdict":"pass","summary":"Revision accepted."}`,
		},
	}
	runtime := NewRuntimeWithSubagentRunnerAndCEOReviewer(ceoRevisionRunner{}, client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                "Repair app",
		WorkspaceDir:        root,
		ApplyModelPatches:   true,
		CEORevisionAttempts: 1,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_ceo_revision_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_REVISION_CHECK=1"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if string(got) != "good" {
		t.Fatalf("content = %q, want good", string(got))
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if report.CEOReview.Summary != "Revision accepted." {
		t.Fatalf("CEO review summary = %q, want final revision review", report.CEOReview.Summary)
	}
	if len(report.CheckResults) != 2 {
		t.Fatalf("CheckResults length = %d, want initial check and revision check", len(report.CheckResults))
	}
	if !containsString(report.ChangedFiles, "ceo-artifacts/coder-ceo-revision-1.md") {
		t.Fatalf("ChangedFiles = %+v, want CEO revision evidence", report.ChangedFiles)
	}
}

func Test_HelperProcess_ceo_revision_check(t *testing.T) {
	if os.Getenv("GO_WANT_CEO_REVISION_CHECK") != "1" {
		return
	}
	os.Stdout.WriteString("check ok\n")
	os.Exit(0)
}
