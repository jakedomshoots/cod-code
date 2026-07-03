package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Workspace_WriteText_writes_file_when_path_is_relative(t *testing.T) {
	// Given
	root := t.TempDir()
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.WriteText(context.Background(), WriteTextRequest{
		Path:    "ceo-artifacts/scanner.md",
		Content: "scanner evidence",
	})

	// Then
	if err != nil {
		t.Fatalf("WriteText returned error: %v", err)
	}
	if result.Path != "ceo-artifacts/scanner.md" {
		t.Fatalf("Path = %q, want ceo-artifacts/scanner.md", result.Path)
	}
	got, err := os.ReadFile(filepath.Join(root, "ceo-artifacts", "scanner.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(got) != "scanner evidence" {
		t.Fatalf("content = %q, want scanner evidence", string(got))
	}
}

func Test_Workspace_WriteText_rejects_path_escape_when_path_traverses_parent(t *testing.T) {
	// Given
	space, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.WriteText(context.Background(), WriteTextRequest{
		Path:    "../outside.md",
		Content: "escape",
	})

	// Then
	if err == nil {
		t.Fatal("expected path escape error")
	}
}

func Test_Workspace_WriteText_rejects_symlink_escape_when_target_is_outside_root(t *testing.T) {
	// Given
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write outside fixture: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.WriteText(context.Background(), WriteTextRequest{
		Path:    "link.txt",
		Content: "escape",
	})

	// Then
	if !errors.Is(err, ErrPathEscapesWorkspace) {
		t.Fatalf("error = %v, want path escape", err)
	}
	got, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatalf("read outside fixture: %v", readErr)
	}
	if string(got) != "keep" {
		t.Fatalf("outside content = %q, want keep", string(got))
	}
}

func Test_Workspace_ReplaceText_replaces_first_match_when_path_is_safe(t *testing.T) {
	// Given
	root := t.TempDir()
	path := filepath.Join(root, "app.txt")
	if err := os.WriteFile(path, []byte("hello old old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.ReplaceText(context.Background(), ReplaceTextRequest{
		Path: "app.txt",
		Old:  "old",
		New:  "new",
	})

	// Then
	if err != nil {
		t.Fatalf("ReplaceText returned error: %v", err)
	}
	if result.Path != "app.txt" {
		t.Fatalf("Path = %q, want app.txt", result.Path)
	}
	if result.Diff == "" {
		t.Fatal("expected diff preview")
	}
	if !strings.Contains(result.Diff, "-hello old old") || !strings.Contains(result.Diff, "+hello new old") {
		t.Fatalf("diff = %q, want old and new content", result.Diff)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new old" {
		t.Fatalf("content = %q, want first match replaced", string(got))
	}
}

func Test_Workspace_ReplaceText_rejects_missing_old_text_when_text_not_found(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.ReplaceText(context.Background(), ReplaceTextRequest{
		Path: "app.txt",
		Old:  "missing",
		New:  "new",
	})

	// Then
	if err == nil {
		t.Fatal("expected text-not-found error")
	}
}

func Test_Workspace_ReplaceText_rejects_symlink_escape_when_target_is_outside_root(t *testing.T) {
	// Given
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write outside fixture: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.ReplaceText(context.Background(), ReplaceTextRequest{
		Path: "link.txt",
		Old:  "old",
		New:  "new",
	})

	// Then
	if !errors.Is(err, ErrPathEscapesWorkspace) {
		t.Fatalf("error = %v, want path escape", err)
	}
	got, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatalf("read outside fixture: %v", readErr)
	}
	if string(got) != "hello old" {
		t.Fatalf("outside content = %q, want unchanged", string(got))
	}
}

func Test_Workspace_CreateText_writes_new_file_when_path_is_safe(t *testing.T) {
	// Given
	root := t.TempDir()
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.CreateText(context.Background(), CreateTextRequest{
		Path:    "docs/notes.md",
		Content: "# Notes\n",
	})

	// Then
	if err != nil {
		t.Fatalf("CreateText returned error: %v", err)
	}
	if result.Path != "docs/notes.md" || result.Diff == "" {
		t.Fatalf("CreateText result = %+v, want path and diff", result)
	}
	got, err := os.ReadFile(filepath.Join(root, "docs", "notes.md"))
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	if string(got) != "# Notes\n" {
		t.Fatalf("content = %q, want created content", string(got))
	}
}

func Test_Workspace_PreviewCreateText_does_not_write_file(t *testing.T) {
	// Given
	root := t.TempDir()
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.PreviewCreateText(context.Background(), CreateTextRequest{
		Path:    "docs/notes.md",
		Content: "# Notes\n",
	})

	// Then
	if err != nil {
		t.Fatalf("PreviewCreateText returned error: %v", err)
	}
	if result.Path != "docs/notes.md" || !strings.Contains(result.Diff, "+# Notes") {
		t.Fatalf("PreviewCreateText result = %+v, want create diff", result)
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "notes.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("created file exists after preview: %v", err)
	}
}

func Test_Workspace_CreateText_rejects_existing_file(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.CreateText(context.Background(), CreateTextRequest{
		Path:    "app.txt",
		Content: "replace",
	})

	// Then
	if !errors.Is(err, ErrFileAlreadyExists) {
		t.Fatalf("error = %v, want ErrFileAlreadyExists", err)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "app.txt"))
	if readErr != nil {
		t.Fatalf("read fixture: %v", readErr)
	}
	if string(got) != "keep" {
		t.Fatalf("content = %q, want unchanged", string(got))
	}
}
