package ceo

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type runtimeArtifactsInput struct {
	Request       JobRequest
	Space         workspace.Workspace
	HasWorkspace  bool
	ArtifactStore runtimeArtifactStore
	Results       []subagent.Result
}

type runtimeArtifacts struct {
	ChangedFiles  []string
	PatchResults  []workspace.ReplaceTextResult
	PatchPreviews []workspace.ReplaceTextResult
	PreviewEvents []PatchPreviewEvent
	PatchAudit    []PatchAuditEntry
	PatchApproval *PatchApproval
}

func buildRuntimeArtifacts(ctx context.Context, input runtimeArtifactsInput) (runtimeArtifacts, error) {
	artifacts := runtimeArtifacts{}
	approvedDigest := strings.TrimSpace(input.Request.ApprovedPreviewDigest)
	approvalPreviews := []workspace.ReplaceTextResult{}
	if len(input.Request.Patches) > 0 {
		if !input.HasWorkspace {
			return runtimeArtifacts{}, fmt.Errorf("workspace is required for patches")
		}
		if input.Request.DryRun || approvedDigest != "" {
			previews, err := previewPatchRequests(ctx, input.Space, input.Request.Patches)
			if err != nil {
				return runtimeArtifacts{}, err
			}
			approvalPreviews = append(approvalPreviews, previews...)
			artifacts.PatchPreviews = append(artifacts.PatchPreviews, previews...)
			artifacts.PreviewEvents = append(artifacts.PreviewEvents, patchPreviewEvents(previews, "cli", "")...)
		}
	}

	if input.Request.PreviewModelPatches || (input.Request.ApplyModelPatches && (input.Request.DryRun || approvedDigest != "")) {
		previews, agentName, err := previewRuntimeModelPatches(ctx, input)
		if err != nil {
			return runtimeArtifacts{}, err
		}
		approvalPreviews = append(approvalPreviews, previews...)
		artifacts.PatchPreviews = append(artifacts.PatchPreviews, previews...)
		artifacts.PreviewEvents = append(artifacts.PreviewEvents, patchPreviewEvents(previews, "model", agentName)...)
	}

	approval, err := patchApprovalForPreviews(approvalPreviews, approvedDigest)
	if err != nil {
		return runtimeArtifacts{}, err
	}
	artifacts.PatchApproval = approval

	if len(input.Request.Patches) > 0 && !input.Request.DryRun {
		applied, err := applyPatchRequests(ctx, input.Space, input.Request.Patches)
		if err != nil {
			return runtimeArtifacts{}, err
		}
		artifacts.PatchResults = append(artifacts.PatchResults, applied...)
		artifacts.PatchAudit = append(artifacts.PatchAudit, patchAuditEntries(applied, "cli", "")...)
		artifacts.ChangedFiles = appendChangedPatchFiles(artifacts.ChangedFiles, applied)
	}

	if input.Request.ApplyModelPatches && !input.Request.DryRun {
		applied, audit, err := applyRuntimeModelPatches(ctx, input)
		if err != nil {
			return runtimeArtifacts{}, err
		}
		artifacts.PatchResults = append(artifacts.PatchResults, applied...)
		artifacts.PatchAudit = append(artifacts.PatchAudit, audit...)
		artifacts.ChangedFiles = appendChangedPatchFiles(artifacts.ChangedFiles, applied)
	}

	if input.ArtifactStore.Enabled && !input.Request.DryRun {
		written, err := writeSubagentEvidenceFiles(ctx, input.ArtifactStore.Space, input.Results)
		if err != nil {
			return runtimeArtifacts{}, err
		}
		artifacts.ChangedFiles = append(artifacts.ChangedFiles, input.ArtifactStore.changedPaths(written)...)
	}
	return artifacts, nil
}

func patchApprovalForPreviews(previews []workspace.ReplaceTextResult, approvedDigest string) (*PatchApproval, error) {
	if len(previews) == 0 {
		if approvedDigest != "" {
			return nil, fmt.Errorf("--approve-preview requires patch previews")
		}
		return nil, nil
	}
	approval := newPatchPreviewApproval(previews)
	if approvedDigest == "" {
		return approval, nil
	}
	if approval.PreviewDigest != approvedDigest {
		return nil, fmt.Errorf("patch approval digest mismatch: approved %s does not match preview %s", approvedDigest, approval.PreviewDigest)
	}
	return newApprovedPatchApproval(previews, approvedDigest), nil
}

func previewRuntimeModelPatches(ctx context.Context, input runtimeArtifactsInput) ([]workspace.ReplaceTextResult, string, error) {
	if !input.HasWorkspace {
		return nil, "", fmt.Errorf("workspace is required for model patch previews")
	}
	modelPatchSelection, err := proposedModelPatchSelection(input.Results)
	if err != nil {
		return nil, "", err
	}
	if err := enforceModelPatchLimit(modelPatchSelection.Patches, input.Request.MaxModelPatches); err != nil {
		return nil, "", err
	}
	previews, err := previewPatchRequests(ctx, input.Space, modelPatchSelection.Patches)
	if err != nil {
		return nil, "", err
	}
	return previews, modelPatchSelection.AgentName, nil
}

func applyRuntimeModelPatches(ctx context.Context, input runtimeArtifactsInput) ([]workspace.ReplaceTextResult, []PatchAuditEntry, error) {
	if !input.HasWorkspace {
		return nil, nil, fmt.Errorf("workspace is required for model patches")
	}
	modelPatchSelection, err := proposedModelPatchSelection(input.Results)
	if err != nil {
		return nil, nil, err
	}
	if err := enforceModelPatchLimit(modelPatchSelection.Patches, input.Request.MaxModelPatches); err != nil {
		return nil, nil, err
	}
	applied, err := applyPatchRequests(ctx, input.Space, modelPatchSelection.Patches)
	if err != nil {
		return nil, nil, err
	}
	return applied, patchAuditEntries(applied, "model", modelPatchSelection.AgentName), nil
}
