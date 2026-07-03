package subagent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ceoharness/internal/model"
)

type failingModelClient struct {
	err error
}

func (c failingModelClient) Complete(context.Context, model.Request) (model.Response, error) {
	return model.Response{}, c.err
}

func Test_RoutingRunner_Run_records_route_metadata_when_agent_matches(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		DefaultMetadata: RouteMetadata{
			Source: "local",
		},
		Clients: map[string]model.Client{
			"scanner": fixedModelClient{text: "scanner routed response"},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "http", ProviderName: "fast"},
		},
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
	if result.ModelSource != "http" || result.ProviderName != "fast" {
		t.Fatalf("metadata = source %q provider %q, want http fast", result.ModelSource, result.ProviderName)
	}
}

func Test_RoutingRunner_Run_uses_fallback_client_when_primary_route_fails(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		Clients: map[string]model.Client{
			"scanner": failingModelClient{err: errors.New("cheap unavailable")},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "command", ProviderName: "cheap"},
		},
		FallbackClient:   fixedModelClient{text: "premium fallback response"},
		FallbackMetadata: RouteMetadata{Source: "command", ProviderName: "premium"},
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
	if result.Summary != "premium fallback response" {
		t.Fatalf("Summary = %q, want fallback response", result.Summary)
	}
	if result.ProviderName != "premium" || result.ProviderFallbackFrom != "cheap" {
		t.Fatalf("provider route = %q fallback_from = %q, want premium from cheap", result.ProviderName, result.ProviderFallbackFrom)
	}
	if result.ProviderFallbackReason != "provider_error" {
		t.Fatalf("fallback reason = %q, want provider_error", result.ProviderFallbackReason)
	}
	if len(result.AttemptErrors) != 1 || !strings.Contains(result.AttemptErrors[0], "cheap unavailable") {
		t.Fatalf("AttemptErrors = %#v, want primary error", result.AttemptErrors)
	}
}

func Test_RoutingRunner_Run_records_typed_fallback_reason_when_primary_command_times_out(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		Clients: map[string]model.Client{
			"scanner": failingModelClient{err: &model.CommandError{Kind: model.CommandErrorKindTimeout}},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "command", ProviderName: "cheap"},
		},
		FallbackClient:   fixedModelClient{text: "premium fallback response"},
		FallbackMetadata: RouteMetadata{Source: "command", ProviderName: "premium"},
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
	if result.ProviderFallbackReason != string(model.CommandErrorKindTimeout) {
		t.Fatalf("fallback reason = %q, want command timeout", result.ProviderFallbackReason)
	}
}

func Test_RoutingRunner_Run_uses_fallback_client_when_primary_confidence_is_low(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		Clients: map[string]model.Client{
			"scanner": fixedModelClient{text: `{"summary":"cheap unsure","confidence":0.2}`},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "command", ProviderName: "cheap"},
		},
		FallbackClient:   fixedModelClient{text: `{"summary":"premium sure","confidence":0.9}`},
		FallbackMetadata: RouteMetadata{Source: "command", ProviderName: "premium"},
		MinConfidence:    0.6,
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
	if result.Summary != "premium sure" {
		t.Fatalf("Summary = %q, want fallback summary", result.Summary)
	}
	if result.ProviderName != "premium" || result.ProviderFallbackFrom != "cheap" || result.ProviderFallbackReason != "low_confidence" {
		t.Fatalf("provider route = %+v, want premium fallback from cheap for low confidence", result)
	}
	if result.Confidence == nil || *result.Confidence != 0.9 {
		t.Fatalf("Confidence = %v, want fallback confidence 0.9", result.Confidence)
	}
	if len(result.AttemptErrors) != 1 || !strings.Contains(result.AttemptErrors[0], "confidence 0.20 below minimum 0.60") {
		t.Fatalf("AttemptErrors = %#v, want low confidence error", result.AttemptErrors)
	}
}

func Test_RoutingRunner_Run_marks_low_confidence_result_failed_when_no_fallback_exists(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		Clients: map[string]model.Client{
			"scanner": fixedModelClient{text: `{"summary":"cheap unsure","confidence":0.2}`},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "command", ProviderName: "cheap"},
		},
		MinConfidence: 0.6,
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
	if result.Status != "fail" {
		t.Fatalf("Status = %q, want fail", result.Status)
	}
	if len(result.AttemptErrors) != 1 || !strings.Contains(result.AttemptErrors[0], "confidence 0.20 below minimum 0.60") {
		t.Fatalf("AttemptErrors = %#v, want low confidence error", result.AttemptErrors)
	}
}
