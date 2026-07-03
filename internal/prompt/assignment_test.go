package prompt

import (
	"context"
	"strings"
	"testing"
)

func Test_Build_includes_assignment_when_supplied(t *testing.T) {
	// Given
	req := Request{
		Task:        "Fix auth flow",
		AgentName:   "security",
		Role:        "review auth risks",
		Assignment:  "Inspect auth risks only.",
		ContextMode: "lean",
	}

	// When
	prompt, err := Build(context.Background(), req)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(prompt.Text, "assignment: Inspect auth risks only.") {
		t.Fatalf("prompt = %q, want assignment", prompt.Text)
	}
}
