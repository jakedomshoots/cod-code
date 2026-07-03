package ceo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

var ErrInvalidCEOReview = errors.New("invalid CEO review")

type ceoReviewInput struct {
	Packet        jobpacket.Packet
	Results       []subagent.Result
	Checks        []checkrunner.Result
	ChangedFiles  []string
	PatchResults  []workspace.ReplaceTextResult
	PatchPreviews []workspace.ReplaceTextResult
	GuardVerdict  string
}

type ceoReviewPayload struct {
	RecommendedVerdict string `json:"recommended_verdict"`
	Summary            string `json:"summary"`
}

func (r Runtime) runCEOReview(ctx context.Context, input ceoReviewInput) (*CEOReview, error) {
	if r.ceoReviewer == nil {
		return nil, nil
	}
	prompt := renderCEOReviewPrompt(input)
	response, err := r.ceoReviewer.Complete(ctx, model.Request{
		Prompt: prompt,
		Metadata: model.RequestMetadata{
			Kind:      "ceo_review",
			AgentName: "ceo",
			AgentRole: "final verdict",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("complete CEO review: %w", err)
	}
	review, err := parseCEOReviewResponse(response)
	if err != nil {
		return nil, err
	}
	if review.PromptBytes == 0 {
		review.PromptBytes = len(prompt)
	}
	if r.ceoReviewerRoute.Source != "" {
		review.ModelSource = r.ceoReviewerRoute.Source
	}
	if r.ceoReviewerRoute.ProviderName != "" {
		review.ProviderName = r.ceoReviewerRoute.ProviderName
	}
	return review, nil
}

func renderCEOReviewPrompt(input ceoReviewInput) string {
	var builder strings.Builder
	builder.WriteString("You are the CEO agent. Return JSON only: ")
	builder.WriteString(`{"recommended_verdict":"pass|fail","summary":"short reason"}`)
	builder.WriteString("\n")
	builder.WriteString("task: ")
	builder.WriteString(input.Packet.Task)
	builder.WriteString("\ntask_profile: kind=")
	builder.WriteString(input.Packet.TaskProfile.Kind)
	builder.WriteString(" risk=")
	builder.WriteString(input.Packet.TaskProfile.RiskLevel)
	builder.WriteString("\n")
	builder.WriteString("guard_verdict: ")
	builder.WriteString(input.GuardVerdict)
	builder.WriteString("\nsubagents:\n")
	for _, result := range input.Results {
		builder.WriteString("- ")
		builder.WriteString(result.AgentName)
		builder.WriteString(" status=")
		builder.WriteString(result.Status)
		builder.WriteString(" summary=")
		builder.WriteString(compactCEOReviewText(result.Summary))
		builder.WriteString("\n")
	}
	builder.WriteString("changed_files:\n")
	writeCEOReviewChangedFiles(&builder, input.ChangedFiles)
	builder.WriteString("patch_results:\n")
	writeCEOReviewPatches(&builder, input.PatchResults)
	builder.WriteString("patch_previews:\n")
	writeCEOReviewPatches(&builder, input.PatchPreviews)
	builder.WriteString("tool_results:\n")
	writeCEOReviewToolResults(&builder, input.Results)
	builder.WriteString("checks:\n")
	writeCEOReviewChecks(&builder, input.Checks)
	return builder.String()
}

func writeCEOReviewChangedFiles(builder *strings.Builder, changedFiles []string) {
	if len(changedFiles) == 0 {
		builder.WriteString("- none\n")
		return
	}
	for _, path := range changedFiles {
		clean := strings.TrimSpace(path)
		if clean == "" {
			continue
		}
		builder.WriteString("- ")
		builder.WriteString(compactCEOReviewText(clean))
		builder.WriteString("\n")
	}
}

func writeCEOReviewPatches(builder *strings.Builder, patches []workspace.ReplaceTextResult) {
	if len(patches) == 0 {
		builder.WriteString("- none\n")
		return
	}
	for _, patch := range patches {
		builder.WriteString("- path=")
		builder.WriteString(patch.Path)
		builder.WriteString(" diff=")
		builder.WriteString(compactCEOReviewText(patch.Diff))
		builder.WriteString("\n")
	}
}

func writeCEOReviewToolResults(builder *strings.Builder, results []subagent.Result) {
	wrote := false
	for _, result := range results {
		for _, tool := range result.ToolResults {
			wrote = true
			builder.WriteString("- agent=")
			builder.WriteString(result.AgentName)
			builder.WriteString(" action=")
			builder.WriteString(tool.Action)
			builder.WriteString(" status=")
			builder.WriteString(tool.Status)
			if tool.Path != "" {
				builder.WriteString(" path=")
				builder.WriteString(tool.Path)
			}
			if tool.Query != "" {
				builder.WriteString(" query=")
				builder.WriteString(compactCEOReviewText(tool.Query))
			}
			if tool.Output != "" {
				builder.WriteString(" output=")
				builder.WriteString(compactCEOReviewText(tool.Output))
			}
			if tool.Error != "" {
				builder.WriteString(" error=")
				builder.WriteString(compactCEOReviewText(tool.Error))
			}
			builder.WriteString("\n")
		}
	}
	if !wrote {
		builder.WriteString("- none\n")
	}
}

func parseCEOReviewResponse(response model.Response) (*CEOReview, error) {
	var payload ceoReviewPayload
	jsonPayload, ok := model.JSONPayload(response.Text)
	if !ok {
		return nil, fmt.Errorf("parse CEO review JSON: %w", ErrInvalidCEOReview)
	}
	if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
		return nil, fmt.Errorf("parse CEO review JSON: %w", err)
	}
	verdict, err := parseCEORecommendedVerdict(payload.RecommendedVerdict)
	if err != nil {
		return nil, err
	}
	return &CEOReview{
		Source:             "model",
		RecommendedVerdict: verdict,
		Summary:            strings.TrimSpace(payload.Summary),
		PromptBytes:        response.PromptBytes,
	}, nil
}

func parseCEORecommendedVerdict(raw string) (string, error) {
	verdict := strings.ToLower(strings.TrimSpace(raw))
	switch verdict {
	case "pass", "fail":
		return verdict, nil
	default:
		return "", fmt.Errorf("recommended verdict %q: %w", raw, ErrInvalidCEOReview)
	}
}

func applyCEOReviewVerdict(guardVerdict string, review *CEOReview) string {
	if guardVerdict != "pass" {
		return guardVerdict
	}
	if review != nil && review.RecommendedVerdict == "fail" {
		return "fail"
	}
	return guardVerdict
}

func compactCEOReviewText(text string) string {
	const maxReviewTextBytes = 220
	compact := strings.Join(strings.Fields(text), " ")
	if len(compact) <= maxReviewTextBytes {
		return compact
	}
	return compact[:maxReviewTextBytes] + "..."
}
