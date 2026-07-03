package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_Workspace_Brief_returns_bounded_file_index_when_workspace_has_files(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("readme"), 0o644); err != nil {
		t.Fatalf("write README fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("notes"), 0o644); err != nil {
		t.Fatalf("write notes fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "ceo-artifacts"), 0o755); err != nil {
		t.Fatalf("mkdir artifacts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ceo-artifacts", "ignored.md"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write ignored fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	brief, err := space.Brief(context.Background(), BriefRequest{MaxFiles: 2})
	// Then
	if err != nil {
		t.Fatalf("Brief returned error: %v", err)
	}
	if brief.FileCount != 3 {
		t.Fatalf("FileCount = %d, want 3", brief.FileCount)
	}
	if len(brief.Files) != 2 {
		t.Fatalf("Files length = %d, want 2", len(brief.Files))
	}
	if !brief.Truncated {
		t.Fatal("expected truncated brief")
	}
	for _, file := range brief.Files {
		if file.Path == "ceo-artifacts/ignored.md" {
			t.Fatalf("brief included artifact file: %+v", brief.Files)
		}
	}
}

func Test_Workspace_Brief_omits_configured_exclude_patterns(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "generated"), 0o755); err != nil {
		t.Fatalf("mkdir generated: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "generated", "client.go"), []byte("package generated"), 0o644); err != nil {
		t.Fatalf("write generated fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.lock"), []byte("lock"), 0o644); err != nil {
		t.Fatalf("write lock fixture: %v", err)
	}
	space, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	brief, err := space.Brief(context.Background(), BriefRequest{
		ExcludePaths: []string{"generated", "*.lock"},
	})
	// Then
	if err != nil {
		t.Fatalf("Brief returned error: %v", err)
	}
	if brief.FileCount != 1 {
		t.Fatalf("FileCount = %d, want 1", brief.FileCount)
	}
	if len(brief.Files) != 1 || brief.Files[0].Path != "app.go" {
		t.Fatalf("Files = %+v, want only app.go", brief.Files)
	}
}
