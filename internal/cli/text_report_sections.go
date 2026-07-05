package cli

import (
	"fmt"
	"strconv"
	"strings"

	"ceoharness/internal/ceo"
)

func writeTextCEODelegation(builder *strings.Builder, report ceo.Report) {
	if report.CEODelegation == nil {
		return
	}
	delegation := report.CEODelegation
	source := strings.TrimSpace(delegation.Source)
	if source == "" {
		source = "model"
	}
	builder.WriteString("Delegation: ")
	builder.WriteString(source)
	if len(delegation.SelectedSubagents) > 0 {
		builder.WriteString(" selected=")
		builder.WriteString(strings.Join(delegation.SelectedSubagents, ","))
	}
	builder.WriteString("\n")
}

func writeTextPatchApproval(builder *strings.Builder, report ceo.Report) {
	if report.PatchApproval == nil {
		return
	}
	approval := report.PatchApproval
	builder.WriteString(fmt.Sprintf(
		"Patch approval: %s digest=%s previews=%d",
		approval.Status,
		approval.PreviewDigest,
		approval.PreviewCount,
	))
	if strings.TrimSpace(approval.ApprovedDigest) != "" {
		builder.WriteString(" approved=" + approval.ApprovedDigest)
	}
	builder.WriteString("\n")
}

func writeTextRunLedger(builder *strings.Builder, report ceo.Report) {
	ledger := report.RunLedger
	if ledger.Owner == "" && ledger.NextAction == "" && ledger.VerificationStatus == "" && ledger.ChangedFileCount == 0 && ledger.ProviderRouteCount == 0 {
		return
	}
	builder.WriteString(fmt.Sprintf(
		"Progress: owner=%s next=%s verification=%s changed=%d provider-routes=%d",
		ledger.Owner,
		ledger.NextAction,
		ledger.VerificationStatus,
		ledger.ChangedFileCount,
		ledger.ProviderRouteCount,
	))
	if len(ledger.ProviderRouteReasons) > 0 {
		builder.WriteString(" reasons=" + strings.Join(ledger.ProviderRouteReasons, ","))
	}
	builder.WriteString("\n")
}

func writeTextVerificationContract(builder *strings.Builder, report ceo.Report) {
	contract := report.VerificationContract
	if contract.Status == "" {
		return
	}
	builder.WriteString(fmt.Sprintf(
		"Verification: %s (%d %s, %d %s)\n",
		contract.Status,
		contract.RequiredCheckCount,
		"required",
		contract.CheckAttemptCount,
		pluralize("attempt", contract.CheckAttemptCount),
	))
}

func writeTextCheckSummary(builder *strings.Builder, report ceo.Report) {
	if len(report.CheckResults) == 0 {
		return
	}
	last := report.CheckResults[len(report.CheckResults)-1]
	builder.WriteString(fmt.Sprintf("\nChecks: %d %s, last %s\n", len(report.CheckResults), pluralize("run", len(report.CheckResults)), last.Status))
}

func writeTextChangedFiles(builder *strings.Builder, report ceo.Report) {
	if len(report.ChangedFiles) == 0 {
		return
	}
	builder.WriteString("Changed: " + strings.Join(report.ChangedFiles, ", ") + "\n")
	if strings.TrimSpace(report.ExecutionPlan.NextAction) != "" {
		builder.WriteString("Next action: " + report.ExecutionPlan.NextAction + "\n")
	}
}

func writeTextQuestions(builder *strings.Builder, req reportOutputRequest) {
	questions := reportQuestions(req.Report)
	if len(questions) == 0 {
		return
	}
	builder.WriteString("\nQuestions:\n")
	for _, question := range questions {
		builder.WriteString("- " + question + "\n")
	}
	if req.Report.JobID != "" {
		builder.WriteString("Resume: cod")
		if strings.TrimSpace(req.WorkspaceDir) != "" {
			builder.WriteString(" --workspace " + strconv.Quote(req.WorkspaceDir))
		}
		builder.WriteString(" --resume " + req.Report.JobID + " --answer " + strconv.Quote("<your answer>") + "\n")
	}
}

func reportQuestions(report ceo.Report) []string {
	questions := []string{}
	seen := map[string]struct{}{}
	for _, result := range report.SubagentResults {
		for _, question := range result.Questions {
			cleanQuestion := strings.TrimSpace(question)
			if cleanQuestion == "" {
				continue
			}
			if _, ok := seen[cleanQuestion]; ok {
				continue
			}
			seen[cleanQuestion] = struct{}{}
			questions = append(questions, cleanQuestion)
		}
	}
	return questions
}
