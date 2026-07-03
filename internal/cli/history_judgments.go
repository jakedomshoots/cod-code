package cli

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"ceoharness/internal/history"
)

type historyRow struct {
	history.Entry
	HumanJudgment     *history.HumanJudgment `json:"human_judgment,omitempty"`
	HumanJudgmentPath string                 `json:"human_judgment_path,omitempty"`
	RecoveryState     string                 `json:"recovery_state"`
	LastVerdict       string                 `json:"last_verdict"`
	Retryable         bool                   `json:"retryable"`
	NextAction        string                 `json:"next_action"`
	EvidencePointer   string                 `json:"evidence_pointer"`
}

type recoveryView struct {
	State           string
	LastVerdict     string
	Retryable       bool
	NextAction      string
	EvidencePointer string
}

func readHumanJudgmentsForHistory(ctx context.Context, store history.Store, entries []history.Entry) (map[string]history.HumanJudgment, error) {
	judgments := map[string]history.HumanJudgment{}
	for _, entry := range entries {
		judgment, err := store.ReadHumanJudgment(ctx, entry.ID)
		if err != nil {
			if errors.Is(err, history.ErrEntryNotFound) {
				continue
			}
			return nil, fmt.Errorf("read human judgment for %s: %w", entry.ID, err)
		}
		judgments[entry.ID] = judgment
	}
	return judgments, nil
}

func historyRowsWithJudgments(entries []history.Entry, judgments map[string]history.HumanJudgment) []historyRow {
	rows := make([]historyRow, 0, len(entries))
	for _, entry := range entries {
		judgment, judged := judgments[entry.ID]
		recovery := buildRecoveryView(entry, judgment, judged)
		row := historyRow{
			Entry:           entry,
			RecoveryState:   recovery.State,
			LastVerdict:     recovery.LastVerdict,
			Retryable:       recovery.Retryable,
			NextAction:      recovery.NextAction,
			EvidencePointer: recovery.EvidencePointer,
		}
		if judged {
			row.HumanJudgment = &judgment
			if path, err := history.HumanJudgmentPath(entry.ID); err == nil {
				row.HumanJudgmentPath = path
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func buildRecoveryView(entry history.Entry, judgment history.HumanJudgment, judged bool) recoveryView {
	lastVerdict := strings.TrimSpace(entry.Verdict)
	if judged {
		switch judgment.Verdict {
		case "accept":
			return recoveryView{
				State:           "accepted",
				LastVerdict:     "accept",
				NextAction:      "none",
				EvidencePointer: reportEvidencePointer(entry.ID),
			}
		case "reject":
			return recoveryView{
				State:           "rejected",
				LastVerdict:     "reject",
				NextAction:      fallbackNextAction(entry, "rerun after rejection"),
				EvidencePointer: reportEvidencePointer(entry.ID),
			}
		}
	}
	switch lastVerdict {
	case "needs_input":
		return recoveryView{
			State:           "needs-input",
			LastVerdict:     lastVerdict,
			NextAction:      fallbackNextAction(entry, "answer required input"),
			EvidencePointer: reportEvidencePointer(entry.ID),
		}
	case "pass":
		return recoveryView{
			State:           "waiting-review",
			LastVerdict:     lastVerdict,
			NextAction:      fallbackNextAction(entry, "judge result"),
			EvidencePointer: reportEvidencePointer(entry.ID),
		}
	case "fail":
		return recoveryView{
			State:           "failed",
			LastVerdict:     lastVerdict,
			Retryable:       true,
			NextAction:      fallbackNextAction(entry, "rerun job"),
			EvidencePointer: reportEvidencePointer(entry.ID),
		}
	default:
		if lastVerdict == "" {
			lastVerdict = "unknown"
		}
		return recoveryView{
			State:           "failed",
			LastVerdict:     lastVerdict,
			Retryable:       true,
			NextAction:      fallbackNextAction(entry, "inspect job"),
			EvidencePointer: reportEvidencePointer(entry.ID),
		}
	}
}

func fallbackNextAction(entry history.Entry, fallback string) string {
	if next := strings.TrimSpace(entry.ExecutionPlanNextAction); next != "" {
		return next
	}
	return fallback
}

func reportEvidencePointer(jobID string) string {
	return filepath.Join(history.JobReportDir, jobID+".json")
}
