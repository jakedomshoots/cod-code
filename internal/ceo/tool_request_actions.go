package ceo

import (
	"context"
	"encoding/json"
	"strings"

	"ceoharness/internal/browseruse"
	"ceoharness/internal/computeruse"
	"ceoharness/internal/subagent"
	"ceoharness/internal/toolmanifest"
	"ceoharness/internal/workspace"
)

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

func runBrowserReadTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	targetURL := strings.TrimSpace(request.URL)
	if targetURL == "" {
		targetURL = strings.TrimSpace(request.Query)
	}
	browserResult := browseruse.Read(ctx, browseruse.Request{
		URL:       targetURL,
		Policy:    state.Request.BrowserPolicy,
		MaxBytes:  request.MaxBytes,
		TimeoutMS: state.Request.ToolCommandTimeoutMS,
	})
	result.Status = browserResult.Status
	result.URL = browserResult.URL
	result.Path = browserResult.URL
	result.Permission = browserResult.Permission
	result.Output = browserResult.Output
	result.Error = browserResult.Error
	result.Bytes = browserResult.Bytes
	result.Truncated = browserResult.Truncated
	result.ReceiptSHA256 = browserResult.ReceiptSHA256
	return result
}

func runComputerSnapshotTool(ctx context.Context, state toolRequestState, request subagent.ToolRequest, result subagent.ToolResult) subagent.ToolResult {
	computerResult := computeruse.Snapshot(ctx, computeruse.Request{
		App:       strings.TrimSpace(request.App),
		Command:   state.Request.ComputerCommand,
		Policy:    state.Request.ComputerPolicy,
		MaxBytes:  request.MaxBytes,
		TimeoutMS: state.Request.ToolCommandTimeoutMS,
	})
	result.Status = computerResult.Status
	result.App = computerResult.App
	result.Path = computerResult.App
	result.Permission = computerResult.Permission
	result.Output = computerResult.Output
	result.Error = computerResult.Error
	result.Bytes = computerResult.Bytes
	result.Truncated = computerResult.Truncated
	result.ExitCode = computerResult.ExitCode
	result.ReceiptSHA256 = computerResult.ReceiptSHA256
	return result
}

func runToolManifestTool(result subagent.ToolResult) subagent.ToolResult {
	body, err := json.Marshal(toolmanifest.Default())
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}
	result.Status = "pass"
	result.Tool = "tools.manifest"
	result.Output = string(body)
	result.Bytes = len(body)
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
