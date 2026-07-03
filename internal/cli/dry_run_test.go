package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_dry_run_previews_cli_patch_without_writing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{"--workspace", root, "--dry-run", "--replace", "app.txt", "old", "new", "Patch", "app", "text"}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged hello old", string(got))
	}
	if _, err := os.Stat(filepath.Join(root, "ceo-artifacts")); !os.IsNotExist(err) {
		t.Fatalf("ceo-artifacts should not exist after dry-run: %v", err)
	}
	var body struct {
		HistoryPath   string `json:"history_path"`
		JobID         string `json:"job_id"`
		ChangedFiles  []string
		PatchApproval struct {
			Status        string `json:"status"`
			PreviewDigest string `json:"preview_digest"`
			PreviewCount  int    `json:"preview_count"`
		} `json:"patch_approval"`
		PatchResults  []struct{} `json:"patch_results"`
		PatchPreviews []struct {
			Path string `json:"path"`
			Diff string `json:"diff"`
		} `json:"patch_previews"`
		RunLedger struct {
			ChangedFileCount int `json:"changed_file_count"`
		} `json:"run_ledger"`
		RunManifest struct {
			DryRun bool `json:"dry_run"`
		} `json:"run_manifest"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.HistoryPath != "" || body.JobID != "" {
		t.Fatalf("dry-run persisted history path/job id: %q/%q", body.HistoryPath, body.JobID)
	}
	if !body.RunManifest.DryRun {
		t.Fatalf("RunManifest.DryRun = false, want true")
	}
	if len(body.ChangedFiles) != 0 || body.RunLedger.ChangedFileCount != 0 {
		t.Fatalf("dry-run changed files = %+v ledger=%d, want none", body.ChangedFiles, body.RunLedger.ChangedFileCount)
	}
	if len(body.PatchResults) != 0 {
		t.Fatalf("PatchResults length = %d, want 0", len(body.PatchResults))
	}
	if body.PatchApproval.Status != "previewed" || body.PatchApproval.PreviewDigest == "" || body.PatchApproval.PreviewCount != 1 {
		t.Fatalf("PatchApproval = %+v, want preview digest", body.PatchApproval)
	}
	if len(body.PatchPreviews) != 1 || body.PatchPreviews[0].Path != "app.txt" || body.PatchPreviews[0].Diff == "" {
		t.Fatalf("PatchPreviews = %+v, want app.txt diff", body.PatchPreviews)
	}
}

func Test_Run_dry_run_previews_model_patch_without_writing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace",
		root,
		"--dry-run",
		"--apply-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_patch",
		"--",
		"Patch",
		"from",
		"coder",
	}
	t.Setenv("GO_WANT_CLI_MODEL_PATCH", "1")

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged hello old", string(got))
	}
	var body struct {
		PatchApproval struct {
			Status        string `json:"status"`
			PreviewDigest string `json:"preview_digest"`
			PreviewCount  int    `json:"preview_count"`
		} `json:"patch_approval"`
		PatchResults  []struct{} `json:"patch_results"`
		PatchPreviews []struct {
			Path string `json:"path"`
			Diff string `json:"diff"`
		} `json:"patch_previews"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.PatchResults) != 0 {
		t.Fatalf("PatchResults length = %d, want 0", len(body.PatchResults))
	}
	if body.PatchApproval.Status != "previewed" || body.PatchApproval.PreviewDigest == "" || body.PatchApproval.PreviewCount != 1 {
		t.Fatalf("PatchApproval = %+v, want model patch preview digest", body.PatchApproval)
	}
	if len(body.PatchPreviews) != 1 || body.PatchPreviews[0].Path != "app.txt" || body.PatchPreviews[0].Diff == "" {
		t.Fatalf("PatchPreviews = %+v, want model patch preview", body.PatchPreviews)
	}
}
