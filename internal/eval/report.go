package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type savedReport struct {
	TaskID         string                 `json:"task_id"`
	ChangedFiles   []string               `json:"changed_files"`
	CheckResults   []commandResult        `json:"check_results"`
	PatchResults   []patchResult          `json:"patch_results"`
	PatchPreviews  []patchResult          `json:"patch_previews"`
	EvidencePaths  []string               `json:"evidence_paths"`
	WorktreeStatus worktreeStatusEvidence `json:"worktree_status"`
	raw            reportFields
}

type commandResult struct {
	Argv     []string `json:"argv"`
	Status   string   `json:"status"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
}

type patchResult struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

type reportFields map[string]json.RawMessage

func loadReport(path string) (savedReport, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return savedReport{}, fmt.Errorf("read report: %w", err)
	}
	var report savedReport
	if err := json.Unmarshal(content, &report); err != nil {
		return savedReport{}, fmt.Errorf("%w: decode report JSON: %w", ErrInvalidReport, err)
	}
	if err := json.Unmarshal(content, &report.raw); err != nil {
		return savedReport{}, fmt.Errorf("%w: decode report fields: %w", ErrInvalidReport, err)
	}
	report.TaskID = strings.TrimSpace(report.TaskID)
	if report.TaskID == "" {
		return savedReport{}, fmt.Errorf("%w: task_id is required", ErrInvalidReport)
	}
	return report, nil
}

func (r savedReport) hasReportField(path string) bool {
	parts := strings.Split(strings.TrimSpace(path), ".")
	if len(parts) == 0 {
		return false
	}
	fields := r.raw
	for index, part := range parts {
		raw, ok := fields[part]
		if !ok || len(raw) == 0 || string(raw) == "null" {
			return false
		}
		if index == len(parts)-1 {
			return !isEmptyJSONValue(raw)
		}
		var nested reportFields
		if err := json.Unmarshal(raw, &nested); err != nil {
			return false
		}
		fields = nested
	}
	return false
}

func isEmptyJSONValue(raw json.RawMessage) bool {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text) == ""
	}
	return string(raw) == "[]" || string(raw) == "{}"
}
