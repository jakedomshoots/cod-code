package ceo

import (
	"context"
	"errors"
	"strings"

	"ceoharness/internal/researchrunner"
	"ceoharness/internal/subagent"
)

func runNetworkResearchTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	if len(state.Request.ResearchCommand) == 0 {
		result.Status = "skipped"
		result.Error = "research command is required"
		return result
	}
	query := strings.TrimSpace(request.Query)
	if query == "" {
		result.Status = "invalid"
		result.Error = "query is required"
		return result
	}
	researchResult, err := researchrunner.NewRunner().Run(ctx, researchrunner.Command{
		Argv:      state.Request.ResearchCommand,
		Query:     query,
		MaxBytes:  request.MaxBytes,
		TimeoutMS: state.Request.ToolCommandTimeoutMS,
	})
	if errors.Is(err, researchrunner.ErrCommandRequired) || errors.Is(err, researchrunner.ErrQueryRequired) {
		result.Status = "invalid"
		result.Error = err.Error()
		return result
	}
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = researchResult.Status
	result.Query = query
	result.Output = researchResult.Output
	result.Error = researchResult.Error
	result.ExitCode = researchResult.ExitCode
	result.Bytes = researchResult.Bytes
	result.Truncated = researchResult.Truncated
	return result
}
