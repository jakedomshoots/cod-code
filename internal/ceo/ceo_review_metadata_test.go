package ceo

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/model"
)

type ceoReviewMetadataClient struct {
	metadata model.RequestMetadata
}

func (c *ceoReviewMetadataClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	if strings.Contains(req.Prompt, "candidate_subagents") {
		return model.Response{
			Text:        `{"selected_subagents":["coder"],"summary":"delegate to coder"}`,
			PromptBytes: len(req.Prompt),
		}, nil
	}
	c.metadata = req.Metadata
	return model.Response{
		Text:        `{"recommended_verdict":"pass","summary":"metadata captured"}`,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runtime_RunJob_sends_ceo_review_metadata_to_model_client(t *testing.T) {
	// Given
	client := &ceoReviewMetadataClient{}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Inspect docs",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.Verdict != "pass" {
		t.Fatalf("Verdict = %q, want pass", report.Verdict)
	}
	if client.metadata.Kind != "ceo_review" {
		t.Fatalf("Kind = %q, want ceo_review", client.metadata.Kind)
	}
	if client.metadata.AgentName != "ceo" || client.metadata.AgentRole != "final verdict" {
		t.Fatalf("metadata = %#v, want CEO final verdict", client.metadata)
	}
}

type ceoMetadataSequenceClient struct {
	metadatas []model.RequestMetadata
}

func (c *ceoMetadataSequenceClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.metadatas = append(c.metadatas, req.Metadata)
	if strings.Contains(req.Prompt, "candidate_subagents") {
		return model.Response{
			Text:        `{"selected_subagents":["coder"],"summary":"delegate to coder"}`,
			PromptBytes: len(req.Prompt),
		}, nil
	}
	return model.Response{
		Text:        `{"recommended_verdict":"pass","summary":"metadata captured"}`,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runtime_RunJob_sends_ceo_delegation_metadata_to_model_client(t *testing.T) {
	// Given
	client := &ceoMetadataSequenceClient{}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix bug",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil {
		t.Fatal("expected CEO delegation")
	}
	if len(client.metadatas) == 0 {
		t.Fatal("expected captured metadata")
	}
	first := client.metadatas[0]
	if first.Kind != "ceo_delegation" {
		t.Fatalf("first metadata kind = %q, want ceo_delegation", first.Kind)
	}
	if first.AgentName != "ceo" || first.AgentRole != "delegation planner" {
		t.Fatalf("first metadata = %#v, want CEO delegation planner", first)
	}
}
