package ceo

import (
	"context"
	"time"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type scheduledSubagentResult struct {
	index  int
	result subagent.Result
	err    error
}

type subagentRunInput struct {
	Packet         jobpacket.Packet
	Agent          jobpacket.Subagent
	Attempts       int
	Backoff        time.Duration
	NoProgressStop int
	WorkspaceBrief string
	PriorFindings  string
	ToolResults    []subagent.ToolResult
}

type subagentsRunInput struct {
	Packet         jobpacket.Packet
	ToolState      toolRequestState
	Attempts       int
	BackoffMS      int
	NoProgressStop int
	Concurrency    int
	WorkspaceBrief string
	MaxOutputBytes int
	Continuation   *ContinuationContext
}

func (r Runtime) runSubagents(ctx context.Context, input subagentsRunInput) ([]subagent.Result, error) {
	packet := input.Packet
	attempts := input.Attempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(input.BackoffMS) * time.Millisecond
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make([]subagent.Result, len(packet.Subagents))
	priorResults := []subagent.Result{}
	priorFindings := ""
	for _, stage := range stagedSubagents(packet.Subagents) {
		stageResults, remainingStage := reusableStageResults(stage, input.Continuation)
		toolState := input.ToolState
		toolState.PriorResults = append([]subagent.Result(nil), priorResults...)
		freshResults, err := r.runSubagentStage(runCtx, subagentStageInput{
			Packet:         packet,
			Stage:          remainingStage,
			Attempts:       attempts,
			Backoff:        backoff,
			NoProgressStop: input.NoProgressStop,
			Concurrency:    input.Concurrency,
			WorkspaceBrief: input.WorkspaceBrief,
			PriorFindings:  priorFindings,
			ToolState:      toolState,
		})
		if err != nil {
			return nil, err
		}
		stageResults = append(stageResults, freshResults...)
		stageResults = compactScheduledSubagentOutputs(stageResults, input.MaxOutputBytes)
		for _, result := range stageResults {
			results[result.index] = result.result
		}
		priorResults = appendPriorStageResults(priorResults, stageResults)
		priorFindings = renderPriorFindings(priorResults)
		if hasNeedsInputResult(stageResults) {
			return compactSubagentResults(results), nil
		}
	}
	return results, nil
}

type subagentStageInput struct {
	Packet         jobpacket.Packet
	Stage          subagentStage
	ToolState      toolRequestState
	Attempts       int
	Backoff        time.Duration
	NoProgressStop int
	Concurrency    int
	WorkspaceBrief string
	PriorFindings  string
}

func (r Runtime) runSubagentStage(ctx context.Context, input subagentStageInput) ([]scheduledSubagentResult, error) {
	results, err := r.collectSubagentStageResults(ctx, input)
	if err != nil {
		return nil, err
	}
	return r.runStageToolFeedback(ctx, input, results)
}

func (r Runtime) runStageToolFeedback(ctx context.Context, input subagentStageInput, results []scheduledSubagentResult) ([]scheduledSubagentResult, error) {
	stageResults := scheduledResultsOnly(results)
	withTools, err := r.runSubagentToolRequests(ctx, stageResults, input.ToolState)
	if err != nil {
		return nil, err
	}
	withFeedback, err := r.runSubagentToolFeedback(ctx, toolFeedbackInput{
		Packet:         input.Packet,
		Results:        withTools,
		Attempts:       input.Attempts,
		BackoffMS:      int(input.Backoff / time.Millisecond),
		NoProgressStop: input.NoProgressStop,
		WorkspaceBrief: input.WorkspaceBrief,
	})
	if err != nil {
		return nil, err
	}
	next := make([]scheduledSubagentResult, 0, len(results))
	for index, result := range results {
		result.result = withFeedback[index]
		next = append(next, result)
	}
	return next, nil
}

func scheduledResultsOnly(results []scheduledSubagentResult) []subagent.Result {
	out := make([]subagent.Result, 0, len(results))
	for _, result := range results {
		out = append(out, result.result)
	}
	return out
}
