package ceo

import (
	"sort"

	"ceoharness/internal/jobpacket"
)

type subagentStage struct {
	index  int
	agents []indexedSubagent
}

type indexedSubagent struct {
	index int
	agent jobpacket.Subagent
}

func stagedSubagents(agents []jobpacket.Subagent) []subagentStage {
	stages := []subagentStage{}
	for index, agent := range agents {
		stageIndex := stageForAgent(agent)
		stagePosition := findStageIndex(stages, stageIndex)
		if stagePosition < 0 {
			stages = append(stages, subagentStage{index: stageIndex})
			stagePosition = len(stages) - 1
		}
		stages[stagePosition].agents = append(stages[stagePosition].agents, indexedSubagent{
			index: index,
			agent: agent,
		})
	}
	sort.Slice(stages, func(left, right int) bool {
		return stages[left].index < stages[right].index
	})
	return stages
}

func findStageIndex(stages []subagentStage, index int) int {
	for position, stage := range stages {
		if stage.index == index {
			return position
		}
	}
	return -1
}

func stageForAgent(agent jobpacket.Subagent) int {
	if agent.Stage > 0 {
		return agent.Stage
	}
	switch agent.Name {
	case "scanner", "planner", "researcher":
		return 1
	case "coder", "security":
		return 2
	case "reviewer":
		return 3
	default:
		return 2
	}
}
