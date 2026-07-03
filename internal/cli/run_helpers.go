package cli

import (
	"context"
	"time"

	"ceoharness/internal/ceo"
)

func contextWithJobTimeout(ctx context.Context, timeoutMS int) (context.Context, func()) {
	if timeoutMS <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
}

func verdictError(report ceo.Report) error {
	if report.Verdict == "needs_input" {
		return ErrVerdictNeedsInput
	}
	if report.Verdict != "pass" {
		return ErrVerdictFailed
	}

	return nil
}
