package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errPlanOnlyEventsUnsupported = errors.New("plan-only does not support --format events")

func writePlanOnlyReport(out io.Writer, report planOnlyReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write plan-only report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderPlanOnlyTextReport(report)); err != nil {
			return fmt.Errorf("write text plan-only report: %w", err)
		}
		return nil
	case reportFormatEvents:
		return errPlanOnlyEventsUnsupported
	default:
		return errors.New(reportFormatGuidance)
	}
}

func renderPlanOnlyTextReport(report planOnlyReport) string {
	var builder strings.Builder
	builder.WriteString("Plan-only preview\n")
	if strings.TrimSpace(report.WorkspaceDir) != "" {
		builder.WriteString("Workspace: ")
		builder.WriteString(report.WorkspaceDir)
		builder.WriteByte('\n')
	}
	builder.WriteString("Task: ")
	builder.WriteString(displayTask(report.JobPacket.Task))
	builder.WriteByte('\n')
	if strings.TrimSpace(report.JobOwner) != "" {
		builder.WriteString("Owner: ")
		builder.WriteString(report.JobOwner)
		builder.WriteByte('\n')
	}
	writePlanOnlyContinuation(&builder, report.Continuation)
	builder.WriteString(fmt.Sprintf(
		"Verification: %s (%d required)\n",
		report.VerificationContract.Status,
		report.VerificationContract.RequiredCheckCount,
	))
	builder.WriteString(fmt.Sprintf("Checks: %d configured\n", report.CheckCommandCount))
	builder.WriteString("Subagents: ")
	builder.WriteString(strings.Join(planOnlySubagentNames(report), ", "))
	builder.WriteByte('\n')
	return builder.String()
}

func writePlanOnlyContinuation(builder *strings.Builder, continuation *planOnlyContinuation) {
	if continuation == nil {
		return
	}
	builder.WriteString(fmt.Sprintf(
		"Continuation: %s saved_delegation=%t planned=%d reusable=%d\n",
		continuation.JobID,
		continuation.UseSavedDelegation,
		continuation.PlannedSubagentCount,
		continuation.ReusableSubagentCount,
	))
}

func planOnlySubagentNames(report planOnlyReport) []string {
	names := make([]string, 0, len(report.JobPacket.Subagents))
	for _, subagent := range report.JobPacket.Subagents {
		if strings.TrimSpace(subagent.Name) == "" {
			continue
		}
		names = append(names, subagent.Name)
	}
	return names
}
