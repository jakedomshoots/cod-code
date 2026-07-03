package cli

import (
	"context"
	"io"
)

func runInbox(ctx context.Context, out io.Writer, opts options) error {
	opts.showReviewQueue = true
	opts.reviewDetails = true
	if !opts.reportFormatSet {
		opts.reportFormat = reportFormatText
	}
	return runReviewQueue(ctx, out, reviewQueueRequestFromOptions(opts))
}
