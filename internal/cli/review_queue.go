package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/history"
)

type reviewQueueReport struct {
	HistoryPath string           `json:"history_path"`
	Queue       []reviewQueueRow `json:"review_queue"`
	TotalCount  int              `json:"total_count"`
	Details     bool             `json:"details,omitempty"`
	Task        string           `json:"task_filter,omitempty"`
	Limit       int              `json:"limit,omitempty"`
	Since       string           `json:"since,omitempty"`
	Until       string           `json:"until,omitempty"`
}

type reviewQueueRow struct {
	historyRow
	ReviewReason     string             `json:"review_reason"`
	SuggestedCommand string             `json:"suggested_command"`
	ReviewContext    *compactJobContext `json:"review_context,omitempty"`
}

type reviewQueueRequest struct {
	Query          historyQuery
	Format         reportFormat
	IncludeDetails bool
}

type reviewQueueRowsRequest struct {
	Entries      []history.Entry
	Judgments    map[string]history.HumanJudgment
	WorkspaceDir string
}

func runReviewQueue(ctx context.Context, out io.Writer, req reviewQueueRequest) error {
	query := req.Query
	readQuery := query
	readQuery.limit = 0
	store, entries, err := readHistoryEntries(ctx, readQuery)
	if err != nil {
		return err
	}
	judgments, err := readHumanJudgmentsForHistory(ctx, store, entries)
	if err != nil {
		return err
	}
	rows := buildReviewQueueRows(reviewQueueRowsRequest{
		Entries:      entries,
		Judgments:    judgments,
		WorkspaceDir: query.workspaceDir,
	})
	rows = limitReviewQueueRows(rows, query.limit)
	if req.IncludeDetails {
		err := addReviewQueueDetails(ctx, rows, query.workspaceDir)
		if err != nil {
			return err
		}
	}
	report := reviewQueueReport{
		HistoryPath: store.Path(),
		Queue:       rows,
		TotalCount:  len(rows),
		Details:     req.IncludeDetails,
		Task:        query.task,
		Limit:       query.limit,
		Since:       query.since,
		Until:       query.until,
	}
	return writeReviewQueueReport(out, report, req.Format)
}

func writeReviewQueueReport(out io.Writer, report reviewQueueReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		return encodeReviewQueueJSON(out, report)
	case reportFormatText:
		if _, err := io.WriteString(out, renderReviewQueueText(report)); err != nil {
			return fmt.Errorf("write text review queue: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderReviewQueueText(report reviewQueueReport) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Review queue: %d %s\n", report.TotalCount, pluralize("job", report.TotalCount)))
	for _, row := range report.Queue {
		builder.WriteString(fmt.Sprintf("- %s [%s] %s\n", row.ID, row.RecoveryState, trimText(oneLine(row.Task), 100)))
		builder.WriteString(fmt.Sprintf("  Verdict: %s\n", row.LastVerdict))
		builder.WriteString(fmt.Sprintf("  Retryable: %s\n", yesNo(row.Retryable)))
		builder.WriteString(fmt.Sprintf("  Evidence: %s\n", row.EvidencePointer))
		if row.HumanJudgment != nil {
			builder.WriteString(fmt.Sprintf("  Human: %s\n", row.HumanJudgment.Verdict))
		}
		builder.WriteString("  Next: ")
		builder.WriteString(row.SuggestedCommand)
		builder.WriteString("\n")
		writeReviewContextText(&builder, row.ReviewContext)
	}
	if report.TotalCount == 0 {
		builder.WriteString("No jobs need human attention.\n")
	}
	return builder.String()
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func encodeReviewQueueJSON(out io.Writer, report reviewQueueReport) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write review queue report: %w", err)
	}
	return nil
}

func buildReviewQueueRows(req reviewQueueRowsRequest) []reviewQueueRow {
	rows := []reviewQueueRow{}
	for _, entry := range req.Entries {
		reason := reviewQueueReason(entry, req.Judgments[entry.ID])
		if reason == "" {
			continue
		}
		row := reviewQueueRow{
			historyRow:       historyRowsWithJudgments([]history.Entry{entry}, req.Judgments)[0],
			ReviewReason:     reason,
			SuggestedCommand: suggestedReviewCommand(entry.ID, reason, req.WorkspaceDir),
		}
		rows = append(rows, row)
	}
	return rows
}

func addReviewQueueDetails(ctx context.Context, rows []reviewQueueRow, workspaceDir string) error {
	for index := range rows {
		context, err := loadCompactJobContext(ctx, workspaceDir, rows[index].ID)
		if err != nil {
			return fmt.Errorf("load review context for %s: %w", rows[index].ID, err)
		}
		rows[index].ReviewContext = &context.Context
	}
	return nil
}

func reviewQueueReason(entry history.Entry, judgment history.HumanJudgment) string {
	if judgment.Verdict == "accept" {
		return ""
	}
	if judgment.Verdict == "reject" {
		return "human_rejected"
	}
	switch entry.Verdict {
	case "needs_input":
		return "needs_input"
	case "pass":
		return "awaiting_human_judgment"
	case "":
		return "unknown_verdict"
	default:
		return "failed_or_unresolved"
	}
}

func suggestedReviewCommand(jobID string, reason string, workspaceDir string) string {
	prefix := fmt.Sprintf("cod --workspace %q", workspaceDir)
	switch reason {
	case "needs_input":
		return fmt.Sprintf("%s --resume %s --answer %q", prefix, jobID, "...")
	case "awaiting_human_judgment":
		return fmt.Sprintf("%s --judge-job %s --human-verdict accept", prefix, jobID)
	default:
		return fmt.Sprintf("%s --rerun %s", prefix, jobID)
	}
}

func limitReviewQueueRows(rows []reviewQueueRow, limit int) []reviewQueueRow {
	if limit < 1 || limit >= len(rows) {
		return rows
	}
	return rows[len(rows)-limit:]
}
