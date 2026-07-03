package ceo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

func waitForRetryBackoff(ctx context.Context, backoff time.Duration) error {
	if backoff <= 0 {
		return nil
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func shouldExtendSubagentAttempts(err error, attempt int, attempts int) bool {
	return errors.Is(err, model.ErrHTTPRateLimited) && attempt >= attempts && attempts < 2
}

func retryBackoffForError(err error, configured time.Duration) time.Duration {
	providerError := providerErrorFieldsFrom(err)
	retryAfter := time.Duration(providerError.retryAfterMS) * time.Millisecond
	if retryAfter > configured {
		return retryAfter
	}
	return configured
}

func attemptStatus(status string) string {
	if status == "" {
		return "fail"
	}
	return status
}

func shouldKeepAttemptRecords(records []subagent.AttemptRecord) bool {
	if len(records) > 1 {
		return true
	}
	return len(records) == 1 && records[0].Status != "pass"
}

func subagentContextCanceled(agentName string, err error) error {
	return fmt.Errorf("run subagent %s: %w", agentName, err)
}
