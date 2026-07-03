package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_previews_model_patch_without_writing(t *testing.T) {
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
		"--preview-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_patch",
		"--",
		"Preview",
		"patch",
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
	if len(body.PatchPreviews) != 1 || body.PatchPreviews[0].Path != "app.txt" || body.PatchPreviews[0].Diff == "" {
		t.Fatalf("PatchPreviews = %+v, want app.txt diff", body.PatchPreviews)
	}
}

func Test_Run_previews_model_create_file_patch_without_writing(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--preview-model-patches",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_create_file_patch",
		"--",
		"Preview",
		"new",
		"file",
	}
	t.Setenv("GO_WANT_CLI_MODEL_CREATE_FILE_PATCH", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "notes.md")); !os.IsNotExist(err) {
		t.Fatalf("created file exists after preview: %v", err)
	}
	var body struct {
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
	if len(body.PatchPreviews) != 1 || body.PatchPreviews[0].Path != "docs/notes.md" || !strings.Contains(body.PatchPreviews[0].Diff, "+# Notes") {
		t.Fatalf("PatchPreviews = %+v, want create file diff", body.PatchPreviews)
	}
}

func Test_Run_rejects_preview_and_apply_model_patches_together(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--apply-model-patches",
		"--preview-model-patches",
		"Patch",
		"app",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil {
		t.Fatal("expected conflicting mode error")
	}
	if !strings.Contains(err.Error(), "choose either model patch preview or model patch application") {
		t.Fatalf("error = %q, want conflicting model patch mode", err.Error())
	}
}
