package model

import (
	"context"
	"testing"
)

func Test_StaticClient_Complete_returns_response_when_prompt_is_supplied(t *testing.T) {
	// Given
	client := NewStaticClient()

	// When
	response, err := client.Complete(context.Background(), Request{
		Prompt: "agent: scanner\nrole: inspect scope",
	})
	// Then
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if response.Text == "" {
		t.Fatal("expected response text")
	}
	if response.PromptBytes == 0 {
		t.Fatal("expected prompt byte count")
	}
}
