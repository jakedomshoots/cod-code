package cli

import (
	"context"
	"fmt"

	"ceoharness/internal/history"
)

func readHistoryEntries(ctx context.Context, query historyQuery) (history.Store, []history.Entry, error) {
	store, err := history.New(query.workspaceDir)
	if err != nil {
		return history.Store{}, nil, err
	}
	entries, err := store.ReadByVerdict(ctx, query.verdict)
	if err != nil {
		return history.Store{}, nil, fmt.Errorf("read history: %w", err)
	}
	timeRange, err := parseHistoryRange(query.since, query.until)
	if err != nil {
		return history.Store{}, nil, err
	}
	entries, err = history.FilterByCreatedAtRange(entries, timeRange)
	if err != nil {
		return history.Store{}, nil, fmt.Errorf("filter history by time: %w", err)
	}
	entries = history.FilterByTaskSubstring(entries, query.task)
	entries = history.LimitEntries(entries, query.limit)
	return store, entries, nil
}
