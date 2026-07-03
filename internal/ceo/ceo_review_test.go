package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

type fakeCEOReviewClient struct {
	text   string
	prompt string
}

func (c *fakeCEOReviewClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.prompt = req.Prompt
	if strings.Contains(req.Prompt, "candidate_subagents") {
		if strings.Contains(req.Prompt, "- database role=") {
			return model.Response{
				Text:        `{"selected_subagents":["coder","database","reviewer"],"summary":"Use the lean migration crew."}`,
				PromptBytes: len(req.Prompt),
			}, nil
		}
		return model.Response{
			Text:        `{"selected_subagents":["scanner","coder","reviewer"],"summary":"Use the default coding crew."}`,
			PromptBytes: len(req.Prompt),
		}, nil
	}
	return model.Response{
		Text:        c.text,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runtime_RunJob_lets_model_ceo_veto_when_guard_verdict_passes(t *testing.T) {
	// Given
	reviewer := &fakeCEOReviewClient{
		text: `{"recommended_verdict":"fail","summary":"The database reviewer missed migration risk."}`,
	}
	runtime := NewRuntimeWithCEOReviewer(reviewer)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix migration bug",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", report.Verdict)
	}
	if report.CEOReview.RecommendedVerdict != "fail" {
		t.Fatalf("RecommendedVerdict = %q, want fail", report.CEOReview.RecommendedVerdict)
	}
	if report.CEOReview.Summary != "The database reviewer missed migration risk." {
		t.Fatalf("CEO summary = %q, want model summary", report.CEOReview.Summary)
	}
	if !strings.Contains(reviewer.prompt, "task: Fix migration bug") {
		t.Fatalf("CEO prompt = %q, want task", reviewer.prompt)
	}
}

func Test_Runtime_RunJob_keeps_failed_guard_when_model_ceo_recommends_pass(t *testing.T) {
	// Given
	reviewer := &fakeCEOReviewClient{
		text: `{"recommended_verdict":"pass","summary":"Looks fine."}`,
	}
	runtime := NewRuntimeWithCEOReviewer(reviewer)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix failing check",
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_fail_check",
		},
		CheckEnv: []string{"GO_WANT_CEO_HELPER_PROCESS=fail"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEOReview.RecommendedVerdict != "pass" {
		t.Fatalf("RecommendedVerdict = %q, want pass", report.CEOReview.RecommendedVerdict)
	}
	if report.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want failed guard to remain fail", report.Verdict)
	}
}

type ceoReviewEvidenceRunner struct{}

func (r ceoReviewEvidenceRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if packet.AgentName == "scanner" && len(packet.ToolResults) == 0 {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			Summary:         "scanner requested file",
			ToolRequests: []subagent.ToolRequest{
				{Action: "read_workspace", Path: "app.txt"},
			},
			Evidence: []string{"requested read"},
		}, nil
	}
	if packet.AgentName == "scanner" {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			Summary:         "scanner saw " + packet.ToolResults[0].Output,
			Evidence:        []string{"tool result used"},
		}, nil
	}
	if packet.AgentName == "coder" {
		return subagent.Result{
			AgentName:       packet.AgentName,
			Role:            packet.Role,
			Status:          "pass",
			Attempts:        1,
			ContextReceived: packet.ContextMode,
			Summary:         "coder patched app",
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt", Old: "old", New: "new"},
			},
			Evidence: []string{"patch proposed"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         "reviewer ok",
		Evidence:        []string{"reviewed"},
	}, nil
}

func Test_Runtime_RunJob_includes_patch_and_tool_evidence_in_ceo_review_prompt(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	reviewer := &fakeCEOReviewClient{text: `{"recommended_verdict":"pass","summary":"Evidence checked."}`}
	runtime := NewRuntimeWithSubagentRunnerAndCEOReviewer(ceoReviewEvidenceRunner{}, reviewer)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Fix app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	for _, want := range []string{
		"changed_files:",
		"app.txt",
		"patch_results:",
		"+new",
		"tool_results:",
		"read_workspace",
		"old",
	} {
		if !strings.Contains(reviewer.prompt, want) {
			t.Fatalf("CEO prompt missing %q:\n%s", want, reviewer.prompt)
		}
	}
}
