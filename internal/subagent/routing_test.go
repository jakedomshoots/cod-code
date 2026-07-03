package subagent

import (
	"context"
	"testing"

	"ceoharness/internal/model"
)

type fixedModelClient struct {
	text string
}

func (c fixedModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text:        c.text,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_RoutingRunner_Run_uses_agent_specific_client_when_agent_matches(t *testing.T) {
	// Given
	runner := NewRoutingRunner(fixedModelClient{text: "default response"}, map[string]model.Client{
		"scanner": fixedModelClient{text: "scanner routed response"},
	})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Fix a failing test",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 256,
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "scanner routed response" {
		t.Fatalf("Summary = %q, want scanner routed response", result.Summary)
	}
}

func Test_RoutingRunner_Run_uses_default_client_when_agent_has_no_route(t *testing.T) {
	// Given
	runner := NewRoutingRunner(fixedModelClient{text: "default response"}, map[string]model.Client{
		"scanner": fixedModelClient{text: "scanner routed response"},
	})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Fix a failing test",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		MaxContextBytes: 256,
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "default response" {
		t.Fatalf("Summary = %q, want default response", result.Summary)
	}
}
