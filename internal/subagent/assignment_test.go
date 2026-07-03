package subagent

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/model"
)

type assignmentPromptClient struct {
	prompt string
}

func (c *assignmentPromptClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.prompt = req.Prompt
	return model.Response{Text: `{"summary":"assignment used"}`, PromptBytes: len(req.Prompt)}, nil
}

func Test_Runner_Run_includes_assignment_in_prompt_and_result(t *testing.T) {
	// Given
	client := &assignmentPromptClient{}
	runner := NewRunnerWithModel(client)

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:           "Fix auth flow",
		AgentName:      "security",
		Role:           "review auth risks",
		Assignment:     "Inspect auth risks only.",
		ContextMode:    "lean",
		AllowedActions: []string{"read_workspace"},
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(client.prompt, "assignment: Inspect auth risks only.") {
		t.Fatalf("prompt = %q, want assignment", client.prompt)
	}
	if result.Assignment != "Inspect auth risks only." {
		t.Fatalf("Assignment = %q, want delegated assignment", result.Assignment)
	}
}
