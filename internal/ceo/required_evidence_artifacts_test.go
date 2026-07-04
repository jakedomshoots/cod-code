package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Runtime_RunJob_writes_required_evidence_artifact_from_task(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Patch app text.\n" +
			"Required evidence artifacts: .omo/evidence/app.md.\n",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, ".omo", "evidence", "app.md"))
	if err != nil {
		t.Fatalf("read evidence artifact: %v", err)
	}
	body := string(content)
	if !strings.Contains(body, "app.txt") || !strings.Contains(body, "Patch app text.") {
		t.Fatalf("evidence artifact = %q, want task and changed file", body)
	}
	if !containsString(report.ChangedFiles, ".omo/evidence/app.md") {
		t.Fatalf("ChangedFiles = %+v, want required evidence artifact", report.ChangedFiles)
	}
}

func Test_requiredEvidenceArtifactPaths_parses_comma_list(t *testing.T) {
	// When
	paths := requiredEvidenceArtifactPaths("Required evidence artifacts: .omo/evidence/a.md, .omo/evidence/b.md.")

	// Then
	if len(paths) != 2 || paths[0] != ".omo/evidence/a.md" || paths[1] != ".omo/evidence/b.md" {
		t.Fatalf("paths = %+v, want two evidence markdown paths", paths)
	}
}
