package subagent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ceoharness/internal/model"
)

func Test_OutputError_Error_handles_nil_receiver(t *testing.T) {
	// Given
	var outputErr *OutputError

	// When
	message := outputErr.Error()

	// Then
	if message != string(OutputErrorKindInvalid) {
		t.Fatalf("Error() = %q, want default invalid output kind", message)
	}
}

func Test_Runner_Run_wraps_invalid_structured_model_output(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(fixedModelClient{text: `{"summary":"bad confidence","confidence":2}`})

	// When
	_, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 512,
	})

	// Then
	var outputErr *OutputError
	if !errors.As(err, &outputErr) {
		t.Fatalf("Run error = %v, want OutputError", err)
	}
	if outputErr.Kind != OutputErrorKindInvalid {
		t.Fatalf("OutputError kind = %q, want invalid output", outputErr.Kind)
	}
	if !strings.Contains(err.Error(), "confidence must be between 0 and 1") {
		t.Fatalf("Run error = %q, want parse reason", err.Error())
	}
}

func Test_Runner_Run_wraps_blank_model_output(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(fixedModelClient{text: " \n\t "})

	// When
	_, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 512,
	})

	// Then
	var outputErr *OutputError
	if !errors.As(err, &outputErr) {
		t.Fatalf("Run error = %v, want OutputError", err)
	}
	if outputErr.Kind != OutputErrorKindEmpty {
		t.Fatalf("OutputError kind = %q, want empty output", outputErr.Kind)
	}
}

type strictOutputModelClient struct {
	text string
}

func (c strictOutputModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text:                    c.text,
		PromptBytes:             len(req.Prompt),
		RequireStructuredOutput: true,
	}, nil
}

func Test_Runner_Run_requires_structured_model_output_when_response_requires_it(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(strictOutputModelClient{text: "loose provider prose"})

	// When
	_, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 512,
	})

	// Then
	var outputErr *OutputError
	if !errors.As(err, &outputErr) {
		t.Fatalf("Run error = %v, want OutputError", err)
	}
	if outputErr.Kind != OutputErrorKindInvalid {
		t.Fatalf("OutputError kind = %q, want invalid output", outputErr.Kind)
	}
	if !strings.Contains(err.Error(), "structured model output required") {
		t.Fatalf("Run error = %q, want structured output reason", err.Error())
	}
}

func Test_RoutingRunner_Run_records_typed_fallback_reason_when_primary_output_is_invalid(t *testing.T) {
	// Given
	runner := NewRoutingRunnerWithConfig(RoutingConfig{
		DefaultClient: fixedModelClient{text: "default response"},
		Clients: map[string]model.Client{
			"scanner": fixedModelClient{text: `{"summary":"bad confidence","confidence":2}`},
		},
		Metadata: map[string]RouteMetadata{
			"scanner": {Source: "command", ProviderName: "cheap"},
		},
		FallbackClient:   fixedModelClient{text: "premium fallback response"},
		FallbackMetadata: RouteMetadata{Source: "command", ProviderName: "premium"},
	})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 512,
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ProviderFallbackReason != string(OutputErrorKindInvalid) {
		t.Fatalf("fallback reason = %q, want invalid model output", result.ProviderFallbackReason)
	}
}
