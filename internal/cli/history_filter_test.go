package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_filters_job_history_when_verdict_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{Task: "Passing job", Verdict: "pass"},
		{Task: "Failing job", Verdict: "fail"},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--verdict", "fail"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			Task    string `json:"task"`
			Verdict string `json:"verdict"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	if body.History[0].Task != "Failing job" || body.History[0].Verdict != "fail" {
		t.Fatalf("history entry = %#v, want failing job", body.History[0])
	}
}

func Test_Run_limits_job_history_when_limit_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	for _, task := range []string{"First job", "Second job", "Third job"} {
		if _, err := store.Append(context.Background(), history.Entry{Task: task, Verdict: "pass"}); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--history", "--limit", "2"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			Task string `json:"task"`
		} `json:"history"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 2 {
		t.Fatalf("history length = %d, want 2", len(body.History))
	}
	if body.History[0].Task != "Second job" || body.History[1].Task != "Third job" {
		t.Fatalf("history tasks = %#v, want latest two", body.History)
	}
}

func Test_Run_filters_job_history_when_since_and_until_flags_are_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{Task: "Old job", Verdict: "pass", CreatedAt: "2026-07-01T09:00:00Z"},
		{Task: "Middle job", Verdict: "pass", CreatedAt: "2026-07-01T12:00:00Z"},
		{Task: "New job", Verdict: "pass", CreatedAt: "2026-07-01T15:00:00Z"},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{
		"--workspace", root,
		"--history",
		"--since", "2026-07-01T10:00:00Z",
		"--until", "2026-07-01T13:00:00Z",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		History []struct {
			Task string `json:"task"`
		} `json:"history"`
		Since string `json:"since,omitempty"`
		Until string `json:"until,omitempty"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.History) != 1 {
		t.Fatalf("history length = %d, want 1", len(body.History))
	}
	if body.History[0].Task != "Middle job" {
		t.Fatalf("Task = %q, want Middle job", body.History[0].Task)
	}
	if body.Since != "2026-07-01T10:00:00Z" || body.Until != "2026-07-01T13:00:00Z" {
		t.Fatalf("range = %q %q, want supplied range", body.Since, body.Until)
	}
}

func Test_Run_returns_error_when_since_flag_is_not_timestamp(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--history", "--since", "not-a-time"})

	// Then
	if err == nil {
		t.Fatal("expected invalid since error")
	}
	if !strings.Contains(err.Error(), "--since must be RFC3339") {
		t.Fatalf("error = %q, want RFC3339 message", err.Error())
	}
}
