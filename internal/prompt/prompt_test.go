package prompt

import (
	"context"
	"strings"
	"testing"
)

func Test_Build_returns_compact_prompt_with_role_when_budget_allows(t *testing.T) {
	// Given
	req := Request{
		Task:           "Fix a failing test",
		AgentName:      "scanner",
		Role:           "inspect scope",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace", "search_workspace"},
		MaxBytes:       512,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "role: inspect scope") {
		t.Fatalf("prompt = %q, want role", prompt.Text)
	}
	if !strings.Contains(prompt.Text, "allowed_actions: read_workspace, search_workspace") {
		t.Fatalf("prompt = %q, want allowed actions", prompt.Text)
	}
	if !strings.Contains(prompt.Text, "response_contract:") {
		t.Fatalf("prompt = %q, want response contract", prompt.Text)
	}
	if !strings.Contains(prompt.Text, `"status":"pass|fail|needs_input"`) || !strings.Contains(prompt.Text, `"confidence":0.0`) {
		t.Fatalf("prompt = %q, want compact structured output contract", prompt.Text)
	}
	if strings.Contains(prompt.Text, "tool_request_format") {
		t.Fatalf("prompt = %q, want no duplicate tool request format", prompt.Text)
	}
	if !strings.Contains(prompt.Text, "do not ask permission") || !strings.Contains(prompt.Text, `"action":"read_workspace"`) {
		t.Fatalf("prompt = %q, want concise tool rules", prompt.Text)
	}
	if prompt.Bytes != len(prompt.Text) {
		t.Fatalf("Bytes = %d, want text length", prompt.Bytes)
	}
	if prompt.Truncated {
		t.Fatal("did not expect truncation")
	}
}

func Test_Build_includes_tool_results_when_feedback_is_supplied(t *testing.T) {
	// Given
	req := Request{
		Task:           "Use tool output",
		AgentName:      "scanner",
		Role:           "inspect scope",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace"},
		ToolResults:    `{"tool_results":[{"action":"read_workspace","status":"pass","output":"hello needle"}]}`,
		MaxBytes:       512,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "tool_results:") || !strings.Contains(prompt.Text, "hello needle") {
		t.Fatalf("prompt = %q, want tool results", prompt.Text)
	}
}

func Test_Build_includes_workspace_brief_when_supplied(t *testing.T) {
	// Given
	req := Request{
		Task:           "Fix workspace bug",
		AgentName:      "scanner",
		Role:           "inspect scope",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace"},
		WorkspaceBrief: "files: app.go, README.md",
		MaxBytes:       512,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "workspace_brief: files: app.go, README.md") {
		t.Fatalf("prompt = %q, want workspace brief", prompt.Text)
	}
}

func Test_Build_includes_prior_findings_when_supplied(t *testing.T) {
	// Given
	req := Request{
		Task:           "Fix workspace bug",
		AgentName:      "coder",
		Role:           "apply bounded changes",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace"},
		PriorFindings:  "- scanner(pass): app.go has failing handler",
		MaxBytes:       512,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "prior_findings: - scanner(pass): app.go has failing handler") {
		t.Fatalf("prompt = %q, want prior findings", prompt.Text)
	}
}

func Test_Build_truncates_context_when_budget_is_exceeded(t *testing.T) {
	// Given
	req := Request{
		Task:        strings.Repeat("a", 200),
		AgentName:   "scanner",
		Role:        "inspect scope",
		ContextMode: "lean",
		MaxBytes:    80,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "response_contract:") {
		t.Fatalf("prompt = %q, want response contract preserved", prompt.Text)
	}
	if !strings.Contains(prompt.Text, "task:\n"+strings.Repeat("a", 80)) {
		t.Fatalf("prompt = %q, want task context capped to budget", prompt.Text)
	}
	if !prompt.Truncated {
		t.Fatal("expected context truncation")
	}
}

func Test_Build_preserves_task_when_optional_context_is_truncated(t *testing.T) {
	// Given
	req := Request{
		Task:           "Fix the auth callback panic",
		AgentName:      "coder",
		Role:           "apply bounded changes",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace"},
		WorkspaceBrief: strings.Repeat("workspace detail ", 40),
		PriorFindings:  strings.Repeat("prior finding ", 40),
		ToolResults:    strings.Repeat("tool result ", 40),
		MaxBytes:       180,
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !prompt.Truncated {
		t.Fatal("expected context truncation")
	}
	if !strings.Contains(prompt.Text, "task:\nFix the auth callback panic") {
		t.Fatalf("prompt = %q, want task preserved under truncation", prompt.Text)
	}
	if !strings.Contains(prompt.Text, "response_contract:") {
		t.Fatalf("prompt = %q, want response contract preserved", prompt.Text)
	}
}
