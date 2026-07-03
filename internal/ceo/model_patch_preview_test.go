package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Runtime_RunJob_previews_coder_model_patch_without_writing(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                "Preview app text patch",
		WorkspaceDir:        root,
		PreviewModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged hello old", string(got))
	}
	if len(report.PatchResults) != 0 {
		t.Fatalf("PatchResults length = %d, want 0", len(report.PatchResults))
	}
	if len(report.PatchPreviews) != 1 {
		t.Fatalf("PatchPreviews length = %d, want 1", len(report.PatchPreviews))
	}
	if report.PatchPreviews[0].Path != "app.txt" || report.PatchPreviews[0].Diff == "" {
		t.Fatalf("PatchPreviews[0] = %+v, want app.txt diff", report.PatchPreviews[0])
	}
}

func Test_Runtime_RunJob_rejects_preview_and_apply_together(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task:                "Conflicting model patch modes",
		ApplyModelPatches:   true,
		PreviewModelPatches: true,
	})

	// Then
	if err == nil {
		t.Fatal("expected conflicting mode error")
	}
	if !strings.Contains(err.Error(), "choose either model patch preview or model patch application") {
		t.Fatalf("error = %q, want conflicting model patch mode", err.Error())
	}
}
