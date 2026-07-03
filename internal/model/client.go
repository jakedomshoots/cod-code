package model

import (
	"context"
	"errors"
	"strings"
)

var ErrPromptRequired = errors.New("model prompt is required")

type Request struct {
	Prompt   string
	Metadata RequestMetadata
}

type RequestMetadata struct {
	Kind        string
	AgentName   string
	AgentRole   string
	ContextMode string
}

type Response struct {
	Text                    string
	PromptBytes             int
	RequestID               string
	Model                   string
	PromptTokens            int
	CompletionTokens        int
	TotalTokens             int
	CostMicroUSD            int64
	RequireStructuredOutput bool
}

type Client interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

type StaticClient struct{}

func NewStaticClient() StaticClient {
	return StaticClient{}
}

func (c StaticClient) Complete(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return Response{}, ErrPromptRequired
	}
	return Response{
		Text:        "local deterministic model response",
		PromptBytes: len(prompt),
	}, nil
}
