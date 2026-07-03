package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type patchApprovalEventLine struct {
	Kind         string `json:"kind"`
	Status       string `json:"status"`
	Source       string `json:"source"`
	Digest       string `json:"digest"`
	PreviewCount int    `json:"preview_count"`
}

func Test_Run_prints_patch_approval_event_when_dry_run_previews_patch(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace", root,
		"--format", "events",
		"--dry-run",
		"--replace", "app.txt", "old", "new",
		"Patch", "app", "text",
	}

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
	preview := requireRunEvent(t, out.Bytes(), "patch_preview")
	if preview.Source != "cli" {
		t.Fatalf("patch preview source = %q, want cli", preview.Source)
	}
	event := requirePatchApprovalEvent(t, out.Bytes())
	if event.Status != "previewed" || event.Digest == "" || event.PreviewCount != 1 {
		t.Fatalf("patch approval event = %+v, want previewed digest", event)
	}
}

func Test_Run_prints_patch_approval_event_when_digest_is_approved(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	var previewOut bytes.Buffer
	if err := Run(context.Background(), &previewOut, []string{
		"--workspace", root,
		"--dry-run",
		"--replace", "app.txt", "old", "new",
		"Patch", "app", "text",
	}); err != nil {
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
		"--format", "events",
		"--approve-preview", previewBody.PatchApproval.PreviewDigest,
		"--replace", "app.txt", "old", "new",
		"Patch", "app", "text",
	})
	// Then
	if err != nil {
		t.Fatalf("apply Run returned error: %v\n%s", err, applyOut.String())
	}
	event := requirePatchApprovalEvent(t, applyOut.Bytes())
	if event.Status != "approved" || event.Digest != previewBody.PatchApproval.PreviewDigest || event.PreviewCount != 1 {
		t.Fatalf("patch approval event = %+v, want approved digest", event)
	}
}

func requirePatchApprovalEvent(t *testing.T, body []byte) patchApprovalEventLine {
	t.Helper()
	return requireRunEvent(t, body, "patch_approval")
}

func requireRunEvent(t *testing.T, body []byte, kind string) patchApprovalEventLine {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(body))
	for {
		var event patchApprovalEventLine
		decodeErr := decoder.Decode(&event)
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			t.Fatalf("events output must be JSONL: %v\n%s", decodeErr, string(body))
		}
		if event.Kind == kind {
			return event
		}
	}
	t.Fatalf("events output missing %s:\n%s", kind, string(body))
	return patchApprovalEventLine{}
}
