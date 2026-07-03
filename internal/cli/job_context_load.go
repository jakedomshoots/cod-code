package cli

import (
	"context"
	"errors"
	"fmt"

	"ceoharness/internal/history"
)

type loadedJobContext struct {
	HistoryPath string
	Source      string
	Context     compactJobContext
}

func loadCompactJobContext(ctx context.Context, workspaceDir string, jobID string) (loadedJobContext, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return loadedJobContext{}, err
	}
	jobID, err = resolveSavedJobID(ctx, workspaceDir, jobID)
	if err != nil {
		return loadedJobContext{}, err
	}
	entry, err := store.FindByID(ctx, jobID)
	if err != nil {
		return loadedJobContext{}, fmt.Errorf("find history job: %w", err)
	}
	loaded := loadedJobContext{
		HistoryPath: store.Path(),
		Source:      "history_entry",
		Context:     contextFromHistoryEntry(entry, workspaceDir),
	}
	reportSnapshot, err := store.ReadReportSnapshot(ctx, jobID)
	if err == nil {
		report, err := decodeJobContextReport(reportSnapshot)
		if err != nil {
			return loadedJobContext{}, err
		}
		loaded.Source = "report_snapshot"
		loaded.Context = contextFromReport(entry, report, workspaceDir)
		return loaded, nil
	}
	if !errors.Is(err, history.ErrEntryNotFound) {
		return loadedJobContext{}, fmt.Errorf("find job context: %w", err)
	}
	return loaded, nil
}
