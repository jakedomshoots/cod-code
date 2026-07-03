package ceo

import (
	"context"
	"encoding/json"
	"fmt"

	"ceoharness/internal/history"
)

func persistWorkspaceReport(ctx context.Context, workspaceDir string, report Report) (Report, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return Report{}, fmt.Errorf("open history: %w", err)
	}
	stored, err := store.Append(ctx, buildHistoryEntry(report))
	if err != nil {
		return Report{}, fmt.Errorf("append history: %w", err)
	}
	report.HistoryPath = store.Path()
	report.JobID = stored.ID
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return Report{}, fmt.Errorf("encode report snapshot: %w", err)
	}
	if _, err := store.SaveReportSnapshot(ctx, stored.ID, payload); err != nil {
		return Report{}, fmt.Errorf("save report snapshot: %w", err)
	}
	return report, nil
}
