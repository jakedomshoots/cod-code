package model

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_HTTPClient_Complete_posts_chat_completion_request(t *testing.T) {
	// Given
	var gotAuth string
	var gotModel string
	var gotPrompt string
	var gotMaxTokens int
	var gotResponseFormat string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body struct {
			Model          string `json:"model"`
			MaxTokens      int    `json:"max_tokens"`
			ResponseFormat struct {
				Type string `json:"type"`
			} `json:"response_format"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotModel = body.Model
		gotMaxTokens = body.MaxTokens
		gotResponseFormat = body.ResponseFormat.Type
		if len(body.Messages) != 1 {
			t.Fatalf("messages length = %d, want 1", len(body.Messages))
		}
		gotPrompt = body.Messages[0].Content
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Id", "req-http-123")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"model":"served-model","choices":[{"message":{"content":"http model response"}}],"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	client, err := NewHTTPClient(HTTPConfig{
		URL:                        server.URL,
		Model:                      "test-model",
		APIKey:                     "test-token",
		InputCostPerMillionTokens:  2,
		OutputCostPerMillionTokens: 8,
		MaxOutputTokens:            64,
		ResponseFormat:             "json_object",
	})
	if err != nil {
		t.Fatalf("NewHTTPClient returned error: %v", err)
	}

	// When
	response, err := client.Complete(context.Background(), Request{Prompt: "hello model"})
	// Then
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if response.Text != "http model response" {
		t.Fatalf("Text = %q, want http model response", response.Text)
	}
	if response.RequestID != "req-http-123" || response.Model != "served-model" {
		t.Fatalf("response metadata = request %q model %q, want request id and served model", response.RequestID, response.Model)
	}
	if response.PromptTokens != 11 || response.CompletionTokens != 7 || response.TotalTokens != 18 {
		t.Fatalf("usage = prompt %d completion %d total %d, want 11/7/18", response.PromptTokens, response.CompletionTokens, response.TotalTokens)
	}
	if response.CostMicroUSD != 78 {
		t.Fatalf("CostMicroUSD = %d, want 78", response.CostMicroUSD)
	}
	if !response.RequireStructuredOutput {
		t.Fatalf("RequireStructuredOutput = false, want true for json_object response format")
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("Authorization = %q, want bearer token", gotAuth)
	}
	if gotModel != "test-model" || gotPrompt != "hello model" {
		t.Fatalf("request model/prompt = %q/%q, want configured model and prompt", gotModel, gotPrompt)
	}
	if gotMaxTokens != 64 {
		t.Fatalf("max_tokens = %d, want 64", gotMaxTokens)
	}
	if gotResponseFormat != "json_object" {
		t.Fatalf("response_format.type = %q, want json_object", gotResponseFormat)
	}
}

func Test_HTTPClient_Complete_uses_configured_timeout(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"late response"}}]}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	client, err := NewHTTPClient(HTTPConfig{
		URL:       server.URL,
		Model:     "test-model",
		TimeoutMS: 1,
	})
	if err != nil {
		t.Fatalf("NewHTTPClient returned error: %v", err)
	}

	// When
	_, err = client.Complete(context.Background(), Request{Prompt: "hello model"})

	// Then
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Complete error = %v, want deadline exceeded", err)
	}
}

func Test_HTTPClient_Complete_classifies_http_status_errors(t *testing.T) {
	tests := []struct {
		name string
		code int
		want error
	}{
		{name: "unauthorized", code: http.StatusUnauthorized, want: ErrHTTPUnauthorized},
		{name: "rate limited", code: http.StatusTooManyRequests, want: ErrHTTPRateLimited},
		{name: "unavailable", code: http.StatusBadGateway, want: ErrHTTPUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.code)
				if _, err := w.Write([]byte(`provider error`)); err != nil {
					t.Fatalf("write response: %v", err)
				}
			}))
			defer server.Close()
			client, err := NewHTTPClient(HTTPConfig{
				URL:   server.URL,
				Model: "test-model",
			})
			if err != nil {
				t.Fatalf("NewHTTPClient returned error: %v", err)
			}

			// When
			_, err = client.Complete(context.Background(), Request{Prompt: "hello model"})

			// Then
			if !errors.Is(err, tt.want) {
				t.Fatalf("Complete error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_HTTPClient_Complete_returns_typed_http_status_error(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
		if _, err := w.Write([]byte(`{"error":"slow down"}`)); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()
	client, err := NewHTTPClient(HTTPConfig{
		URL:   server.URL,
		Model: "test-model",
	})
	if err != nil {
		t.Fatalf("NewHTTPClient returned error: %v", err)
	}

	// When
	_, err = client.Complete(context.Background(), Request{Prompt: "hello model"})

	// Then
	if !errors.Is(err, ErrHTTPRateLimited) {
		t.Fatalf("Complete error = %v, want rate limited sentinel", err)
	}
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("Complete error = %v, want HTTPStatusError", err)
	}
	if statusErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("StatusCode = %d, want 429", statusErr.StatusCode)
	}
	if statusErr.Kind != HTTPErrorKindRateLimited {
		t.Fatalf("Kind = %q, want rate_limited", statusErr.Kind)
	}
	if statusErr.RetryAfterMS != 2000 {
		t.Fatalf("RetryAfterMS = %d, want 2000", statusErr.RetryAfterMS)
	}
	if !strings.Contains(statusErr.Body, "slow down") {
		t.Fatalf("Body = %q, want provider error body", statusErr.Body)
	}
}
