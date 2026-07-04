package ceo

import (
	"context"
	"fmt"
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
		requests := result.ToolRequests
		if len(requests) == 0 {
			requests = inferredReadWorkspaceRequests(ctx, result, state)
			next[index].ToolRequests = requests
		}
		toolResults := make([]subagent.ToolResult, 0, len(requests))
		for requestIndex, request := range requests {
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

func inferredReadWorkspaceRequests(ctx context.Context, result subagent.Result, state toolRequestState) []subagent.ToolRequest {
	if result.Status != "needs_input" || !state.HasWorkspace || !subagentCanRunAction(result, string(jobpacket.ActionReadWorkspace)) {
		return nil
	}
	text := result.Summary + "\n" + strings.Join(result.Questions, "\n")
	if strings.TrimSpace(text) == "" {
		return nil
	}
	brief, err := state.Space.Brief(ctx, workspace.BriefRequest{MaxFiles: workspace.DefaultBriefMaxFiles})
	if err != nil {
		return nil
	}
	requests := []subagent.ToolRequest{}
	for _, file := range brief.Files {
		if strings.Contains(text, file.Path) {
			requests = append(requests, subagent.ToolRequest{
				Action: string(jobpacket.ActionReadWorkspace),
				Path:   file.Path,
			})
			if len(requests) >= 4 {
				break
			}
		}
	}
	return requests
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
	if isGenericWorkspaceReadPath(request.Path) {
		return runWorkspaceBriefTool(ctx, state, request, result)
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

func isGenericWorkspaceReadPath(path string) bool {
	switch strings.ToLower(strings.TrimSpace(path)) {
	case "", ".", "/", "workspace", "read_workspace":
		return true
	default:
		return false
	}
}

func runWorkspaceBriefTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	brief, err := state.Space.Brief(ctx, workspace.BriefRequest{MaxFiles: workspace.DefaultBriefMaxFiles})
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "workspace files: %d\n", brief.FileCount)
	for _, file := range brief.Files {
		fmt.Fprintf(&builder, "- %s (%d bytes)\n", file.Path, file.Bytes)
	}
	if brief.Truncated {
		builder.WriteString("truncated: true\n")
	}
	writeSmallWorkspaceContents(ctx, state, request, brief, &builder)
	result.Status = "pass"
	result.Path = "workspace"
	result.Output = strings.TrimSpace(builder.String())
	result.Bytes = len(result.Output)
	result.Truncated = brief.Truncated
	return result
}

func writeSmallWorkspaceContents(ctx context.Context, state toolRequestState, request subagent.ToolRequest, brief workspace.Brief, builder *strings.Builder) {
	maxBytes := request.MaxBytes
	if maxBytes < 1 {
		maxBytes = 4096
	}
	if len(brief.Files) > 10 || maxBytes <= 0 {
		return
	}
	remaining := maxBytes
	wroteHeader := false
	for _, file := range brief.Files {
		if file.Bytes > int64(maxBytes) || remaining <= 0 {
			continue
		}
		readResult, err := state.Space.ReadText(ctx, workspace.ReadTextRequest{
			Path:     file.Path,
			MaxBytes: remaining,
		})
		if err != nil || strings.TrimSpace(readResult.Content) == "" {
			continue
		}
		if !wroteHeader {
			builder.WriteString("\nfile_contents:\n")
			wroteHeader = true
		}
		fmt.Fprintf(builder, "--- %s\n%s\n", readResult.Path, readResult.Content)
		remaining -= readResult.Bytes
		if readResult.Truncated {
			builder.WriteString("[truncated]\n")
			return
		}
	}
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
