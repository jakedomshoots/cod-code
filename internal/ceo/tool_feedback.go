package ceo

import (
	"context"
	"fmt"
	"time"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

type toolFeedbackInput struct {
	Packet         jobpacket.Packet
	Results        []subagent.Result
	Attempts       int
	BackoffMS      int
	NoProgressStop int
	WorkspaceBrief string
}

type feedbackPassInput struct {
	Packet         jobpacket.Packet
	Agent          jobpacket.Subagent
	Prior          subagent.Result
	Attempts       int
	Backoff        time.Duration
	NoProgressStop int
	WorkspaceBrief string
}

func (r Runtime) runSubagentToolFeedback(ctx context.Context, input toolFeedbackInput) ([]subagent.Result, error) {
	next := append([]subagent.Result(nil), input.Results...)
	agents := subagentByName(input.Packet.Subagents)
	backoff := time.Duration(input.BackoffMS) * time.Millisecond
	for index, result := range next {
		if !hasFeedbackToolResults(result.ToolResults) {
			continue
		}
		agent, ok := agents[result.AgentName]
		if !ok {
			continue
		}
		followUp, err := r.runSubagentFeedbackPass(ctx, feedbackPassInput{
			Packet:         input.Packet,
			Agent:          agent,
			Prior:          result,
			Attempts:       input.Attempts,
			Backoff:        backoff,
			NoProgressStop: input.NoProgressStop,
			WorkspaceBrief: input.WorkspaceBrief,
		})
		if err != nil {
			return nil, err
		}
		next[index] = followUp
	}
	return next, nil
}

func (r Runtime) runSubagentFeedbackPass(ctx context.Context, input feedbackPassInput) (subagent.Result, error) {
	attempts := input.Attempts
	if attempts < 1 {
		attempts = 1
	}
	followUp, err := r.runSubagentWithAttempts(ctx, subagentRunInput{
		Packet:         input.Packet,
		Agent:          input.Agent,
		Attempts:       attempts,
		Backoff:        input.Backoff,
		NoProgressStop: input.NoProgressStop,
		WorkspaceBrief: input.WorkspaceBrief,
		PriorFindings:  input.Prior.PriorFindings,
		ToolResults:    input.Prior.ToolResults,
	})
	if err != nil {
		return subagent.Result{}, fmt.Errorf("run tool feedback for %s: %w", input.Agent.Name, err)
	}
	followUp.Stage = input.Prior.Stage
	followUp.AllowedActions = append([]string(nil), input.Prior.AllowedActions...)
	followUp.ToolRequests = append([]subagent.ToolRequest(nil), input.Prior.ToolRequests...)
	followUp.ToolResults = append([]subagent.ToolResult(nil), input.Prior.ToolResults...)
	followUp.InitialSummary = input.Prior.Summary
	followUp.ToolFeedbackPasses = input.Prior.ToolFeedbackPasses + 1
	followUp.PriorFindings = input.Prior.PriorFindings
	return followUp, nil
}

func subagentByName(subagents []jobpacket.Subagent) map[string]jobpacket.Subagent {
	agents := make(map[string]jobpacket.Subagent, len(subagents))
	for _, agent := range subagents {
		agents[agent.Name] = agent
	}
	return agents
}

func hasFeedbackToolResults(results []subagent.ToolResult) bool {
	for _, result := range results {
		switch result.Status {
		case "pass", "fail", "error":
			return true
		}
	}
	return false
}
