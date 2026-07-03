package ceo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/model"
)

var ErrInvalidCEODelegation = errors.New("invalid CEO delegation")

type ceoDelegationPayload struct {
	SelectedSubagents []string             `json:"selected_subagents"`
	NewSubagents      []jobpacket.Subagent `json:"new_subagents"`
	Assignments       map[string]string    `json:"assignments"`
	Summary           string               `json:"summary"`
}

func (r Runtime) runCEODelegation(ctx context.Context, packet jobpacket.Packet, enabled bool) (jobpacket.Packet, *CEODelegation, error) {
	if !enabled || r.ceoReviewer == nil {
		return packet, nil, nil
	}
	prompt := renderCEODelegationPrompt(packet)
	response, err := r.ceoReviewer.Complete(ctx, model.Request{
		Prompt: prompt,
		Metadata: model.RequestMetadata{
			Kind:      "ceo_delegation",
			AgentName: "ceo",
			AgentRole: "delegation planner",
		},
	})
	if err != nil {
		return jobpacket.Packet{}, nil, fmt.Errorf("complete CEO delegation: %w", err)
	}
	selected, delegation, err := parseCEODelegationResponse(response, packet.Subagents)
	if err != nil {
		return jobpacket.Packet{}, nil, err
	}
	if delegation.PromptBytes == 0 {
		delegation.PromptBytes = len(prompt)
	}
	if r.ceoReviewerRoute.Source != "" {
		delegation.ModelSource = r.ceoReviewerRoute.Source
	}
	if r.ceoReviewerRoute.ProviderName != "" {
		delegation.ProviderName = r.ceoReviewerRoute.ProviderName
	}
	packet.Subagents = selected
	packet.MaxSubagents = len(selected)
	return packet, delegation, nil
}

func renderCEODelegationPrompt(packet jobpacket.Packet) string {
	var builder strings.Builder
	builder.WriteString("You are the CEO delegation planner. Return JSON only: ")
	builder.WriteString(`{"selected_subagents":["name"],"new_subagents":[{"name":"specialist","role":"narrow role","provider":"configured_provider","allowed_actions":["read_workspace"]}],"assignments":{"name":"specific assignment"},"summary":"short reason"}`)
	builder.WriteString("\n")
	builder.WriteString("task: ")
	builder.WriteString(packet.Task)
	builder.WriteString("\ntask_profile: kind=")
	builder.WriteString(packet.TaskProfile.Kind)
	builder.WriteString(" risk=")
	builder.WriteString(packet.TaskProfile.RiskLevel)
	builder.WriteString("\ncandidate_subagents:\n")
	for _, candidate := range packet.Subagents {
		builder.WriteString("- ")
		builder.WriteString(candidate.Name)
		builder.WriteString(" role=")
		builder.WriteString(candidate.Role)
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("rules: select at least one subagent; select the smallest useful set; for narrow code edits usually select the patch owner only; prefer candidates; add new_subagents only for narrow specialist work; total selected subagents must stay at %d or fewer; assignments are optional and must stay lean\n", packet.MaxSubagents))
	return builder.String()
}

func parseCEODelegationResponse(response model.Response, candidates []jobpacket.Subagent) ([]jobpacket.Subagent, *CEODelegation, error) {
	var payload ceoDelegationPayload
	jsonPayload, ok := model.JSONPayload(response.Text)
	if !ok {
		return nil, nil, fmt.Errorf("parse CEO delegation JSON: %w", ErrInvalidCEODelegation)
	}
	if err := json.Unmarshal([]byte(jsonPayload), &payload); err != nil {
		return nil, nil, fmt.Errorf("parse CEO delegation JSON: %w", err)
	}
	selected, names, newSubagents, assignments, err := selectedDelegationSubagents(payload.SelectedSubagents, payload.NewSubagents, payload.Assignments, candidates, len(candidates))
	if err != nil {
		return nil, nil, err
	}
	return selected, &CEODelegation{
		Source:            "model",
		SelectedSubagents: names,
		NewSubagents:      newSubagents,
		Assignments:       assignments,
		Summary:           strings.TrimSpace(payload.Summary),
		PromptBytes:       response.PromptBytes,
	}, nil
}

func selectedDelegationSubagents(names []string, rawNewSubagents []jobpacket.Subagent, rawAssignments map[string]string, candidates []jobpacket.Subagent, maxSubagents int) ([]jobpacket.Subagent, []string, []jobpacket.Subagent, map[string]string, error) {
	if len(names) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("selected_subagents: %w", ErrInvalidCEODelegation)
	}
	if maxSubagents > 0 && len(names) > maxSubagents {
		return nil, nil, nil, nil, fmt.Errorf("selected_subagents count: %w", ErrInvalidCEODelegation)
	}
	newSubagents, err := jobpacket.NormalizeCustomSubagents(rawNewSubagents)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("new_subagents: %w", err)
	}
	if len(candidates)+len(newSubagents) > jobpacket.MaxDelegatedSubagents {
		return nil, nil, nil, nil, fmt.Errorf("new_subagents count: %w", ErrInvalidCEODelegation)
	}
	candidateByName := map[string]jobpacket.Subagent{}
	for _, candidate := range candidates {
		candidateByName[candidate.Name] = candidate
	}
	for _, subagent := range newSubagents {
		if _, duplicate := candidateByName[subagent.Name]; duplicate {
			return nil, nil, nil, nil, fmt.Errorf("new_subagents duplicate %q: %w", subagent.Name, ErrInvalidCEODelegation)
		}
		candidateByName[subagent.Name] = subagent
	}
	selected := make([]jobpacket.Subagent, 0, len(names))
	selectedNames := make([]string, 0, len(names))
	seen := map[string]struct{}{}
	for index, rawName := range names {
		name := strings.TrimSpace(rawName)
		candidate, ok := candidateByName[name]
		if !ok {
			return nil, nil, nil, nil, fmt.Errorf("selected_subagents[%d] %q: %w", index, rawName, ErrInvalidCEODelegation)
		}
		if _, duplicate := seen[name]; duplicate {
			return nil, nil, nil, nil, fmt.Errorf("selected_subagents[%d] duplicate %q: %w", index, name, ErrInvalidCEODelegation)
		}
		seen[name] = struct{}{}
		selected = append(selected, candidate)
		selectedNames = append(selectedNames, name)
	}
	for index, subagent := range newSubagents {
		if _, selectedNew := seen[subagent.Name]; !selectedNew {
			return nil, nil, nil, nil, fmt.Errorf("new_subagents[%d] %q: %w", index, subagent.Name, ErrInvalidCEODelegation)
		}
	}
	assignments, err := selectedDelegationAssignments(rawAssignments, seen)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for index := range selected {
		if assignment := assignments[selected[index].Name]; assignment != "" {
			selected[index].Assignment = assignment
		}
	}
	return selected, selectedNames, newSubagents, assignments, nil
}

func selectedDelegationAssignments(raw map[string]string, selected map[string]struct{}) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	assignments := map[string]string{}
	for rawName, rawAssignment := range raw {
		name := strings.TrimSpace(rawName)
		if _, ok := selected[name]; !ok {
			return nil, fmt.Errorf("assignments[%s]: %w", rawName, ErrInvalidCEODelegation)
		}
		assignment := strings.TrimSpace(rawAssignment)
		if assignment != "" {
			assignments[name] = assignment
		}
	}
	if len(assignments) == 0 {
		return nil, nil
	}
	return assignments, nil
}
