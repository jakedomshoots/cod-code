package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Runtime_RunJob_records_check_fix_patch_counts_in_history(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(checkFixRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("bad"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Repair app",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		CheckFixAttempts:  1,
		CheckCommand: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_check_fix_file",
		},
		CheckEnv: []string{"GO_WANT_CEO_CHECK_FIX=1", "GO_CEO_FIX_TARGET=" + target},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	historyBytes, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"))
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	historyText := string(historyBytes)
	for _, want := range []string{
		`"model_patch_count":1`,
		`"check_fix_count":1`,
	} {
		if !strings.Contains(historyText, want) {
			t.Fatalf("history = %q, want %s", historyText, want)
		}
	}
}
