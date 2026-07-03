package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
)

type contextTraceLookupReport struct {
	HistoryPath  string                 `json:"history_path"`
	Source       string                 `json:"source"`
	ContextTrace contextTraceLookupBody `json:"context_trace"`
}

type contextTraceLookupBody struct {
	JobID  string                  `json:"job_id"`
	Source string                  `json:"source"`
	Agents []ceo.ContextTraceEntry `json:"agents"`
}

func runContextTraceLookup(ctx context.Context, out io.Writer, workspaceDir string, jobID string, format reportFormat) error {
	loaded, err := loadContextTrace(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	report := contextTraceLookupReport{
		HistoryPath: loaded.HistoryPath,
		Source:      loaded.Source,
		ContextTrace: contextTraceLookupBody{
			JobID:  loaded.JobID,
			Source: loaded.Source,
			Agents: loaded.Agents,
		},
	}
	return writeContextTraceLookupReport(out, report, format)
}

type loadedContextTrace struct {
	HistoryPath string
	Source      string
	JobID       string
	Agents      []ceo.ContextTraceEntry
}

func loadContextTrace(ctx context.Context, workspaceDir string, jobID string) (loadedContextTrace, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return loadedContextTrace{}, err
	}
	jobID, err = resolveSavedJobID(ctx, workspaceDir, jobID)
	if err != nil {
		return loadedContextTrace{}, err
	}
	payload, err := store.ReadReportSnapshot(ctx, jobID)
	if err != nil {
		return loadedContextTrace{}, fmt.Errorf("read context trace snapshot: %w", err)
	}
	report, err := decodeJobContextReport(payload)
	if err != nil {
		return loadedContextTrace{}, err
	}
	if len(report.ContextTrace) == 0 {
		return loadedContextTrace{}, fmt.Errorf("context trace for %s: %w", jobID, history.ErrEntryNotFound)
	}
	if strings.TrimSpace(report.JobID) != "" {
		jobID = report.JobID
	}
	return loadedContextTrace{
		HistoryPath: store.Path(),
		Source:      "report_snapshot",
		JobID:       jobID,
		Agents:      report.ContextTrace,
	}, nil
}

func writeContextTraceLookupReport(out io.Writer, report contextTraceLookupReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write context trace: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderContextTraceText(report)); err != nil {
			return fmt.Errorf("write text context trace: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderContextTraceText(report contextTraceLookupReport) string {
	var builder strings.Builder
	builder.WriteString("Context trace: ")
	builder.WriteString(report.ContextTrace.JobID)
	builder.WriteString("\n")
	writeJobContextLine(&builder, "Source", report.Source)
	for _, agent := range report.ContextTrace.Agents {
		builder.WriteString("- ")
		builder.WriteString(agent.AgentName)
		if agent.Role != "" {
			builder.WriteString(" ")
			builder.WriteString(agent.Role)
		}
		builder.WriteString(fmt.Sprintf(" budget=%dB actual=%dB", agent.MaxContextBytes, agent.ContextBytes))
		if agent.ContextTruncated {
			builder.WriteString(" truncated=true")
		}
		if agent.TaskSummary != "" {
			builder.WriteString(" task=" + agent.TaskSummary)
		}
		if agent.PriorFindings.Count > 0 {
			builder.WriteString(fmt.Sprintf(" prior_findings=%d/%dB", agent.PriorFindings.Count, agent.PriorFindings.Bytes))
		}
		if len(agent.ExcludedContent.WorkspaceExcludes) > 0 {
			builder.WriteString(" excludes=" + strings.Join(agent.ExcludedContent.WorkspaceExcludes, ","))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
