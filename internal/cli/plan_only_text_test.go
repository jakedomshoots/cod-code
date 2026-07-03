package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_plan_only_prints_text_preview_when_text_format_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--plan-only", "--format", "text", "--continue-job", "job-000001"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	body := out.String()
	for _, want := range []string{
		"Plan-only preview",
		"Task: Fix auth bug",
		"Owner: coder",
		"Continuation: job-000001 saved_delegation=true planned=3 reusable=3",
		"Subagents: coder, security, reviewer",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("text plan preview missing %q:\n%s", want, body)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(body), "{") {
		t.Fatalf("text plan preview should not be JSON:\n%s", body)
	}
}

func Test_Run_plan_only_rejects_events_format(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--plan-only", "--format", "events", "Fix", "auth", "bug"})

	// Then
	if err == nil {
		t.Fatal("expected plan-only events format error")
	}
	if !strings.Contains(err.Error(), "plan-only does not support --format events") {
		t.Fatalf("error = %q, want plan-only events guidance", err.Error())
	}
}
