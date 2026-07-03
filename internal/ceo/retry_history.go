package ceo

import (
	"fmt"
	"strings"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type RepairFailureDetail struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

type RetryHistoryEntry struct {
	Kind              string                `json:"kind"`
	Attempt           int                   `json:"attempt"`
	Status            string                `json:"status"`
	Reason            string                `json:"reason,omitempty"`
	FailedChecks      []RepairFailureDetail `json:"failed_checks,omitempty"`
	CorrectivePrompt  string                `json:"corrective_prompt,omitempty"`
	ModelPatchStatus  string                `json:"model_patch_status,omitempty"`
	ChangedFiles      []string              `json:"changed_files,omitempty"`
	FinalVerdict      string                `json:"final_verdict,omitempty"`
	NoProgressStopped bool                  `json:"no_progress_stopped,omitempty"`
}

func repairFailureDetails(checks []checkrunner.Result, scorerChecks []RepairFailureDetail) []RepairFailureDetail {
	details := make([]RepairFailureDetail, 0, 1+len(scorerChecks))
	if len(checks) > 0 {
		last := checks[len(checks)-1]
		if last.Status != "pass" {
			details = append(details, checkFailureDetail(last))
		}
	}
	for _, check := range scorerChecks {
		if check.Status != "pass" {
			details = append(details, check)
		}
	}
	return details
}

func checkFailureDetail(check checkrunner.Result) RepairFailureDetail {
	message := fmt.Sprintf("exit code %d", check.ExitCode)
	if stderr := strings.TrimSpace(check.Stderr); stderr != "" {
		message = stderr
	}
	return RepairFailureDetail{
		Name:     "command:" + strings.Join(check.Argv, " "),
		Status:   check.Status,
		Message:  message,
		Evidence: strings.TrimSpace(check.Stdout),
	}
}

func ceoReviewFailureDetail(review CEOReview) []RepairFailureDetail {
	return []RepairFailureDetail{{
		Name:    "ceo_review",
		Status:  review.RecommendedVerdict,
		Message: strings.TrimSpace(review.Summary),
	}}
}

func changedFilesFromPatchResults(results []workspace.ReplaceTextResult) []string {
	changed := make([]string, 0, len(results))
	for _, result := range results {
		changed = append(changed, result.Path)
	}
	return changed
}

func patchSignature(patches []PatchRequest) string {
	var builder strings.Builder
	for _, patch := range patches {
		builder.WriteString(patch.Path)
		builder.WriteByte(0)
		builder.WriteString(patch.Old)
		builder.WriteByte(0)
		builder.WriteString(patch.New)
		builder.WriteByte(0)
		builder.WriteString(patch.Content)
		builder.WriteByte(0)
	}
	return builder.String()
}

func failedPatchNoProgressSignature(patches []PatchRequest) string {
	return "patch:" + patchSignature(patches)
}

func applyFailureNoProgressSignature(patches []PatchRequest, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "apply failed"
	}
	return "apply_failed:" + patchSignature(patches) + "\x00" + message
}

func patchesAreNoOp(patches []PatchRequest) bool {
	if len(patches) == 0 {
		return true
	}
	for _, patch := range patches {
		if isCreateFilePatch(patch) || patch.Old != patch.New {
			return false
		}
	}
	return true
}

func markLastRetryNoProgress(results []subagent.Result, history []RetryHistoryEntry) {
	if len(history) > 0 {
		history[len(history)-1].NoProgressStopped = true
	}
	if len(results) > 0 {
		results[len(results)-1] = markNoProgressStopped(results[len(results)-1])
	}
}
