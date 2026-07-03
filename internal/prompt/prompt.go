package prompt

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrTaskRequired = errors.New("prompt task is required")

const responseContract = `response_contract: return one JSON object only: {"status":"pass|fail|needs_input","summary":"short result","confidence":0.0,"evidence":[],"questions":[],"tool_requests":[],"patches":[]}`

type Request struct {
	Task           string
	AgentName      string
	Role           string
	Assignment     string
	ContextMode    string
	AllowedActions []string
	WorkspaceBrief string
	PriorFindings  string
	ToolResults    string
	MaxBytes       int
}

type Prompt struct {
	Text            string
	Bytes           int
	ContextBytes    int
	Truncated       bool
	TruncatedFields []string
	Task            string
	Assignment      string
	WorkspaceBrief  string
	PriorFindings   string
	ToolResults     string
}

func Build(ctx context.Context, req Request) (Prompt, error) {
	if err := ctx.Err(); err != nil {
		return Prompt{}, err
	}
	task := strings.TrimSpace(req.Task)
	if task == "" {
		return Prompt{}, ErrTaskRequired
	}
	assignmentValue := strings.TrimSpace(req.Assignment)
	workspaceBriefValue := strings.TrimSpace(req.WorkspaceBrief)
	priorFindingsValue := strings.TrimSpace(req.PriorFindings)
	toolResultsValue := strings.TrimSpace(req.ToolResults)

	truncated := false
	truncatedFields := []string{}
	if req.MaxBytes > 0 {
		remaining := req.MaxBytes
		task, truncatedFields = takeBudgetedContext("task", task, &remaining, truncatedFields)
		assignmentValue, truncatedFields = takeBudgetedContext("assignment", assignmentValue, &remaining, truncatedFields)
		toolResultsValue, truncatedFields = takeBudgetedContext("tool_results", toolResultsValue, &remaining, truncatedFields)
		priorFindingsValue, truncatedFields = takeBudgetedContext("prior_findings", priorFindingsValue, &remaining, truncatedFields)
		workspaceBriefValue, truncatedFields = takeBudgetedContext("workspace_brief", workspaceBriefValue, &remaining, truncatedFields)
		truncated = len(truncatedFields) > 0
	}

	workspaceBrief := ""
	if workspaceBriefValue != "" {
		workspaceBrief = "\nworkspace_brief: " + workspaceBriefValue
	}
	priorFindings := ""
	if priorFindingsValue != "" {
		priorFindings = "\nprior_findings: " + priorFindingsValue
	}
	assignment := ""
	if assignmentValue != "" {
		assignment = "\nassignment: " + assignmentValue
	}
	toolResults := ""
	if toolResultsValue != "" {
		toolResults = "\ntool_results: " + toolResultsValue
	}
	text := fmt.Sprintf("agent: %s\nrole: %s%s\nmode: %s\nallowed_actions: %s\ntask:\n%s\n%s%s%s%s",
		strings.TrimSpace(req.AgentName),
		strings.TrimSpace(req.Role),
		assignment,
		strings.TrimSpace(req.ContextMode),
		strings.Join(req.AllowedActions, ", "),
		task,
		responseContract,
		workspaceBrief,
		priorFindings,
		toolResults,
	)
	return Prompt{
		Text:            text,
		Bytes:           len(text),
		ContextBytes:    len(task) + len(assignmentValue) + len(workspaceBriefValue) + len(priorFindingsValue) + len(toolResultsValue),
		Truncated:       truncated,
		TruncatedFields: append([]string(nil), truncatedFields...),
		Task:            task,
		Assignment:      assignmentValue,
		WorkspaceBrief:  workspaceBriefValue,
		PriorFindings:   priorFindingsValue,
		ToolResults:     toolResultsValue,
	}, nil
}

func takeBudgetedContext(field string, text string, remaining *int, truncatedFields []string) (string, []string) {
	if text == "" {
		return "", truncatedFields
	}
	if *remaining <= 0 {
		return "", append(truncatedFields, field)
	}
	if len(text) <= *remaining {
		*remaining -= len(text)
		return text, truncatedFields
	}
	truncated := truncateBytes(text, *remaining)
	*remaining = 0
	return truncated, append(truncatedFields, field)
}

func truncateBytes(text string, maxBytes int) string {
	end := 0
	for index := range text {
		if index > maxBytes {
			break
		}
		end = index
	}
	if end == 0 && len(text) > 0 {
		return ""
	}
	return text[:end]
}
