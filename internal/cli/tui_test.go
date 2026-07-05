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
			{id: "job-000001", task: "Passing job", verdict: "pass", action: "accept", actionCommand: "cod --workspace /tmp/work --judge-job job-000001 --human-verdict accept"},
			{id: "job-000002", task: "Needs answer", verdict: "needs_input", inboxReason: "needs_input", action: "answer", actionCommand: "cod --workspace /tmp/work --resume job-000002 --answer \"...\""},
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
	if gotAction != "cod --workspace /tmp/work --resume job-000002 --answer \"...\"" {
		t.Fatalf("action = %q, want resume command", gotAction)
	}

	// When — the 'a' shortcut must dispatch the same primary action as enter on the selected job.
	_, gotAction = gotModel.applyKey("a")

	// Then
	if gotAction != "cod --workspace /tmp/work --resume job-000002 --answer \"...\"" {
		t.Fatalf("a action = %q, want primary action command", gotAction)
	}

	// When — the 'r' shortcut must surface the rerun command for the selected job.
	_, gotAction = gotModel.applyKey("r")
	// Then — rerun command quotes the workspace path through workspaceArg.
	if gotAction != "cod --workspace \"/tmp/work\" --rerun job-000002" {
		t.Fatalf("r action = %q, want rerun command for selected job", gotAction)
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
		"╭─ Cod Code",
		"Chat-first coding terminal",
		"Session",
		"Conversation",
		"You       Needs customer input",
		"Cod       needs_input · waiting on you",
		"Diff      app.txt",
		"Check     go test ./... pass",
		"Action    answer · cod --workspace",
		"Rerun     cod --workspace",
		"Activity",
		"INPUT",
		"Needs input (1)",
		"REVIEW",
		"Needs decision (1)",
		"› job-000002",
		"Status",
		"Providers 1 provider · 1 attempt · 1 pass · 0 fail",
		"Composer",
		"Keys      j/k move · enter/a act · r rerun · q quit",
		"Inbox     cod inbox --workspace",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("snapshot missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_tui_snapshot_uses_provider_zero_state_guidance_when_no_provider_history(t *testing.T) {
	// Given
	root := t.TempDir()
	saveTUIDashboardJob(t, root, history.Entry{
		Task:    "Needs review",
		Verdict: "needs_input",
	})
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--tui", "--snapshot"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	if !strings.Contains(text, "Providers no evidence yet; run cod doctor or provider proof.") {
		t.Fatalf("snapshot must surface provider zero-state guidance when no provider history exists:\n%s", text)
	}
	if !strings.Contains(text, "Status") {
		t.Fatalf("snapshot must surface Status block above provider zero-state guidance:\n%s", text)
	}
	if !strings.Contains(text, "Composer") {
		t.Fatalf("snapshot must surface Composer block on provider zero state:\n%s", text)
	}
	if strings.Contains(text, "0 providers") {
		t.Fatalf("snapshot must not surface legacy '0 providers' fallback in provider health block:\n%s", text)
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
		"╭─ Cod Code",
		"Conversation",
		"No chat yet.",
		"No saved jobs yet.",
		"cod start",
		"cod doctor --workspace",
		"Providers no evidence yet; run cod doctor or provider proof.",
		"Status",
		"Composer",
		"Inbox     cod inbox --workspace",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("empty snapshot missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_chat_and_dev_aliases_open_tui(t *testing.T) {
	for _, verb := range []string{"chat", "dev"} {
		var out bytes.Buffer
		root := t.TempDir()

		if err := Run(context.Background(), &out, []string{verb, "--workspace", root, "--snapshot"}); err != nil {
			t.Fatalf("Run %s returned error: %v\n%s", verb, err, out.String())
		}
		if body := out.String(); !strings.Contains(body, "╭─ Cod Code") || !strings.Contains(body, "Composer") {
			t.Fatalf("%s output missing TUI markers:\n%s", verb, body)
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

	// When — exercise enter (primary action), a (alias for primary), and r (rerun).
	err := RunWithIO(context.Background(), strings.NewReader("down\nup\nenter\na\nr\nq\n"), &out, []string{"--workspace", root, "--tui"})
	// Then
	if err != nil {
		t.Fatalf("RunWithIO returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"╭─ Cod Code",
		"Conversation",
		"Activity",
		"INPUT",
		"Needs input (1)",
		"REVIEW",
		"Needs decision (1)",
		"› job-000002",
		"Needs customer input",
		"Action    answer · cod --workspace",
		"Rerun     cod --workspace",
		"Composer",
		"Keys      j/k move · enter/a act · r rerun · q quit",
		"Action dispatched: cod --workspace",
		"--resume job-000002 --answer \"...\"",
		"Action dispatched: cod --workspace",
		"--rerun job-000002",
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
