package subagent

import (
	"bytes"
	"encoding/json"
	"strings"
)

type ToolRequest struct {
	Action     string `json:"action"`
	Path       string `json:"path,omitempty"`
	Query      string `json:"query,omitempty"`
	URL        string `json:"url,omitempty"`
	App        string `json:"app,omitempty"`
	Tool       string `json:"tool,omitempty"`
	MaxBytes   int    `json:"max_bytes,omitempty"`
	MaxMatches int    `json:"max_matches,omitempty"`
}

func (r *ToolRequest) UnmarshalJSON(data []byte) error {
	var path string
	if err := json.Unmarshal(data, &path); err == nil {
		*r = ToolRequest{Path: path}
		return nil
	}
	type toolRequestAlias ToolRequest
	var decoded toolRequestAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = ToolRequest(decoded)
	return nil
}

type ToolMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

type ToolResult struct {
	Action        string      `json:"action"`
	Status        string      `json:"status"`
	Path          string      `json:"path,omitempty"`
	Query         string      `json:"query,omitempty"`
	URL           string      `json:"url,omitempty"`
	App           string      `json:"app,omitempty"`
	Tool          string      `json:"tool,omitempty"`
	Permission    string      `json:"permission,omitempty"`
	ReceiptSHA256 string      `json:"receipt_sha256,omitempty"`
	Output        string      `json:"output,omitempty"`
	Error         string      `json:"error,omitempty"`
	Bytes         int         `json:"bytes,omitempty"`
	Truncated     bool        `json:"truncated,omitempty"`
	MatchCount    int         `json:"match_count,omitempty"`
	Matches       []ToolMatch `json:"matches,omitempty"`
	ExitCode      int         `json:"exit_code,omitempty"`
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
	Evidence     stringList      `json:"evidence"`
	Questions    stringList      `json:"questions"`
	ToolRequests []ToolRequest   `json:"tool_requests"`
	Patches      []PatchProposal `json:"patches"`
}

type stringList []string

func (l *stringList) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*l = stringList{one}
		return nil
	}

	var many []json.RawMessage
	if err := json.Unmarshal(data, &many); err == nil {
		values := make([]string, 0, len(many))
		for _, item := range many {
			values = append(values, rawJSONText(item))
		}
		*l = stringList(values)
		return nil
	}

	*l = stringList{rawJSONText(data)}
	return nil
}

func rawJSONText(data []byte) string {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		return text
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err == nil {
		return compact.String()
	}
	return strings.TrimSpace(string(data))
}
