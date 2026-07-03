package subagent

import (
	"context"
	"testing"

	"ceoharness/internal/model"
)

func Test_RoutingRunner_Run_uses_requested_provider_client(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		ProviderClients: map[string]model.Client{
			"premium": fixedModelClient{text: "premium response"},
		},
		ProviderMetadata: map[string]RouteMetadata{
			"premium": {Source: "command", ProviderName: "premium"},
		},
	})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Review checkout UX",
		AgentName:       "ux_reviewer",
		Role:            "review UX",
		ProviderName:    "premium",
		ContextMode:     "lean",
		MaxContextBytes: 256,
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "premium response" || result.ProviderName != "premium" {
		t.Fatalf("result = %#v, want premium provider response", result)
	}
}
