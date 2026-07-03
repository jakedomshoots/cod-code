package ceo

import (
	"context"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

func NewRuntime() Runtime {
	return Runtime{
		runner: subagent.NewRunner(),
		checks: checkrunner.NewRunner(),
	}
}

func NewRuntimeWithSubagentRunner(runner SubagentRunner) Runtime {
	return Runtime{
		runner: runner,
		checks: checkrunner.NewRunner(),
	}
}

func NewRuntimeWithCEOReviewer(reviewer model.Client) Runtime {
	return NewRuntimeWithCEOReviewerAndRoute(reviewer, subagent.RouteMetadata{})
}

func NewRuntimeWithCEOReviewerAndRoute(reviewer model.Client, route subagent.RouteMetadata) Runtime {
	return Runtime{
		runner:           subagent.NewRunner(),
		checks:           checkrunner.NewRunner(),
		ceoReviewer:      reviewer,
		ceoReviewerRoute: route,
	}
}

func NewRuntimeWithSubagentRunnerAndCEOReviewer(runner SubagentRunner, reviewer model.Client) Runtime {
	return NewRuntimeWithSubagentRunnerAndCEOReviewerRoute(runner, reviewer, subagent.RouteMetadata{})
}

func NewRuntimeWithSubagentRunnerAndCEOReviewerRoute(runner SubagentRunner, reviewer model.Client, route subagent.RouteMetadata) Runtime {
	return Runtime{
		runner:           runner,
		checks:           checkrunner.NewRunner(),
		ceoReviewer:      reviewer,
		ceoReviewerRoute: route,
	}
}

func (r Runtime) Run(ctx context.Context, task string) (Report, error) {
	return r.RunJob(ctx, JobRequest{Task: task})
}
