package history

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_Store_Append_writes_jsonl_entry_when_workspace_is_set(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	stored, err := store.Append(context.Background(), Entry{
		Task:          "Fix a failing test",
		Verdict:       "pass",
		ChangedFiles:  []string{"ceo-artifacts/scanner.md"},
		SubagentCount: 3,
		CheckCount:    1,
		PatchCount:    0,
	})
	// Then
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if stored.ID != "job-000001" {
		t.Fatalf("ID = %q, want job-000001", stored.ID)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"))
	if err != nil {
		t.Fatalf("read history file: %v", err)
	}
	text := string(got)
	if !strings.Contains(text, `"task":"Fix a failing test"`) {
		t.Fatalf("history = %q, want task", text)
	}
	if !strings.Contains(text, `"id":"job-000001"`) {
		t.Fatalf("history = %q, want id", text)
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatalf("history = %q, want trailing newline", text)
	}
}

func Test_Store_ReadAll_reads_appended_entries_when_history_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"Scan repo", "Patch app"} {
		if _, err := store.Append(context.Background(), Entry{Task: task, Verdict: "pass"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	entries, err := store.ReadAll(context.Background())
	// Then
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries length = %d, want 2", len(entries))
	}
	if entries[0].Task != "Scan repo" || entries[1].Task != "Patch app" {
		t.Fatalf("tasks = %q, %q; want appended order", entries[0].Task, entries[1].Task)
	}
	if entries[0].ID != "job-000001" || entries[1].ID != "job-000002" {
		t.Fatalf("ids = %q, %q; want sequential ids", entries[0].ID, entries[1].ID)
	}
}

func Test_Store_Append_sets_created_at_when_entry_has_no_timestamp(t *testing.T) {
	// Given
	root := t.TempDir()
	fixedTime := time.Date(2026, 7, 1, 12, 30, 0, 123, time.UTC)
	store, err := NewWithClock(root, func() time.Time { return fixedTime })
	if err != nil {
		t.Fatalf("NewWithClock returned error: %v", err)
	}

	// When
	stored, err := store.Append(context.Background(), Entry{
		Task:    "Scan repo",
		Verdict: "pass",
	})
	// Then
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	want := "2026-07-01T12:30:00.000000123Z"
	if stored.CreatedAt != want {
		t.Fatalf("CreatedAt = %q, want %q", stored.CreatedAt, want)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"))
	if err != nil {
		t.Fatalf("read history file: %v", err)
	}
	if !strings.Contains(string(got), `"created_at":"`+want+`"`) {
		t.Fatalf("history = %q, want created_at", string(got))
	}
}

func Test_Store_ReadByVerdict_returns_matching_entries_when_verdict_is_set(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []Entry{
		{Task: "Passing job", Verdict: "pass"},
		{Task: "Failing job", Verdict: "fail"},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	filtered, err := store.ReadByVerdict(context.Background(), "fail")
	// Then
	if err != nil {
		t.Fatalf("ReadByVerdict returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("entries length = %d, want 1", len(filtered))
	}
	if filtered[0].Task != "Failing job" {
		t.Fatalf("Task = %q, want Failing job", filtered[0].Task)
	}
}

func Test_Store_ReadRecent_returns_latest_entries_when_limit_is_set(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"First job", "Second job", "Third job"} {
		if _, err := store.Append(context.Background(), Entry{Task: task, Verdict: "pass"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	entries, err := store.ReadRecent(context.Background(), 2)
	// Then
	if err != nil {
		t.Fatalf("ReadRecent returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries length = %d, want 2", len(entries))
	}
	if entries[0].Task != "Second job" || entries[1].Task != "Third job" {
		t.Fatalf("tasks = %q, %q; want latest two", entries[0].Task, entries[1].Task)
	}
}

func Test_FilterByCreatedAtRange_returns_entries_inside_bounds(t *testing.T) {
	// Given
	entries := []Entry{
		{Task: "Old job", CreatedAt: "2026-07-01T09:00:00Z"},
		{Task: "Middle job", CreatedAt: "2026-07-01T12:00:00Z"},
		{Task: "New job", CreatedAt: "2026-07-01T15:00:00Z"},
		{Task: "Legacy job"},
	}
	since := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	until := time.Date(2026, 7, 1, 13, 0, 0, 0, time.UTC)

	// When
	filtered, err := FilterByCreatedAtRange(entries, TimeRange{Since: since, Until: until})
	// Then
	if err != nil {
		t.Fatalf("FilterByCreatedAtRange returned error: %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("entries length = %d, want 1", len(filtered))
	}
	if filtered[0].Task != "Middle job" {
		t.Fatalf("Task = %q, want Middle job", filtered[0].Task)
	}
}

func Test_FilterByCreatedAtRange_returns_error_when_timestamp_is_invalid(t *testing.T) {
	// Given
	entries := []Entry{{Task: "Broken job", CreatedAt: "not-a-time"}}
	since := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)

	// When
	_, err := FilterByCreatedAtRange(entries, TimeRange{Since: since})

	// Then
	if err == nil {
		t.Fatal("expected invalid timestamp error")
	}
}

func Test_FilterByTaskSubstring_returns_matching_entries_when_query_is_set(t *testing.T) {
	// Given
	entries := []Entry{
		{Task: "Fix checkout retry"},
		{Task: "Refactor parser"},
		{Task: "CHECKOUT smoke"},
	}

	// When
	filtered := FilterByTaskSubstring(entries, "checkout")

	// Then
	if len(filtered) != 2 {
		t.Fatalf("entries length = %d, want 2", len(filtered))
	}
	if filtered[0].Task != "Fix checkout retry" || filtered[1].Task != "CHECKOUT smoke" {
		t.Fatalf("tasks = %#v, want checkout matches", filtered)
	}
}

func Test_Store_FindByID_returns_entry_when_id_exists(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), Entry{Task: "Scan repo", Verdict: "pass"}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	entry, err := store.FindByID(context.Background(), "job-000001")
	// Then
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if entry.Task != "Scan repo" {
		t.Fatalf("Task = %q, want Scan repo", entry.Task)
	}
}
