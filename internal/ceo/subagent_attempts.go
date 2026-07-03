package ceo

import (
	"context"
	"errors"
	"time"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func (r Runtime) runSubagentWithAttempts(ctx context.Context, input subagentRunInput) (subagent.Result, error) {
	started := time.Now()
	var result subagent.Result
	attemptErrors := []string{}
	attemptRecords := []subagent.AttemptRecord{}
	lastProviderError := providerErrorFields{}
	packet := input.Packet
	agent := input.Agent
	attempts := input.Attempts
	backoff := input.Backoff
	noProgress := newNoProgressTracker(input.NoProgressStop)
	for attempt := 1; attempt <= attempts; attempt++ {
		nextResult, err := r.runner.Run(ctx, subagent.TaskPacket{
			Task:            packet.Task,
			AgentName:       agent.Name,
			Role:            agent.Role,
			Assignment:      agent.Assignment,
			ProviderName:    agent.ProviderName,
			ContextMode:     packet.ContextPolicy.Mode,
			AllowedActions:  jobpacket.ActionStrings(agent.AllowedActions),
			WorkspaceBrief:  input.WorkspaceBrief,
			PriorFindings:   input.PriorFindings,
			ToolResults:     append([]subagent.ToolResult(nil), input.ToolResults...),
			MaxContextBytes: contextBudgetForAgent(packet, agent),
		})
		if err != nil {
			if ctx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				return subagent.Result{}, subagentContextCanceled(agent.Name, err)
			}
			providerError := providerErrorFieldsFrom(err)
			lastProviderError = providerError
			attemptErrors = append(attemptErrors, err.Error())
			attemptRecords = append(attemptRecords, subagent.AttemptRecord{
				Attempt:              attempt,
				Status:               "fail",
				Error:                err.Error(),
				ProviderErrorKind:    providerError.kind,
				ProviderHTTPStatus:   providerError.httpStatus,
				ProviderRetryAfterMS: providerError.retryAfterMS,
			})
			if shouldExtendSubagentAttempts(err, attempt, attempts) {
				attempts = attempt + 1
			}
			if noProgress.observeError(err.Error()) {
				stopped := failedSubagentResult(packet, agent, attempt, time.Since(started).Milliseconds(), attemptErrors, attemptRecords, lastProviderError)
				return markNoProgressStopped(stopped), nil
			}
			if attempt < attempts {
				if err := waitForRetryBackoff(ctx, retryBackoffForError(err, backoff)); err != nil {
					return subagent.Result{}, subagentContextCanceled(agent.Name, err)
				}
				continue
			}
			return failedSubagentResult(packet, agent, attempt, time.Since(started).Milliseconds(), attemptErrors, attemptRecords, lastProviderError), nil
		}
		nextResult.Attempts = attempt
		nextResult.DurationMS = time.Since(started).Milliseconds()
		if nextResult.PriorFindings == "" && !nextResult.ContextTruncated {
			nextResult.PriorFindings = input.PriorFindings
		}
		nextResult.AttemptErrors = append(append([]string(nil), attemptErrors...), nextResult.AttemptErrors...)
		attemptRecords = append(attemptRecords, subagent.AttemptRecord{
			Attempt: attempt,
			Status:  attemptStatus(nextResult.Status),
		})
		if shouldKeepAttemptRecords(attemptRecords) {
			nextResult.AttemptRecords = append([]subagent.AttemptRecord(nil), attemptRecords...)
		}
		result = nextResult
		if result.Status == "pass" {
			break
		}
		if noProgress.observeResult(result) {
			result = markNoProgressStopped(result)
			break
		}
		if attempt < attempts {
			if err := waitForRetryBackoff(ctx, backoff); err != nil {
				return subagent.Result{}, subagentContextCanceled(agent.Name, err)
			}
		}
	}
	return result, nil
}
