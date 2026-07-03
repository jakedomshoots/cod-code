package ceo

import (
	"encoding/json"
	"fmt"
	"strings"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

const defaultMaxModelPatches = 5

type modelPatchEnvelope struct {
	Patches []PatchRequest `json:"patches"`
}

type modelPatchSelection struct {
	AgentName string
	Patches   []PatchRequest
}

func proposedModelPatches(results []subagent.Result) ([]PatchRequest, error) {
	selection, err := proposedModelPatchSelection(results)
	if err != nil {
		return nil, err
	}
	return selection.Patches, nil
}

func proposedModelPatchSelection(results []subagent.Result) (modelPatchSelection, error) {
	var owner subagent.Result
	ownerFound := false
	for _, result := range results {
		if !hasPatchAction(result.AllowedActions) {
			continue
		}
		if ownerFound {
			return modelPatchSelection{}, fmt.Errorf("multiple patch-capable subagents: %s and %s", owner.AgentName, result.AgentName)
		}
		owner = result
		ownerFound = true
	}
	if !ownerFound || owner.Status != "pass" {
		return modelPatchSelection{}, nil
	}
	patches, err := coderPatchProposals(owner)
	if err != nil {
		return modelPatchSelection{}, err
	}
	return modelPatchSelection{AgentName: owner.AgentName, Patches: patches}, nil
}

func hasPatchAction(actions []string) bool {
	for _, action := range actions {
		if action == string(jobpacket.ActionProposePatch) {
			return true
		}
	}
	return false
}

func coderPatchProposals(result subagent.Result) ([]PatchRequest, error) {
	if len(result.PatchProposals) > 0 {
		patches := patchRequestsFromProposals(result.PatchProposals)
		return patches, validatePatchRequests(patches)
	}
	return parseCoderPatchProposal(result.Summary)
}

func patchRequestsFromProposals(proposals []subagent.PatchProposal) []PatchRequest {
	patches := make([]PatchRequest, 0, len(proposals))
	for _, proposal := range proposals {
		patches = append(patches, PatchRequest{
			Path:    proposal.Path,
			Old:     proposal.Old,
			New:     proposal.New,
			Content: proposal.Content,
		})
	}
	return patches
}

func parseCoderPatchProposal(summary string) ([]PatchRequest, error) {
	cleanSummary := strings.TrimSpace(summary)
	if cleanSummary == "" || !strings.HasPrefix(cleanSummary, "{") {
		return nil, nil
	}

	var envelope modelPatchEnvelope
	if err := json.Unmarshal([]byte(cleanSummary), &envelope); err != nil {
		return nil, fmt.Errorf("parse coder patch proposal: %w", err)
	}
	if err := validatePatchRequests(envelope.Patches); err != nil {
		return nil, err
	}
	return envelope.Patches, nil
}

func validatePatchRequests(patches []PatchRequest) error {
	for index, patch := range patches {
		if strings.TrimSpace(patch.Path) == "" {
			return fmt.Errorf("invalid coder patch proposal at patch %d: path is required", index+1)
		}
		if patch.Content != "" && patch.Old != "" {
			return fmt.Errorf("invalid coder patch proposal at patch %d: choose content or old/new", index+1)
		}
		if isCreateFilePatch(patch) {
			continue
		}
		if patch.Old == "" {
			return fmt.Errorf("invalid coder patch proposal at patch %d: old text is required", index+1)
		}
	}
	return nil
}

func enforceModelPatchLimit(patches []PatchRequest, configuredLimit int) error {
	limit := configuredLimit
	if limit < 1 {
		limit = defaultMaxModelPatches
	}
	if len(patches) > limit {
		return fmt.Errorf("coder proposed %d model patches; max model patches is %d", len(patches), limit)
	}
	return nil
}
