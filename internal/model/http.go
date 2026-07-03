package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrHTTPConfigInvalid = errors.New("invalid http model config")
	ErrHTTPRequestFailed = errors.New("http model request failed")
	ErrHTTPUnauthorized  = errors.New("http model unauthorized")
	ErrHTTPRateLimited   = errors.New("http model rate limited")
	ErrHTTPUnavailable   = errors.New("http model unavailable")
)

type HTTPConfig struct {
	URL                        string
	Model                      string
	APIKey                     string
	InputCostPerMillionTokens  float64
	OutputCostPerMillionTokens float64
	TimeoutMS                  int
	MaxOutputTokens            int
	ResponseFormat             string
}

type HTTPClient struct {
	client                     *http.Client
	url                        string
	model                      string
	apiKey                     string
	inputCostPerMillionTokens  float64
	outputCostPerMillionTokens float64
	maxOutputTokens            int
	responseFormat             string
}

func NewHTTPClient(cfg HTTPConfig) (HTTPClient, error) {
	trimmedURL := strings.TrimSpace(cfg.URL)
	if trimmedURL == "" {
		return HTTPClient{}, fmt.Errorf("url: %w", ErrHTTPConfigInvalid)
	}
	if _, err := url.ParseRequestURI(trimmedURL); err != nil {
		return HTTPClient{}, fmt.Errorf("url: %w", ErrHTTPConfigInvalid)
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		return HTTPClient{}, fmt.Errorf("model: %w", ErrHTTPConfigInvalid)
	}
	timeout := 60 * time.Second
	if cfg.TimeoutMS > 0 {
		timeout = time.Duration(cfg.TimeoutMS) * time.Millisecond
	}
	return HTTPClient{
		client:                     &http.Client{Timeout: timeout},
		url:                        trimmedURL,
		model:                      model,
		apiKey:                     strings.TrimSpace(cfg.APIKey),
		inputCostPerMillionTokens:  cfg.InputCostPerMillionTokens,
		outputCostPerMillionTokens: cfg.OutputCostPerMillionTokens,
		maxOutputTokens:            cfg.MaxOutputTokens,
		responseFormat:             strings.TrimSpace(cfg.ResponseFormat),
	}, nil
}

func (c HTTPClient) Complete(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return Response{}, ErrPromptRequired
	}
	payload := chatCompletionRequest{
		Model:          c.model,
		MaxTokens:      c.maxOutputTokens,
		ResponseFormat: c.chatResponseFormat(),
		Messages: []chatMessage{
			{Role: "user", Content: req.Prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("marshal chat request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("create chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("post chat request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		content, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return Response{}, fmt.Errorf("read chat error response: %w", readErr)
		}
		return Response{}, fmt.Errorf("chat request failed: %w", newHTTPStatusError(resp.StatusCode, string(content), resp.Header.Get("Retry-After")))
	}
	var decoded chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Response{}, fmt.Errorf("decode chat response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return Response{}, fmt.Errorf("decode chat response: %w", ErrHTTPConfigInvalid)
	}
	return Response{
		Text:                    decoded.Choices[0].Message.Content,
		PromptBytes:             len(req.Prompt),
		RequestID:               resp.Header.Get("X-Request-Id"),
		Model:                   decoded.Model,
		PromptTokens:            decoded.Usage.PromptTokens,
		CompletionTokens:        decoded.Usage.CompletionTokens,
		TotalTokens:             decoded.Usage.TotalTokens,
		CostMicroUSD:            c.costMicroUSD(decoded.Usage),
		RequireStructuredOutput: c.requiresStructuredOutput(),
	}, nil
}

func (c HTTPClient) costMicroUSD(usage chatUsage) int64 {
	cost := float64(usage.PromptTokens)*c.inputCostPerMillionTokens +
		float64(usage.CompletionTokens)*c.outputCostPerMillionTokens
	return int64(math.Round(cost))
}

func (c HTTPClient) chatResponseFormat() *chatResponseFormat {
	if c.responseFormat == "" {
		return nil
	}
	return &chatResponseFormat{Type: c.responseFormat}
}

func (c HTTPClient) requiresStructuredOutput() bool {
	return strings.EqualFold(c.responseFormat, "json_object")
}

type chatCompletionRequest struct {
	Model          string              `json:"model"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
	ResponseFormat *chatResponseFormat `json:"response_format,omitempty"`
	Messages       []chatMessage       `json:"messages"`
}

type chatResponseFormat struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
