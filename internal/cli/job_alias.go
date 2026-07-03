package cli

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/history"
)

func resolveSavedJobID(ctx context.Context, workspaceDir string, rawID string) (string, error) {
	cleanID := strings.TrimSpace(rawID)
	switch strings.ToLower(cleanID) {
	case "latest", "last":
		return latestSavedJobID(ctx, workspaceDir)
	default:
		return cleanID, nil
	}
}

func latestSavedJobID(ctx context.Context, workspaceDir string) (string, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return "", err
	}
	entries, err := store.ReadRecent(ctx, 1)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 || strings.TrimSpace(entries[0].ID) == "" {
		return "", fmt.Errorf("latest job: %w", history.ErrEntryNotFound)
	}
	return entries[0].ID, nil
}
