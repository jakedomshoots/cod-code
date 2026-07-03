package subagent

import (
	"encoding/json"
	"fmt"
	"strings"

	"ceoharness/internal/model"
)

type ToolRequest struct {
	Action     string `json:"action"`
	Path       string `json:"path,omitempty"`
	Query      string `json:"query,omitempty"`
	MaxBytes   int    `json:"max_bytes,omitempty"`
	MaxMatches int    `json:"max_matches,omitempty"`
}

type ToolMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

type ToolResult struct {
	Action     string      `json:"action"`
	Status     string      `json:"status"`
	Path       string      `json:"path,omitempty"`
	Query      string      `json:"query,omitempty"`
	Output     string      `json:"output,omitempty"`
	Error      string      `json:"error,omitempty"`
	Bytes      int         `json:"bytes,omitempty"`
	Truncated  bool        `json:"truncated,omitempty"`
	MatchCount int         `json:"match_count,omitempty"`
	Matches    []ToolMatch `json:"matches,omitempty"`
	ExitCode   int         `json:"exit_code,omitempty"`
}

type PatchProposal struct {
	Path    string `json:"path"`
	Old     string `json:"old,omitempty"`
	New     string `json:"new,omitempty"`
	Content string `json:"content,omitempty"`
}

type ModelOutput struct {
	Status         string
	Summary        string
	Confidence     *float64
	Evidence       []string
	Questions      []string
	ToolRequests   []ToolRequest
	PatchProposals []PatchProposal
	Structured     bool
}

type modelOutputEnvelope struct {
	Status       string          `json:"status"`
	Summary      string          `json:"summary"`
	Confidence   *float64        `json:"confidence,omitempty"`
	Evidence     []string        `json:"evidence"`
	Questions    []string        `json:"questions"`
	ToolRequests []ToolRequest   `json:"tool_requests"`
	Patches      []PatchProposal `json:"patches"`
}

type toolResultEnvelope struct {
	ToolResults []ToolResult `json:"tool_results"`
}

const maxRenderedToolTextBytes = 1200

func ParseToolRequests(text string) ([]ToolRequest, error) {
	output, err := ParseModelOutput(text)
	if err != nil {
		return nil, err
	}
	return output.ToolRequests, nil
}

func ParseModelOutput(text string) (ModelOutput, error) {
	return parseModelOutput(text, false)
}

func parseModelOutput(text string, requireStructured bool) (ModelOutput, error) {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return ModelOutput{Status: "pass"}, nil
	}
	payload, ok := model.JSONPayload(clean)
	if !ok || !hasModelOutputEnvelope(payload) {
		if requireStructured {
			return ModelOutput{}, fmt.Errorf("structured model output required")
		}
		return ModelOutput{Status: "pass", Summary: clean}, nil
	}
	var envelope modelOutputEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return ModelOutput{}, fmt.Errorf("parse model output: %w", err)
	}
	if err := validateToolRequests(envelope.ToolRequests); err != nil {
		return ModelOutput{}, err
	}
	if envelope.Confidence != nil && (*envelope.Confidence < 0 || *envelope.Confidence > 1) {
		return ModelOutput{}, fmt.Errorf("confidence must be between 0 and 1")
	}
	summary := strings.TrimSpace(envelope.Summary)
	if summary == "" {
		summary = payload
	}
	return ModelOutput{
		Status:         modelOutputStatus(envelope.Status),
		Summary:        summary,
		Confidence:     envelope.Confidence,
		Evidence:       cleanEvidence(envelope.Evidence),
		Questions:      cleanQuestions(envelope.Questions),
		ToolRequests:   append([]ToolRequest(nil), envelope.ToolRequests...),
		PatchProposals: append([]PatchProposal(nil), envelope.Patches...),
		Structured:     true,
	}, nil
}

func RenderToolResults(results []ToolResult) string {
	if len(results) == 0 {
		return ""
	}
	body, err := json.Marshal(toolResultEnvelope{ToolResults: compactToolResultsForRender(results)})
	if err != nil {
		return ""
	}
	return string(body)
}

func compactToolResultsForRender(results []ToolResult) []ToolResult {
	compacted := make([]ToolResult, 0, len(results))
	for _, result := range results {
		output, outputTruncated := compactRenderedToolText(result.Output)
		errorText, errorTruncated := compactRenderedToolText(result.Error)
		result.Output = output
		result.Error = errorText
		result.Truncated = result.Truncated || outputTruncated || errorTruncated
		compacted = append(compacted, result)
	}
	return compacted
}

func compactRenderedToolText(text string) (string, bool) {
	if len(text) <= maxRenderedToolTextBytes {
		return text, false
	}
	return truncateRenderedToolText(text, maxRenderedToolTextBytes) + "\n[truncated]", true
}

func truncateRenderedToolText(text string, maxBytes int) string {
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

func hasModelOutputEnvelope(text string) bool {
	if !strings.HasPrefix(text, "{") {
		return false
	}
	for _, field := range []string{`"status"`, `"summary"`, `"confidence"`, `"evidence"`, `"questions"`, `"tool_requests"`, `"patches"`} {
		if strings.Contains(text, field) {
			return true
		}
	}
	return false
}

func validateToolRequests(requests []ToolRequest) error {
	for index, request := range requests {
		if strings.TrimSpace(request.Action) == "" {
			return fmt.Errorf("tool_requests[%d] action is required", index)
		}
	}
	return nil
}

func cleanEvidence(evidence []string) []string {
	cleaned := make([]string, 0, len(evidence))
	for _, item := range evidence {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		cleaned = append(cleaned, clean)
	}
	return cleaned
}

func cleanQuestions(questions []string) []string {
	cleaned := make([]string, 0, len(questions))
	for _, item := range questions {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		cleaned = append(cleaned, clean)
	}
	return cleaned
}

func modelOutputStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "needs_input":
		return "needs_input"
	case "fail":
		return "fail"
	default:
		return "pass"
	}
}
