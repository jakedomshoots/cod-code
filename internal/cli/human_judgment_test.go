package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func Test_Run_writes_human_judgment_when_verdict_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--judge-job", "latest",
		"--human-verdict", "accept",
		"--judgment-note", "Ship it.",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JudgmentPath string `json:"judgment_path"`
		Judgment     struct {
			JobID   string `json:"job_id"`
			Verdict string `json:"verdict"`
			Note    string `json:"note"`
		} `json:"judgment"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JudgmentPath != "ceo-artifacts/human-judgments/job-000001.json" {
		t.Fatalf("JudgmentPath = %q, want sidecar path", body.JudgmentPath)
	}
	if body.Judgment.JobID != "job-000001" || body.Judgment.Verdict != "accept" || body.Judgment.Note != "Ship it." {
		t.Fatalf("judgment = %#v, want accepted latest job", body.Judgment)
	}
}

func Test_Run_reads_human_judgment_when_verdict_is_omitted(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	if err := Run(context.Background(), &bytes.Buffer{}, []string{
		"--workspace", root,
		"--judge-job", "job-000001",
		"--human-verdict", "reject",
		"--judgment-note", "Missing test evidence.",
	}); err != nil {
		t.Fatalf("seed judgment returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--judge-job", "last"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Judgment struct {
			JobID   string `json:"job_id"`
			Verdict string `json:"verdict"`
			Note    string `json:"note"`
		} `json:"judgment"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Judgment.JobID != "job-000001" || body.Judgment.Verdict != "reject" || body.Judgment.Note != "Missing test evidence." {
		t.Fatalf("judgment = %#v, want rejected saved judgment", body.Judgment)
	}
}

func Test_Run_rejects_human_verdict_without_judge_job(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--human-verdict", "accept"})

	// Then
	if err == nil {
		t.Fatal("expected human verdict without judge job error")
	}
	if !strings.Contains(err.Error(), "--human-verdict requires --judge-job") {
		t.Fatalf("error = %q, want judge-job guidance", err.Error())
	}
}
