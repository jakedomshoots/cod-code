package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_Run_writes_artifacts_to_external_artifact_root(t *testing.T) {
	var out bytes.Buffer
	root := t.TempDir()
	artifactRoot := filepath.Join(t.TempDir(), "runtime")

	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--artifact-root", artifactRoot,
		"Fix", "a", "failing", "test",
	})

	requireRunSuccess(t, err, out.String())
	if _, statErr := os.Stat(filepath.Join(root, "ceo-artifacts")); !os.IsNotExist(statErr) {
		t.Fatalf("workspace ceo-artifacts should not exist: %v", statErr)
	}
	requireExistingFile(t, filepath.Join(artifactRoot, "ceo-artifacts", "scanner.md"))
	requireExistingFile(t, filepath.Join(artifactRoot, "ceo-artifacts", "jobs.jsonl"))
	requireExistingFile(t, filepath.Join(artifactRoot, "ceo-artifacts", "jobs", "job-000001.json"))

	var body struct {
		ChangedFiles []string `json:"changed_files"`
		JobID        string   `json:"job_id"`
		RunLedger    struct {
			ChangedFileCount int      `json:"changed_file_count"`
			ChangedFiles     []string `json:"changed_files"`
		} `json:"run_ledger"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobID != "job-000001" {
		t.Fatalf("JobID = %q, want job-000001", body.JobID)
	}
	if len(body.ChangedFiles) != 0 || body.RunLedger.ChangedFileCount != 0 || len(body.RunLedger.ChangedFiles) != 0 {
		t.Fatalf("changed files = %+v ledger=%+v, want external artifacts excluded", body.ChangedFiles, body.RunLedger)
	}
}

func requireExistingFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}
