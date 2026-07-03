package ceo

import (
	"strings"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

const traceSummaryLimit = 140

type ContextTraceEntry struct {
	AgentName         string               `json:"agent_name"`
	Role              string               `json:"role,omitempty"`
	Stage             int                  `json:"stage,omitempty"`
	ProviderName      string               `json:"provider_name,omitempty"`
	TaskSummary       string               `json:"task_summary"`
	AssignmentSummary string               `json:"assignment_summary,omitempty"`
	BudgetUnit        string               `json:"budget_unit"`
	MaxContextBytes   int                  `json:"max_context_bytes"`
	ContextBytes      int                  `json:"context_bytes"`
	PromptBytes       int                  `json:"prompt_bytes,omitempty"`
	ContextTruncated  bool                 `json:"context_truncated"`
	TruncatedFields   []string             `json:"truncated_fields,omitempty"`
	WorkspaceBrief    ContextTraceBrief    `json:"workspace_brief,omitempty"`
	PriorFindings     ContextTraceFindings `json:"prior_findings"`
	ExcludedContent   ContextTraceExcluded `json:"excluded_content"`
}

type ContextTraceBrief struct {
	FileCount      int  `json:"file_count,omitempty"`
	ShownFileCount int  `json:"shown_file_count,omitempty"`
	Bytes          int  `json:"bytes,omitempty"`
	Truncated      bool `json:"truncated,omitempty"`
}

type ContextTraceFindings struct {
	Count int `json:"count"`
	Bytes int `json:"bytes"`
}

type ContextTraceExcluded struct {
	WorkspaceExcludes      []string `json:"workspace_excludes,omitempty"`
	WorkspaceExcludedCount int      `json:"workspace_excluded_count,omitempty"`
	ContentKinds           []string `json:"content_kinds,omitempty"`
}

func buildContextTrace(packet jobpacket.Packet, brief *workspace.Brief, results []subagent.Result) []ContextTraceEntry {
	if len(results) == 0 {
		return nil
	}
	agents := agentsByName(packet.Subagents)
	entries := make([]ContextTraceEntry, 0, len(results))
	for _, result := range results {
		agent := agents[result.AgentName]
		entries = append(entries, ContextTraceEntry{
			AgentName:         result.AgentName,
			Role:              result.Role,
			Stage:             result.Stage,
			ProviderName:      result.ProviderName,
			TaskSummary:       redactTraceSecrets(trimTraceSummary(packet.Task)),
			AssignmentSummary: redactTraceSecrets(trimTraceSummary(result.Assignment)),
			BudgetUnit:        "bytes",
			MaxContextBytes:   contextBudgetForAgent(packet, agent),
			ContextBytes:      result.ContextBytes,
			PromptBytes:       result.PromptBytes,
			ContextTruncated:  result.ContextTruncated,
			TruncatedFields:   append([]string(nil), result.ContextTruncatedFields...),
			WorkspaceBrief:    contextTraceBrief(brief),
			PriorFindings:     contextTraceFindings(result.PriorFindings),
			ExcludedContent:   contextTraceExcluded(brief),
		})
	}
	return entries
}

func agentsByName(agents []jobpacket.Subagent) map[string]jobpacket.Subagent {
	out := make(map[string]jobpacket.Subagent, len(agents))
	for _, agent := range agents {
		out[agent.Name] = agent
	}
	return out
}

func contextTraceBrief(brief *workspace.Brief) ContextTraceBrief {
	if brief == nil {
		return ContextTraceBrief{}
	}
	return ContextTraceBrief{
		FileCount:      brief.FileCount,
		ShownFileCount: len(brief.Files),
		Bytes:          len(renderWorkspaceBrief(brief)),
		Truncated:      brief.Truncated,
	}
}

func contextTraceFindings(text string) ContextTraceFindings {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ContextTraceFindings{}
	}
	return ContextTraceFindings{
		Count: len(strings.Split(clean, "\n")),
		Bytes: len(clean),
	}
}

func contextTraceExcluded(brief *workspace.Brief) ContextTraceExcluded {
	excluded := ContextTraceExcluded{
		ContentKinds: []string{"raw_prompts", "repo_file_contents", "environment_values"},
	}
	if brief == nil {
		return excluded
	}
	excluded.WorkspaceExcludes = append([]string(nil), brief.ExcludePaths...)
	excluded.WorkspaceExcludedCount = brief.ExcludedPathCount
	return excluded
}

func trimTraceSummary(text string) string {
	clean := strings.Join(strings.Fields(text), " ")
	if len(clean) <= traceSummaryLimit {
		return clean
	}
	end := 0
	for index := range clean {
		if index > traceSummaryLimit {
			break
		}
		end = index
	}
	if end == 0 {
		return ""
	}
	return clean[:end] + "..."
}

func redactTraceSecrets(text string) string {
	words := strings.Fields(text)
	for index, word := range words {
		if traceWordHasSecret(word) {
			words[index] = traceRedactedWord(word)
		}
	}
	return strings.Join(words, " ")
}

func traceWordHasSecret(word string) bool {
	lower := strings.ToLower(word)
	return strings.Contains(lower, "api_key=") ||
		strings.Contains(lower, "token=") ||
		strings.Contains(lower, "secret=") ||
		strings.Contains(lower, "password=") ||
		strings.Contains(word, "sk-")
}

func traceRedactedWord(word string) string {
	if before, _, found := strings.Cut(word, "="); found && !traceSecretKeyName(before) {
		return before + "=[redacted_secret]"
	}
	return "[redacted_secret]"
}

func traceSecretKeyName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "key") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "password")
}
