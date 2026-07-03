package subagent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ceoharness/internal/model"
	"ceoharness/internal/prompt"
)

var ErrTaskRequired = errors.New("subagent task is required")

type TaskPacket struct {
	Task            string       `json:"task"`
	AgentName       string       `json:"agent_name"`
	Role            string       `json:"role"`
	Assignment      string       `json:"assignment,omitempty"`
	ProviderName    string       `json:"provider,omitempty"`
	ContextMode     string       `json:"context_mode"`
	AllowedActions  []string     `json:"allowed_actions,omitempty"`
	WorkspaceBrief  string       `json:"workspace_brief,omitempty"`
	PriorFindings   string       `json:"prior_findings,omitempty"`
	ToolResults     []ToolResult `json:"tool_results,omitempty"`
	MaxContextBytes int          `json:"max_context_bytes"`
}

type AttemptRecord struct {
	Attempt              int    `json:"attempt"`
	Status               string `json:"status"`
	Error                string `json:"error,omitempty"`
	ProviderErrorKind    string `json:"provider_error_kind,omitempty"`
	ProviderHTTPStatus   int    `json:"provider_http_status,omitempty"`
	ProviderRetryAfterMS int64  `json:"provider_retry_after_ms,omitempty"`
}

type Result struct {
	AgentName                     string          `json:"agent_name"`
	Role                          string          `json:"role"`
	Assignment                    string          `json:"assignment,omitempty"`
	Stage                         int             `json:"stage,omitempty"`
	Status                        string          `json:"status"`
	Attempts                      int             `json:"attempts"`
	DurationMS                    int64           `json:"duration_ms"`
	ModelSource                   string          `json:"model_source"`
	ProviderName                  string          `json:"provider_name,omitempty"`
	ProviderRequestID             string          `json:"provider_request_id,omitempty"`
	ProviderModel                 string          `json:"provider_model,omitempty"`
	ProviderFallbackFrom          string          `json:"provider_fallback_from,omitempty"`
	ProviderFallbackReason        string          `json:"provider_fallback_reason,omitempty"`
	ProviderPromptTokens          int             `json:"provider_prompt_tokens,omitempty"`
	ProviderCompletionTokens      int             `json:"provider_completion_tokens,omitempty"`
	ProviderTotalTokens           int             `json:"provider_total_tokens,omitempty"`
	ProviderEstimatedCostMicroUSD int64           `json:"provider_estimated_cost_microusd,omitempty"`
	ProviderErrorKind             string          `json:"provider_error_kind,omitempty"`
	ProviderHTTPStatus            int             `json:"provider_http_status,omitempty"`
	ProviderRetryAfterMS          int64           `json:"provider_retry_after_ms,omitempty"`
	ContextReceived               string          `json:"context_received"`
	AllowedActions                []string        `json:"allowed_actions,omitempty"`
	ContextBytes                  int             `json:"context_bytes"`
	ContextTruncated              bool            `json:"context_truncated"`
	ContextTruncatedFields        []string        `json:"context_truncated_fields,omitempty"`
	OutputTruncated               bool            `json:"output_truncated,omitempty"`
	Reused                        bool            `json:"reused,omitempty"`
	NoProgressStopped             bool            `json:"no_progress_stopped,omitempty"`
	PromptBytes                   int             `json:"prompt_bytes"`
	PriorFindings                 string          `json:"prior_findings,omitempty"`
	Confidence                    *float64        `json:"confidence,omitempty"`
	Summary                       string          `json:"summary"`
	Questions                     []string        `json:"questions,omitempty"`
	PatchProposals                []PatchProposal `json:"patches,omitempty"`
	AttemptErrors                 []string        `json:"attempt_errors,omitempty"`
	AttemptRecords                []AttemptRecord `json:"attempt_records,omitempty"`
	ToolRequests                  []ToolRequest   `json:"tool_requests,omitempty"`
	ToolResults                   []ToolResult    `json:"tool_results,omitempty"`
	InitialSummary                string          `json:"initial_summary,omitempty"`
	ToolFeedbackPasses            int             `json:"tool_feedback_passes,omitempty"`
	Evidence                      []string        `json:"evidence"`
}

type Runner struct {
	client model.Client
}

func NewRunner() Runner {
	return NewRunnerWithModel(model.NewStaticClient())
}

func NewRunnerWithModel(client model.Client) Runner {
	return Runner{client: client}
}

func (r Runner) Run(ctx context.Context, packet TaskPacket) (Result, error) {
	started := time.Now()
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	cleanTask := strings.TrimSpace(packet.Task)
	if cleanTask == "" {
		return Result{}, ErrTaskRequired
	}
	workspaceBrief := strings.TrimSpace(packet.WorkspaceBrief)
	priorFindings := strings.TrimSpace(packet.PriorFindings)
	assignment := strings.TrimSpace(packet.Assignment)
	builtPrompt, err := prompt.Build(ctx, prompt.Request{
		Task:        cleanTask,
		AgentName:   packet.AgentName,
		Role:        packet.Role,
		Assignment:  assignment,
		ContextMode: packet.ContextMode,
		AllowedActions: append(
			[]string(nil),
			packet.AllowedActions...,
		),
		WorkspaceBrief: workspaceBrief,
		PriorFindings:  priorFindings,
		ToolResults:    RenderToolResults(packet.ToolResults),
		MaxBytes:       packet.MaxContextBytes,
	})
	if err != nil {
		return Result{}, fmt.Errorf("build prompt: %w", err)
	}
	client := r.client
	if client == nil {
		client = model.NewStaticClient()
	}
	response, err := client.Complete(ctx, model.Request{
		Prompt: builtPrompt.Text,
		Metadata: model.RequestMetadata{
			Kind:        "subagent",
			AgentName:   packet.AgentName,
			AgentRole:   packet.Role,
			ContextMode: packet.ContextMode,
		},
	})
	if err != nil {
		return Result{}, fmt.Errorf("complete prompt: %w", err)
	}
	if strings.TrimSpace(response.Text) == "" {
		return Result{}, newEmptyOutputError()
	}
	modelOutput, err := parseModelOutput(response.Text, response.RequireStructuredOutput)
	if err != nil {
		return Result{}, newInvalidOutputError(err)
	}
	promptBytes := response.PromptBytes
	if promptBytes == 0 {
		promptBytes = builtPrompt.Bytes
	}

	return Result{
		AgentName:                     packet.AgentName,
		Role:                          packet.Role,
		Assignment:                    builtPrompt.Assignment,
		Status:                        modelOutput.Status,
		Attempts:                      1,
		DurationMS:                    time.Since(started).Milliseconds(),
		ModelSource:                   "local",
		ProviderRequestID:             response.RequestID,
		ProviderModel:                 response.Model,
		ProviderPromptTokens:          response.PromptTokens,
		ProviderCompletionTokens:      response.CompletionTokens,
		ProviderTotalTokens:           response.TotalTokens,
		ProviderEstimatedCostMicroUSD: response.CostMicroUSD,
		ContextReceived:               packet.ContextMode,
		AllowedActions:                append([]string(nil), packet.AllowedActions...),
		ContextBytes:                  builtPrompt.ContextBytes,
		ContextTruncated:              builtPrompt.Truncated,
		ContextTruncatedFields:        append([]string(nil), builtPrompt.TruncatedFields...),
		PromptBytes:                   promptBytes,
		PriorFindings:                 builtPrompt.PriorFindings,
		Confidence:                    modelOutput.Confidence,
		Summary:                       modelOutput.Summary,
		Questions:                     modelOutput.Questions,
		PatchProposals:                modelOutput.PatchProposals,
		ToolRequests:                  modelOutput.ToolRequests,
		Evidence:                      evidenceForModelOutput(modelOutput),
	}, nil
}

func evidenceForModelOutput(output ModelOutput) []string {
	if len(output.Evidence) > 0 {
		return append([]string(nil), output.Evidence...)
	}
	return []string{
		"task packet parsed",
		"lean context received",
	}
}
