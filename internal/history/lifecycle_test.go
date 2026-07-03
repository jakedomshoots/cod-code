package history

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Store_Append_round_trips_lifecycle_fields_when_present(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = store.Append(context.Background(), Entry{
		Task:           "Fix a failing test",
		Verdict:        "pass",
		LifecycleState: "passed",
		LifecycleEvents: []LifecycleEvent{
			{Index: 1, State: "created", Summary: "job created"},
			{Index: 2, State: "passed", PreviousState: "reviewing", Summary: "CEO final verdict passed"},
		},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// Then
	entries, err := store.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if entries[0].LifecycleState != "passed" {
		t.Fatalf("LifecycleState = %q, want passed", entries[0].LifecycleState)
	}
	if len(entries[0].LifecycleEvents) != 2 {
		t.Fatalf("LifecycleEvents length = %d, want 2", len(entries[0].LifecycleEvents))
	}
	if entries[0].LifecycleEvents[1].PreviousState != "reviewing" {
		t.Fatalf("PreviousState = %q, want reviewing", entries[0].LifecycleEvents[1].PreviousState)
	}
}

func Test_Store_ReadAll_reads_legacy_history_without_lifecycle_fields(t *testing.T) {
	// Given
	root := t.TempDir()
	path := filepath.Join(root, JobLogPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create history dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"id":"job-000001","task":"Legacy job","verdict":"pass"}`+"\n"), 0o644); err != nil {
		t.Fatalf("write legacy history: %v", err)
	}
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	entries, err := store.ReadAll(context.Background())

	// Then
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries length = %d, want 1", len(entries))
	}
	if entries[0].LifecycleState != "" {
		t.Fatalf("LifecycleState = %q, want empty legacy state", entries[0].LifecycleState)
	}
	if len(entries[0].LifecycleEvents) != 0 {
		t.Fatalf("LifecycleEvents = %+v, want none for legacy history", entries[0].LifecycleEvents)
	}
}
