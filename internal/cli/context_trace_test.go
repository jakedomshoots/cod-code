package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_prints_context_trace_for_saved_job_when_context_command_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("OPENAI_API_KEY=sk-proj-secret"), 0o644); err != nil {
		t.Fatalf("write secret fixture: %v", err)
	}
	var runOut bytes.Buffer
	if err := Run(context.Background(), &runOut, []string{
		"--workspace", root,
		"--workspace-brief-exclude", ".env",
		"Fix", "checkout",
	}); err != nil {
		t.Fatalf("initial Run returned error: %v\n%s", err, runOut.String())
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"context", "--workspace", root, "latest"})

	// Then
	if err != nil {
		t.Fatalf("context trace Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		ContextTrace struct {
			JobID  string `json:"job_id"`
			Source string `json:"source"`
			Agents []struct {
				AgentName       string `json:"agent_name"`
				Role            string `json:"role"`
				TaskSummary     string `json:"task_summary"`
				BudgetUnit      string `json:"budget_unit"`
				MaxContextBytes int    `json:"max_context_bytes"`
				ContextBytes    int    `json:"context_bytes"`
				WorkspaceBrief  struct {
					FileCount      int  `json:"file_count"`
					ShownFileCount int  `json:"shown_file_count"`
					Bytes          int  `json:"bytes"`
					Truncated      bool `json:"truncated"`
				} `json:"workspace_brief"`
				PriorFindings struct {
					Count int `json:"count"`
					Bytes int `json:"bytes"`
				} `json:"prior_findings"`
				ExcludedContent struct {
					WorkspaceExcludes []string `json:"workspace_excludes"`
				} `json:"excluded_content"`
			} `json:"agents"`
		} `json:"context_trace"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.ContextTrace.JobID != "job-000001" || body.ContextTrace.Source != "report_snapshot" {
		t.Fatalf("context trace = %+v, want saved report snapshot", body.ContextTrace)
	}
	if len(body.ContextTrace.Agents) != 3 {
		t.Fatalf("agents = %+v, want default subagent traces", body.ContextTrace.Agents)
	}
	first := body.ContextTrace.Agents[0]
	if first.AgentName != "scanner" || first.Role == "" || first.TaskSummary != "Fix checkout" {
		t.Fatalf("first agent = %+v, want scanner task trace", first)
	}
	if first.BudgetUnit != "bytes" || first.MaxContextBytes == 0 || first.ContextBytes == 0 {
		t.Fatalf("first agent = %+v, want byte budget metadata", first)
	}
	if first.WorkspaceBrief.FileCount != 1 || first.WorkspaceBrief.Bytes == 0 {
		t.Fatalf("workspace brief = %+v, want compact brief metadata", first.WorkspaceBrief)
	}
	if len(first.ExcludedContent.WorkspaceExcludes) != 1 || first.ExcludedContent.WorkspaceExcludes[0] != ".env" {
		t.Fatalf("excluded content = %+v, want .env exclude", first.ExcludedContent)
	}
	if body.ContextTrace.Agents[1].PriorFindings.Count == 0 || body.ContextTrace.Agents[1].PriorFindings.Bytes == 0 {
		t.Fatalf("coder prior findings = %+v, want prior finding counts", body.ContextTrace.Agents[1].PriorFindings)
	}
}

func Test_Run_context_trace_marks_truncation_when_saved_job_used_tiny_budget(t *testing.T) {
	// Given
	root := t.TempDir()
	var runOut bytes.Buffer
	if err := Run(context.Background(), &runOut, []string{
		"--workspace", root,
		"--max-context-bytes", "12",
		strings.Repeat("tiny budget task ", 10),
	}); err != nil {
		t.Fatalf("initial Run returned error: %v\n%s", err, runOut.String())
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--context-trace", "latest"})

	// Then
	if err != nil {
		t.Fatalf("context trace Run returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `"context_truncated": true`) {
		t.Fatalf("trace output = %s, want truncation marker", out.String())
	}
	if !strings.Contains(out.String(), `"max_context_bytes": 12`) {
		t.Fatalf("trace output = %s, want tiny budget", out.String())
	}
}

func Test_Run_context_trace_redacts_secrets_and_omits_repo_file_contents(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main\nconst repoNeedle = \"full repo content must stay out\""), 0o644); err != nil {
		t.Fatalf("write app fixture: %v", err)
	}
	var runOut bytes.Buffer
	if err := Run(context.Background(), &runOut, []string{
		"--workspace", root,
		"Investigate", "OPENAI_API_KEY=sk-proj-secret1234567890",
	}); err != nil {
		t.Fatalf("initial Run returned error: %v\n%s", err, runOut.String())
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--context-trace", "latest",
		"--format", "text",
	})

	// Then
	if err != nil {
		t.Fatalf("context trace Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	if strings.Contains(text, "sk-proj-secret") || strings.Contains(text, "OPENAI_API_KEY=") {
		t.Fatalf("trace leaked secret:\n%s", text)
	}
	if strings.Contains(text, "full repo content must stay out") {
		t.Fatalf("trace leaked repo content:\n%s", text)
	}
	if !strings.Contains(text, "[redacted_secret]") {
		t.Fatalf("trace text = %s, want redaction marker", text)
	}
}
