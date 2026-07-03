package jobpacket

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrTaskRequired    = errors.New("task is required")
	ErrInvalidSubagent = errors.New("invalid subagent")
)

const (
	MaxDelegatedSubagents  = 8
	DefaultMaxSubagents    = 3
	DefaultMaxContextBytes = 4096
)

type Packet struct {
	Task             string        `json:"task"`
	TaskProfile      TaskProfile   `json:"task_profile"`
	MaxSubagents     int           `json:"max_subagents"`
	CEO              CEO           `json:"ceo"`
	ContextPolicy    ContextPolicy `json:"context_policy"`
	Subagents        []Subagent    `json:"subagents"`
	EvidenceRequired []string      `json:"evidence_required"`
}

type CEO struct {
	Authority string `json:"authority"`
}

type ContextPolicy struct {
	Mode     string `json:"mode"`
	MaxBytes int    `json:"max_bytes"`
}

type TaskProfile struct {
	Kind      string   `json:"kind"`
	RiskLevel string   `json:"risk_level"`
	RiskAreas []string `json:"risk_areas,omitempty"`
}

type Subagent struct {
	Name            string   `json:"name"`
	Role            string   `json:"role"`
	Assignment      string   `json:"assignment,omitempty"`
	ProviderName    string   `json:"provider,omitempty"`
	Stage           int      `json:"stage,omitempty"`
	MaxContextBytes int      `json:"max_context_bytes,omitempty"`
	AllowedActions  []Action `json:"allowed_actions,omitempty"`
}

type BuildOptions struct {
	Task            string
	Subagents       []Subagent
	MaxSubagents    int
	MaxContextBytes int
}

func Build(task string) (Packet, error) {
	return BuildWithSubagents(task, nil)
}

func BuildWithSubagents(task string, subagents []Subagent) (Packet, error) {
	return BuildWithOptions(BuildOptions{Task: task, Subagents: subagents})
}

func BuildWithOptions(opts BuildOptions) (Packet, error) {
	cleanTask := strings.TrimSpace(opts.Task)
	if cleanTask == "" {
		return Packet{}, ErrTaskRequired
	}
	maxContextBytes := opts.MaxContextBytes
	if maxContextBytes < 0 {
		return Packet{}, fmt.Errorf("max context bytes %d: %w", maxContextBytes, ErrInvalidSubagent)
	}
	if maxContextBytes == 0 {
		maxContextBytes = DefaultMaxContextBytes
	}
	maxSubagents := opts.MaxSubagents
	if maxSubagents < 0 || maxSubagents > MaxDelegatedSubagents {
		return Packet{}, fmt.Errorf("max subagents %d: %w", maxSubagents, ErrInvalidSubagent)
	}
	if maxSubagents == 0 && len(opts.Subagents) == 0 {
		maxSubagents = DefaultMaxSubagents
	}
	taskProfile := classifyTaskProfile(cleanTask)
	delegatedSubagents, err := subagentsForProfile(taskProfile, opts.Subagents, maxSubagents)
	if err != nil {
		return Packet{}, err
	}

	return Packet{
		Task:         cleanTask,
		TaskProfile:  taskProfile,
		MaxSubagents: len(delegatedSubagents),
		CEO: CEO{
			Authority: "final",
		},
		ContextPolicy: ContextPolicy{
			Mode:     "lean",
			MaxBytes: maxContextBytes,
		},
		Subagents: delegatedSubagents,
		EvidenceRequired: []string{
			"changed files",
			"test output",
			"risks",
		},
	}, nil
}

func subagentsForProfile(profile TaskProfile, subagents []Subagent, maxSubagents int) ([]Subagent, error) {
	if len(subagents) == 0 {
		return defaultSubagents(profile, maxSubagents), nil
	}
	normalized, err := NormalizeCustomSubagents(subagents)
	if err != nil {
		return nil, err
	}
	return limitSubagents(profile, normalized, maxSubagents), nil
}

func NormalizeCustomSubagents(subagents []Subagent) ([]Subagent, error) {
	if len(subagents) > MaxDelegatedSubagents {
		return nil, fmt.Errorf("subagents count %d: %w", len(subagents), ErrInvalidSubagent)
	}
	normalized := make([]Subagent, 0, len(subagents))
	seen := map[string]struct{}{}
	for index, subagent := range subagents {
		name := strings.TrimSpace(subagent.Name)
		role := strings.TrimSpace(subagent.Role)
		assignment := strings.TrimSpace(subagent.Assignment)
		providerName := strings.TrimSpace(subagent.ProviderName)
		if name == "" || role == "" {
			return nil, fmt.Errorf("subagents[%d]: %w", index, ErrInvalidSubagent)
		}
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("subagents[%d] duplicate name %q: %w", index, name, ErrInvalidSubagent)
		}
		actions, ok := NormalizeActions(subagent.AllowedActions)
		if !ok {
			return nil, fmt.Errorf("subagents[%d] allowed actions: %w", index, ErrInvalidSubagent)
		}
		if len(actions) == 0 {
			actions = DefaultActionsForAgent(name)
		}
		if !ValidStage(subagent.Stage) {
			return nil, fmt.Errorf("subagents[%d] stage: %w", index, ErrInvalidSubagent)
		}
		if subagent.MaxContextBytes < 0 {
			return nil, fmt.Errorf("subagents[%d] max context bytes: %w", index, ErrInvalidSubagent)
		}
		seen[name] = struct{}{}
		normalized = append(normalized, Subagent{Name: name, Role: role, Assignment: assignment, ProviderName: providerName, Stage: subagent.Stage, MaxContextBytes: subagent.MaxContextBytes, AllowedActions: actions})
	}
	return normalized, nil
}
