package history

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_Store_SaveHumanJudgment_reads_judgment_when_job_id_is_valid(t *testing.T) {
	// Given
	root := t.TempDir()
	fixedTime := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	store, err := NewWithClock(root, func() time.Time { return fixedTime })
	if err != nil {
		t.Fatalf("NewWithClock returned error: %v", err)
	}

	// When
	path, err := store.SaveHumanJudgment(context.Background(), HumanJudgment{
		JobID:   "job-000001",
		Verdict: "accept",
		Note:    "Looks good.",
	})
	if err != nil {
		t.Fatalf("SaveHumanJudgment returned error: %v", err)
	}
	got, err := store.ReadHumanJudgment(context.Background(), "job-000001")

	// Then
	if err != nil {
		t.Fatalf("ReadHumanJudgment returned error: %v", err)
	}
	if path != "ceo-artifacts/human-judgments/job-000001.json" {
		t.Fatalf("path = %q, want human judgment path", path)
	}
	if got.JobID != "job-000001" || got.Verdict != "accept" || got.Note != "Looks good." {
		t.Fatalf("judgment = %#v, want saved judgment", got)
	}
	if got.CreatedAt != "2026-07-02T12:00:00Z" {
		t.Fatalf("CreatedAt = %q, want fixed timestamp", got.CreatedAt)
	}
	raw, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatalf("read judgment file: %v", err)
	}
	if !strings.Contains(string(raw), `"verdict":"accept"`) {
		t.Fatalf("judgment file = %q, want verdict", string(raw))
	}
}

func Test_Store_SaveHumanJudgment_rejects_invalid_verdict(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = store.SaveHumanJudgment(context.Background(), HumanJudgment{
		JobID:   "job-000001",
		Verdict: "maybe",
	})

	// Then
	if !errors.Is(err, ErrInvalidHumanJudgment) {
		t.Fatalf("error = %v, want ErrInvalidHumanJudgment", err)
	}
}
