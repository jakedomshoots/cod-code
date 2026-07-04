package eval

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_RunLocalAgentBenchmark_syntheticCEO_writes_required_evidence_artifact_and_all_required_files(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "ceo-packet"), `#!/bin/sh
while [ "$#" -gt 0 ]; do
  case "$1" in
    --replace)
      path="$2"
      new="$4"
      printf '%s' "$new" > "$path"
      shift 4
      ;;
    *) shift ;;
  esac
done
printf '{"verdict":"pass"}\n'
`)
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Docs task",
		"objective":"Refresh docs.",
		"required_changed_files":["docs/ONE.md","docs/TWO.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:         tasksDir,
		OutputDir:        filepath.Join(root, "benchmark"),
		TimeoutSeconds:   5,
		Agents:           []string{"ceo_harness"},
		CEOHarnessBinary: filepath.Join(binDir, "ceo-packet"),
		BenchmarkTaskID:  "docs-one",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	result := summary.Results[0]
	if result.Status != localAgentStatusPass || result.EvidenceStatus != localAgentEvidenceComplete {
		t.Fatalf("result = %+v, want pass with complete evidence", result)
	}
	requireFile(t, filepath.Join(result.WorkspaceDir, ".omo/evidence/docs-one.md"))
	requireFile(t, filepath.Join(result.WorkspaceDir, "docs/TWO.md"))
	if _, err := os.Stat(result.ScorePath); err != nil {
		t.Fatalf("score path missing: %v", err)
	}
}
