package subagent

import (
	"context"
	"testing"

	"ceoharness/internal/model"
)

type metadataCaptureClient struct {
	metadata model.RequestMetadata
}

func (c *metadataCaptureClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.metadata = req.Metadata
	return model.Response{Text: `{"summary":"metadata captured"}`}, nil
}

func Test_Runner_Run_sends_subagent_metadata_to_model_client(t *testing.T) {
	// Given
	client := &metadataCaptureClient{}
	runner := NewRunnerWithModel(client)

	// When
	_, err := runner.Run(context.Background(), TaskPacket{
		Task:        "Inspect docs",
		AgentName:   "scanner",
		Role:        "inspect scope",
		ContextMode: "lean",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if client.metadata.Kind != "subagent" {
		t.Fatalf("Kind = %q, want subagent", client.metadata.Kind)
	}
	if client.metadata.AgentName != "scanner" || client.metadata.AgentRole != "inspect scope" {
		t.Fatalf("metadata = %#v, want scanner inspect scope", client.metadata)
	}
	if client.metadata.ContextMode != "lean" {
		t.Fatalf("ContextMode = %q, want lean", client.metadata.ContextMode)
	}
}
