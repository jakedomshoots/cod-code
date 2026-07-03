package ceo

import (
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_proposedModelPatches_reads_single_patch_capable_specialist(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "ux_coder",
			Status:         "pass",
			AllowedActions: []string{string(jobpacket.ActionProposePatch)},
			Summary:        "patch ready",
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt", Old: "old", New: "new"},
			},
		},
	}

	// When
	patches, err := proposedModelPatches(results)

	// Then
	if err != nil {
		t.Fatalf("proposedModelPatches returned error: %v", err)
	}
	if len(patches) != 1 || patches[0].Path != "app.txt" {
		t.Fatalf("patches = %+v, want specialist app patch", patches)
	}
}

func Test_proposedModelPatches_rejects_multiple_patch_capable_subagents(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "coder",
			Status:         "pass",
			AllowedActions: []string{string(jobpacket.ActionProposePatch)},
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt", Old: "old", New: "new"},
			},
		},
		{
			AgentName:      "ux_coder",
			Status:         "pass",
			AllowedActions: []string{string(jobpacket.ActionProposePatch)},
			PatchProposals: []subagent.PatchProposal{
				{Path: "ux.txt", Old: "old", New: "new"},
			},
		},
	}

	// When
	_, err := proposedModelPatches(results)

	// Then
	if err == nil || !strings.Contains(err.Error(), "multiple patch-capable subagents") {
		t.Fatalf("error = %v, want multiple patch-capable subagents", err)
	}
}
