package ceo

import (
	"context"
	"fmt"

	"ceoharness/internal/workspace"
)

type PatchAuditEntry struct {
	Path      string `json:"path"`
	Source    string `json:"source"`
	AgentName string `json:"agent_name,omitempty"`
}

type PatchPreviewEvent struct {
	Path      string
	Source    string
	AgentName string
}

func applyPatchRequests(ctx context.Context, space workspace.Workspace, patches []PatchRequest) ([]workspace.ReplaceTextResult, error) {
	results := make([]workspace.ReplaceTextResult, 0, len(patches))
	for _, patch := range patches {
		if isCreateFilePatch(patch) {
			result, err := space.CreateText(ctx, workspace.CreateTextRequest{
				Path:    patch.Path,
				Content: patch.Content,
			})
			if err != nil {
				return nil, fmt.Errorf("create file %s: %w", patch.Path, err)
			}
			results = append(results, result)
			continue
		}
		result, err := space.ReplaceText(ctx, workspace.ReplaceTextRequest{
			Path: patch.Path,
			Old:  patch.Old,
			New:  patch.New,
		})
		if err != nil {
			return nil, fmt.Errorf("replace text in %s: %w", patch.Path, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func previewPatchRequests(ctx context.Context, space workspace.Workspace, patches []PatchRequest) ([]workspace.ReplaceTextResult, error) {
	results := make([]workspace.ReplaceTextResult, 0, len(patches))
	for _, patch := range patches {
		if isCreateFilePatch(patch) {
			result, err := space.PreviewCreateText(ctx, workspace.CreateTextRequest{
				Path:    patch.Path,
				Content: patch.Content,
			})
			if err != nil {
				return nil, fmt.Errorf("preview create file %s: %w", patch.Path, err)
			}
			results = append(results, result)
			continue
		}
		result, err := space.PreviewReplaceText(ctx, workspace.ReplaceTextRequest{
			Path: patch.Path,
			Old:  patch.Old,
			New:  patch.New,
		})
		if err != nil {
			return nil, fmt.Errorf("preview text in %s: %w", patch.Path, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func patchAuditEntries(results []workspace.ReplaceTextResult, source string, agentName string) []PatchAuditEntry {
	entries := make([]PatchAuditEntry, 0, len(results))
	for _, result := range results {
		entries = append(entries, PatchAuditEntry{
			Path:      result.Path,
			Source:    source,
			AgentName: agentName,
		})
	}
	return entries
}

func patchPreviewEvents(results []workspace.ReplaceTextResult, source string, agentName string) []PatchPreviewEvent {
	events := make([]PatchPreviewEvent, 0, len(results))
	for _, result := range results {
		events = append(events, PatchPreviewEvent{
			Path:      result.Path,
			Source:    source,
			AgentName: agentName,
		})
	}
	return events
}

func appendChangedPatchFiles(changedFiles []string, patchResults []workspace.ReplaceTextResult) []string {
	for _, result := range patchResults {
		changedFiles = append(changedFiles, result.Path)
	}
	return changedFiles
}

func isCreateFilePatch(patch PatchRequest) bool {
	return patch.Content != "" && patch.Old == ""
}
