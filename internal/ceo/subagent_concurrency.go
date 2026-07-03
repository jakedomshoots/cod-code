package ceo

import (
	"context"

	"ceoharness/internal/jobpacket"
)

func (r Runtime) collectSubagentStageResults(ctx context.Context, input subagentStageInput) ([]scheduledSubagentResult, error) {
	stage := input.Stage
	if len(stage.agents) == 0 {
		return []scheduledSubagentResult{}, nil
	}
	resultCh := make(chan scheduledSubagentResult, len(stage.agents))
	nextAgent := 0
	active := 0
	limit := stageConcurrencyLimit(input.Concurrency, len(stage.agents))
	for active < limit && nextAgent < len(stage.agents) {
		r.runStageAgent(ctx, input, stage.agents[nextAgent], resultCh)
		nextAgent++
		active++
	}

	results := make([]scheduledSubagentResult, len(stage.agents))
	var firstErr error
	for received := 0; received < len(stage.agents) && active > 0; received++ {
		next := <-resultCh
		active--
		if next.err != nil && firstErr == nil {
			firstErr = next.err
		} else {
			results[stageAgentPosition(stage.agents, next.index)] = next
		}
		if firstErr == nil && nextAgent < len(stage.agents) {
			r.runStageAgent(ctx, input, stage.agents[nextAgent], resultCh)
			nextAgent++
			active++
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (r Runtime) runStageAgent(ctx context.Context, input subagentStageInput, item indexedSubagent, resultCh chan<- scheduledSubagentResult) {
	go func() {
		result, err := r.runSubagentWithAttempts(ctx, subagentRunInput{
			Packet:         input.Packet,
			Agent:          item.agent,
			Attempts:       input.Attempts,
			Backoff:        input.Backoff,
			NoProgressStop: input.NoProgressStop,
			WorkspaceBrief: input.WorkspaceBrief,
			PriorFindings:  input.PriorFindings,
		})
		result.Stage = input.Stage.index
		result.AllowedActions = jobpacket.ActionStrings(item.agent.AllowedActions)
		resultCh <- scheduledSubagentResult{
			index:  item.index,
			result: result,
			err:    err,
		}
	}()
}

func stageConcurrencyLimit(configured int, agentCount int) int {
	if agentCount < 1 {
		return 0
	}
	if configured < 1 || configured > agentCount {
		return agentCount
	}
	return configured
}

func stageAgentPosition(agents []indexedSubagent, index int) int {
	for position, agent := range agents {
		if agent.index == index {
			return position
		}
	}
	return 0
}
