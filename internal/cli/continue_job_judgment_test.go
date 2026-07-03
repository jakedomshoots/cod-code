package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_rejects_continue_job_when_human_judgment_rejected_source_job(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	if err := Run(context.Background(), &bytes.Buffer{}, []string{
		"--workspace", root,
		"--judge-job", "job-000001",
		"--human-verdict", "reject",
		"--judgment-note", "Evidence missing.",
	}); err != nil {
		t.Fatalf("judgment Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--continue-job", "job-000001"})

	// Then
	if err == nil {
		t.Fatal("expected rejected human judgment to block continue-job")
	}
	if !strings.Contains(err.Error(), "human judgment rejected job job-000001") {
		t.Fatalf("error = %q, want rejected judgment guidance", err.Error())
	}
}

func Test_Run_allows_continue_job_when_human_judgment_accepted_source_job(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	if err := Run(context.Background(), &bytes.Buffer{}, []string{
		"--workspace", root,
		"--judge-job", "job-000001",
		"--human-verdict", "accept",
	}); err != nil {
		t.Fatalf("judgment Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--continue-job", "job-000001"})
	// Then
	if err != nil {
		t.Fatalf("continue Run returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"job_id": "job-000002"`) {
		t.Fatalf("continue output = %s, want new job", out.String())
	}
}
