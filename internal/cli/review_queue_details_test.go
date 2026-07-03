package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/ceo"
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_Run_review_queue_includes_compact_context_when_details_flag_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	saveReviewDetailJob(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--review-queue", "--review-details"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Queue []struct {
			ID            string             `json:"id"`
			ReviewReason  string             `json:"review_reason"`
			ReviewContext *compactJobContext `json:"review_context"`
		} `json:"review_queue"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.Queue) != 1 || body.Queue[0].ID != "job-000001" {
		t.Fatalf("queue = %#v, want one detailed review row", body.Queue)
	}
	context := body.Queue[0].ReviewContext
	if context == nil {
		t.Fatalf("review_context missing from %#v", body.Queue[0])
	}
	if context.NextAction != "answer subagent questions" || len(context.Questions) != 1 {
		t.Fatalf("review_context = %#v, want next action and question", context)
	}
	if len(context.ChangedFiles) != 1 || context.ChangedFiles[0] != "internal/cli/app.go" {
		t.Fatalf("changed files = %#v, want snapshot changed file", context.ChangedFiles)
	}
	if len(context.FailedChecks) != 1 || !strings.Contains(context.FailedChecks[0].FailureExcerpt, "compile failed") {
		t.Fatalf("failed checks = %#v, want compact failure excerpt", context.FailedChecks)
	}
}

func Test_Run_review_queue_prints_compact_context_when_text_details_flag_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	saveReviewDetailJob(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--review-queue", "--review-details", "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Action: answer subagent questions",
		"Question: Which package should I change?",
		"Changed: internal/cli/app.go",
		"Failed check: go test ./...",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("review queue text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_rejects_review_details_without_review_queue(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--review-details"})

	// Then
	if err == nil {
		t.Fatal("expected usage error")
	}
	if !strings.Contains(err.Error(), "--review-details requires --review-queue") {
		t.Fatalf("error = %q, want review queue guidance", err.Error())
	}
}

func Test_Run_review_queue_details_loads_context_only_after_limit_is_applied(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"Old bad snapshot", "Latest needs review"} {
		if _, err := store.Append(context.Background(), history.Entry{Task: task, Verdict: "needs_input"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}
	badSnapshotPath := filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json")
	if err := os.MkdirAll(filepath.Dir(badSnapshotPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(badSnapshotPath, []byte("{"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err = Run(context.Background(), &out, []string{
		"--workspace", root,
		"--review-queue",
		"--review-details",
		"--limit", "1",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Queue []struct {
			ID            string             `json:"id"`
			ReviewContext *compactJobContext `json:"review_context"`
		} `json:"review_queue"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.Queue) != 1 || body.Queue[0].ID != "job-000002" || body.Queue[0].ReviewContext == nil {
		t.Fatalf("queue = %#v, want only latest row with context", body.Queue)
	}
}

func saveReviewDetailJob(t *testing.T, root string) {
	t.Helper()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:                    "Fix checkout",
		Verdict:                 "needs_input",
		ChangedFiles:            []string{"internal/cli/app.go"},
		ExecutionPlanNextAction: "answer subagent questions",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	report := ceo.Report{
		JobID:   "job-000001",
		Verdict: "needs_input",
		JobPacket: jobpacket.Packet{
			Task: "Fix checkout",
		},
		Resume: &ceo.ResumeContext{Questions: []string{"Which package should I change?"}},
		SubagentResults: []subagent.Result{{
			AgentName: "scanner",
			Status:    "needs_input",
			Summary:   "Need package.",
			Questions: []string{"Which package should I change?"},
		}},
		ChangedFiles: []string{"internal/cli/app.go"},
		CheckResults: []checkrunner.Result{{
			Argv:     []string{"go", "test", "./..."},
			Status:   "fail",
			ExitCode: 1,
			Stderr:   "compile failed",
		}},
		ExecutionPlan: ceo.ExecutionPlan{NextAction: "answer subagent questions"},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", payload); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
}
