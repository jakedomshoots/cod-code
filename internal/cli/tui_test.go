package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_TUIModel_navigates_jobs_and_dispatches_selected_action(t *testing.T) {
	// Given
	model := tuiModel{
		workspace: "/tmp/work",
		jobs: []tuiJob{
			{id: "job-000001", task: "Passing job", verdict: "pass", action: "accept", actionCommand: "ceo-packet --workspace /tmp/work --judge-job job-000001 --human-verdict accept"},
			{id: "job-000002", task: "Needs answer", verdict: "needs_input", inboxReason: "needs_input", action: "answer", actionCommand: "ceo-packet --workspace /tmp/work --resume job-000002 --answer \"...\""},
		},
	}

	// When
	gotModel, gotAction := model.applyKey("down")

	// Then
	if gotModel.selected != 1 {
		t.Fatalf("selected = %d, want 1", gotModel.selected)
	}
	if gotAction != "" {
		t.Fatalf("action = %q, want none while navigating", gotAction)
	}

	// When
	gotModel, gotAction = gotModel.applyKey("enter")

	// Then
	if gotModel.selected != 1 {
		t.Fatalf("selected changed = %d, want 1", gotModel.selected)
	}
	if gotAction != "ceo-packet --workspace /tmp/work --resume job-000002 --answer \"...\"" {
		t.Fatalf("action = %q, want resume command", gotAction)
	}
}

func Test_Run_tui_snapshot_shows_dashboard_when_jobs_exist(t *testing.T) {
	// Given
	root := t.TempDir()
	saveTUIDashboardJob(t, root, history.Entry{
		Task:       "Passing checkout",
		Verdict:    "pass",
		CheckCount: 1,
		PatchCount: 1,
		ProviderHealth: []history.ProviderHealth{{
			ProviderName: "main",
			AttemptCount: 1,
			PassCount:    1,
		}},
	})
	saveTUIDashboardJob(t, root, history.Entry{
		Task:       "Needs customer input",
		Verdict:    "needs_input",
		CheckCount: 1,
	})
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--tui", "--snapshot"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"CEO Harness Dashboard",
		"Jobs (2)",
		"> job-000002 [needs_input] Needs customer input",
		"Inbox: needs_input",
		"Provider health: 1 provider, 1 attempt, 1 pass, 0 fail",
		"Patch preview: app.txt",
		"Check output: go test ./... pass",
		"Action: answer -> ceo-packet --workspace",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_tui_snapshot_shows_setup_guidance_when_workspace_empty(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--tui", "--snapshot"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"CEO Harness Dashboard",
		"No saved jobs yet.",
		"ceo-packet --quickstart",
		"ceo-packet --workspace",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("empty snapshot missing %q:\n%s", want, text)
		}
	}
}

func Test_RunWithIO_tui_processes_navigation_and_dispatches_action(t *testing.T) {
	// Given
	root := t.TempDir()
	saveTUIDashboardJob(t, root, history.Entry{
		Task:    "Passing checkout",
		Verdict: "pass",
	})
	saveTUIDashboardJob(t, root, history.Entry{
		Task:    "Needs customer input",
		Verdict: "needs_input",
	})
	var out bytes.Buffer

	// When
	err := RunWithIO(context.Background(), strings.NewReader("down\nup\nenter\nq\n"), &out, []string{"--workspace", root, "--tui"})

	// Then
	if err != nil {
		t.Fatalf("RunWithIO returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"> job-000001 [pass] Passing checkout",
		"> job-000002 [needs_input] Needs customer input",
		"Dispatched action: ceo-packet --workspace",
		"--resume job-000002 --answer \"...\"",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("interactive tui output missing %q:\n%s", want, text)
		}
	}
}

func saveTUIDashboardJob(t *testing.T, root string, entry history.Entry) {
	t.Helper()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}
	stored, err := store.Append(context.Background(), entry)
	if err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	payload := `{
  "schema_version": 1,
  "job_id": "` + stored.ID + `",
  "verdict": "` + stored.Verdict + `",
  "job_packet": {"task": "` + stored.Task + `"},
  "patch_previews": [{"path":"app.txt","diff":"--- app.txt\n+++ app.txt\n-old\n+new"}],
  "check_results": [{"argv":["go","test","./..."],"status":"pass","stdout":"ok ceoharness","stderr":""}]
}`
	if _, err := store.SaveReportSnapshot(context.Background(), stored.ID, []byte(payload)); err != nil {
		t.Fatalf("SaveReportSnapshot: %v", err)
	}
}
