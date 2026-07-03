package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_Workspace_ReadText_reads_bounded_file_content(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello world"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.ReadText(context.Background(), ReadTextRequest{Path: "app.txt", MaxBytes: 5})
	// Then
	if err != nil {
		t.Fatalf("ReadText returned error: %v", err)
	}
	if result.Content != "hello" || !result.Truncated {
		t.Fatalf("ReadText = %+v, want truncated hello", result)
	}
}

func Test_Workspace_ReadText_rejects_path_escape(t *testing.T) {
	// Given
	space, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = space.ReadText(context.Background(), ReadTextRequest{Path: "../secret.txt"})

	// Then
	if !errors.Is(err, ErrPathEscapesWorkspace) {
		t.Fatalf("error = %v, want path escape", err)
	}
}

func Test_Workspace_SearchText_returns_bounded_matches(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "one.txt"), []byte("needle one\nneedle two"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "ignored.txt"), []byte("needle ignored"), 0o644); err != nil {
		t.Fatalf("write ignored fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	result, err := space.SearchText(context.Background(), SearchTextRequest{Query: "needle", MaxMatches: 1})
	// Then
	if err != nil {
		t.Fatalf("SearchText returned error: %v", err)
	}
	if len(result.Matches) != 1 {
		t.Fatalf("match count = %d, want 1", len(result.Matches))
	}
	if result.Matches[0].Path != "one.txt" || result.Matches[0].Line != 1 {
		t.Fatalf("match = %+v, want first non-ignored file match", result.Matches[0])
	}
}
