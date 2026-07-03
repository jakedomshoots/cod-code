package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_applies_cli_patch_when_preview_digest_is_approved(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	var previewOut bytes.Buffer
	previewArgs := []string{"--workspace", root, "--dry-run", "--replace", "app.txt", "old", "new", "Patch", "app", "text"}
	if err := Run(context.Background(), &previewOut, previewArgs); err != nil {
		t.Fatalf("preview Run returned error: %v", err)
	}
	var previewBody struct {
		PatchApproval struct {
			PreviewDigest string `json:"preview_digest"`
		} `json:"patch_approval"`
	}
	if err := json.Unmarshal(previewOut.Bytes(), &previewBody); err != nil {
		t.Fatalf("preview output must be JSON: %v\n%s", err, previewOut.String())
	}
	var applyOut bytes.Buffer

	// When
	err := Run(context.Background(), &applyOut, []string{
		"--workspace", root,
		"--approve-preview", previewBody.PatchApproval.PreviewDigest,
		"--replace", "app.txt", "old", "new",
		"Patch", "app", "text",
	})

	// Then
	if err != nil {
		t.Fatalf("apply Run returned error: %v\n%s", err, applyOut.String())
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	var applyBody struct {
		PatchApproval struct {
			Status         string `json:"status"`
			ApprovedDigest string `json:"approved_digest"`
		} `json:"patch_approval"`
	}
	if err := json.Unmarshal(applyOut.Bytes(), &applyBody); err != nil {
		t.Fatalf("apply output must be JSON: %v\n%s", err, applyOut.String())
	}
	if applyBody.PatchApproval.Status != "approved" || applyBody.PatchApproval.ApprovedDigest != previewBody.PatchApproval.PreviewDigest {
		t.Fatalf("PatchApproval = %+v, want approved digest", applyBody.PatchApproval)
	}
}

func Test_Run_rejects_cli_patch_when_preview_digest_mismatches(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--approve-preview", "bad-digest",
		"--replace", "app.txt", "old", "new",
		"Patch", "app", "text",
	})

	// Then
	if err == nil {
		t.Fatal("expected preview approval mismatch")
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged hello old", string(got))
	}
}
