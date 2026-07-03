package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_filters_job_history_when_task_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"Fix checkout retry", "Refactor parser", "CHECKOUT smoke"} {
		if _, err := store.Append(context.Background(), history.Entry{Task: task, Verdict: "pass"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--task", "checkout"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			Task string `json:"task"`
		} `json:"history"`
		TaskFilter string `json:"task_filter,omitempty"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.TaskFilter != "checkout" {
		t.Fatalf("TaskFilter = %q, want checkout", body.TaskFilter)
	}
	if len(body.History) != 2 {
		t.Fatalf("history length = %d, want 2", len(body.History))
	}
	if body.History[0].Task != "Fix checkout retry" || body.History[1].Task != "CHECKOUT smoke" {
		t.Fatalf("history tasks = %#v, want checkout matches", body.History)
	}
}
