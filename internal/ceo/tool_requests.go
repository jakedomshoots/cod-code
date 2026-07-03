package ceo

import (
	"context"
	"strings"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type toolRequestState struct {
	Request      JobRequest
	Space        workspace.Workspace
	HasWorkspace bool
	PriorResults []subagent.Result
}

func (r Runtime) runSubagentToolRequests(ctx context.Context, results []subagent.Result, state toolRequestState) ([]subagent.Result, error) {
	next := append([]subagent.Result(nil), results...)
	for index, result := range next {
		toolResults := make([]subagent.ToolResult, 0, len(result.ToolRequests))
		for requestIndex, request := range result.ToolRequests {
			if state.Request.MaxToolRequests > 0 && requestIndex >= state.Request.MaxToolRequests {
				toolResults = append(toolResults, toolLimitResult(request))
				continue
			}
			toolResults = append(toolResults, r.runSubagentToolRequest(ctx, result, request, state))
		}
		next[index].ToolResults = toolResults
	}
	return next, nil
}

func toolLimitResult(request subagent.ToolRequest) subagent.ToolResult {
	return subagent.ToolResult{
		Action: strings.TrimSpace(request.Action),
		Path:   strings.TrimSpace(request.Path),
		Query:  strings.TrimSpace(request.Query),
		Status: "skipped",
		Error:  "tool request limit reached",
	}
}

func (r Runtime) runSubagentToolRequest(ctx context.Context, result subagent.Result, request subagent.ToolRequest, state toolRequestState) subagent.ToolResult {
	action := strings.TrimSpace(request.Action)
	toolResult := subagent.ToolResult{
		Action: action,
		Path:   strings.TrimSpace(request.Path),
		Query:  strings.TrimSpace(request.Query),
	}
	if action == "" || !jobpacket.IsKnownAction(jobpacket.Action(action)) {
		toolResult.Status = "invalid"
		toolResult.Error = "unknown action"
		return toolResult
	}
	if !subagentCanRunAction(result, action) {
		toolResult.Status = "denied"
		toolResult.Error = "action is not allowed for subagent"
		return toolResult
	}
	switch jobpacket.Action(action) {
	case jobpacket.ActionReadWorkspace:
		return runReadWorkspaceTool(ctx, state, request, toolResult)
	case jobpacket.ActionSearchWorkspace:
		return runSearchWorkspaceTool(ctx, state, request, toolResult)
	case jobpacket.ActionNetworkResearch:
		return runNetworkResearchTool(ctx, state, request, toolResult)
	case jobpacket.ActionRunChecks:
		return r.runChecksTool(ctx, state, toolResult)
	case jobpacket.ActionVerifyEvidence:
		return runVerifyEvidenceTool(state, toolResult)
	default:
		toolResult.Status = "skipped"
		toolResult.Error = "action is not executable by runtime"
		return toolResult
	}
}

func runReadWorkspaceTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	if !state.HasWorkspace {
		result.Status = "skipped"
		result.Error = "workspace is required"
		return result
	}
	readResult, err := state.Space.ReadText(ctx, workspace.ReadTextRequest{
		Path:     request.Path,
		MaxBytes: request.MaxBytes,
	})
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "pass"
	result.Path = readResult.Path
	result.Output = readResult.Content
	result.Bytes = readResult.Bytes
	result.Truncated = readResult.Truncated
	return result
}

func runSearchWorkspaceTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	if !state.HasWorkspace {
		result.Status = "skipped"
		result.Error = "workspace is required"
		return result
	}
	searchResult, err := state.Space.SearchText(ctx, workspace.SearchTextRequest{
		Query:      request.Query,
		MaxMatches: request.MaxMatches,
	})
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "pass"
	result.Query = searchResult.Query
	result.MatchCount = len(searchResult.Matches)
	result.Matches = workspaceMatches(searchResult.Matches)
	return result
}

func (r Runtime) runChecksTool(ctx context.Context, state toolRequestState, result subagent.ToolResult) subagent.ToolResult {
	if len(checkCommands(state.Request)) == 0 {
		result.Status = "skipped"
		result.Error = "check command is required"
		return result
	}
	checks, err := r.runChecks(ctx, state.Request)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	if len(checks) == 0 {
		result.Status = "skipped"
		result.Error = "check command is required"
		return result
	}
	last := checks[len(checks)-1]
	result.Status = last.Status
	result.Output = last.Stdout
	result.Error = last.Stderr
	result.ExitCode = last.ExitCode
	return result
}

func workspaceMatches(matches []workspace.SearchTextMatch) []subagent.ToolMatch {
	converted := make([]subagent.ToolMatch, 0, len(matches))
	for _, match := range matches {
		converted = append(converted, subagent.ToolMatch{
			Path: match.Path,
			Line: match.Line,
			Text: match.Text,
		})
	}
	return converted
}

func subagentCanRunAction(result subagent.Result, action string) bool {
	for _, allowedAction := range result.AllowedActions {
		if allowedAction == action {
			return true
		}
	}
	return false
}
