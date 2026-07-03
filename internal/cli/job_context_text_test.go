package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ceoharness/internal/ceo"
	"ceoharness/internal/checkrunner"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_Run_prints_compact_job_context_text_when_format_text_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	saveTextJobContext(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--job-context", "job-000001",
		"--format", "text",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{
		"Job context: job-000001",
		"Verdict: needs_input",
		"Task: Fix checkout",
		"Next: answer subagent questions",
		"Question: Which package should I change?",
		"Changed: app.go, README.md",
		"Failed check: go test ./... exit=1",
		"Subagent: scanner [needs_input] Need target package",
		"CEO: Blocked on missing package",
		"--resume job-000001 --answer",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("job context text missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_rejects_job_context_events_format(t *testing.T) {
	// Given
	root := t.TempDir()
	saveTextJobContext(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--job-context", "job-000001",
		"--format", "events",
	})

	// Then
	if err == nil {
		t.Fatal("expected events format error")
	}
	if !strings.Contains(err.Error(), "only available for run reports") {
		t.Fatalf("error = %q, want run report guidance", err.Error())
	}
}

func saveTextJobContext(t *testing.T, root string) {
	t.Helper()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), history.Entry{
		Task:                    "Fix checkout",
		Verdict:                 "needs_input",
		ChangedFiles:            []string{"app.go"},
		ExecutionPlanNextAction: "answer subagent questions",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	report := ceo.Report{
		JobID:   "job-000001",
		Verdict: "needs_input",
		JobPacket: jobpacket.Packet{
			Task:        "Fix checkout",
			TaskProfile: jobpacket.TaskProfile{Kind: "coding", RiskLevel: "medium"},
		},
		Resume: &ceo.ResumeContext{Questions: []string{"Which package should I change?"}},
		SubagentResults: []subagent.Result{{
			AgentName: "scanner",
			Status:    "needs_input",
			Summary:   "Need target package",
			Questions: []string{"Which package should I change?"},
		}},
		ChangedFiles: []string{"app.go", "README.md"},
		CheckResults: []checkrunner.Result{{
			Argv:     []string{"go", "test", "./..."},
			Status:   "fail",
			ExitCode: 1,
			Stderr:   "checkout panic",
		}},
		ExecutionPlan: ceo.ExecutionPlan{NextAction: "answer subagent questions"},
		CEOReview:     &ceo.CEOReview{Summary: "Blocked on missing package"},
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", payload); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
}
