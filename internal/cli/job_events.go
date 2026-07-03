package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
)

type savedJobEvents struct {
	RunEvents []ceo.RunEvent `json:"run_events"`
}

func runJobEventsLookup(ctx context.Context, out io.Writer, workspaceDir string, jobID string) error {
	store, err := history.New(workspaceDir)
	if err != nil {
		return err
	}
	jobID, err = resolveSavedJobID(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	snapshot, err := store.ReadReportSnapshotWithMetadata(ctx, jobID)
	if err != nil {
		return fmt.Errorf("find job events: %w", err)
	}
	var events savedJobEvents
	if err := json.Unmarshal(snapshot.Payload, &events); err != nil {
		return fmt.Errorf("decode job events: %w", err)
	}
	if err := writeRunEvents(out, events.RunEvents); err != nil {
		return fmt.Errorf("write job events: %w", err)
	}
	return nil
}
