package subagent

import (
	"context"
	"testing"

	"ceoharness/internal/model"
)

type fencedJSONModelClient struct{}

func (c fencedJSONModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: "```json\n{\"summary\":\"fenced ok\",\"tool_requests\":[{\"action\":\"read_workspace\",\"path\":\"README.md\"}]}\n```",
	}, nil
}

func Test_Runner_Run_parses_tool_requests_when_model_returns_fenced_json(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(fencedJSONModelClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		AllowedActions:  []string{"read_workspace"},
		MaxContextBytes: 512,
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "fenced ok" {
		t.Fatalf("Summary = %q, want fenced ok", result.Summary)
	}
	if len(result.ToolRequests) != 1 || result.ToolRequests[0].Path != "README.md" {
		t.Fatalf("ToolRequests = %+v, want parsed fenced request", result.ToolRequests)
	}
}
