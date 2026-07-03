package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Workspace_RollbackReplaceText_restores_simple_replace_result(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	result, err := space.ReplaceText(context.Background(), ReplaceTextRequest{
		Path: "app.txt",
		Old:  "old",
		New:  "new",
	})
	if err != nil {
		t.Fatalf("ReplaceText returned error: %v", err)
	}

	// When
	rollback, err := space.RollbackReplaceText(context.Background(), result)
	// Then
	if err != nil {
		t.Fatalf("RollbackReplaceText returned error: %v", err)
	}
	if rollback.Path != "app.txt" || rollback.Diff == "" {
		t.Fatalf("rollback result = %+v, want path and diff", rollback)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want restored", string(got))
	}
}

func Test_Workspace_RollbackReplaceText_restores_trailing_newline_result(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	result, err := space.ReplaceText(context.Background(), ReplaceTextRequest{
		Path: "app.txt",
		Old:  "old",
		New:  "new",
	})
	if err != nil {
		t.Fatalf("ReplaceText returned error: %v", err)
	}

	// When
	_, err = space.RollbackReplaceText(context.Background(), result)
	// Then
	if err != nil {
		t.Fatalf("RollbackReplaceText returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "hello old\n" {
		t.Fatalf("content = %q, want restored with newline", string(got))
	}
}

func Test_Workspace_RollbackReplaceText_rejects_unsupported_diff(t *testing.T) {
	// Given
	space, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.RollbackReplaceText(context.Background(), ReplaceTextResult{
		Path: "app.txt",
		Diff: "not a harness diff",
	})

	// Then
	if err == nil {
		t.Fatal("expected unsupported rollback diff error")
	}
}
