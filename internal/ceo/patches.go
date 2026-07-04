package ceo

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

const maxContentPatchExistingReadBytes = 1_000_000

func applyPatchRequests(ctx context.Context, space workspace.Workspace, patches []PatchRequest) ([]workspace.ReplaceTextResult, error) {
	results := make([]workspace.ReplaceTextResult, 0, len(patches))
	for _, patch := range patches {
		if isCreateFilePatch(patch) {
			result, err := space.CreateText(ctx, workspace.CreateTextRequest{
				Path:    patch.Path,
				Content: patch.Content,
			})
			if errors.Is(err, workspace.ErrFileAlreadyExists) {
				result, err = overwriteTextWithContent(ctx, space, patch, true)
			}
			if err != nil {
				return nil, fmt.Errorf("apply content patch %s: %w", patch.Path, err)
			}
			results = append(results, result)
			continue
		}
		result, err := replaceModelPatch(ctx, space, patch, true)
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
			if errors.Is(err, workspace.ErrFileAlreadyExists) {
				result, err = overwriteTextWithContent(ctx, space, patch, false)
			}
			if err != nil {
				return nil, fmt.Errorf("preview content patch %s: %w", patch.Path, err)
			}
			results = append(results, result)
			continue
		}
		result, err := replaceModelPatch(ctx, space, patch, false)
		if err != nil {
			return nil, fmt.Errorf("preview text in %s: %w", patch.Path, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func replaceModelPatch(ctx context.Context, space workspace.Workspace, patch PatchRequest, write bool) (workspace.ReplaceTextResult, error) {
	var (
		result workspace.ReplaceTextResult
		err    error
	)
	request := workspace.ReplaceTextRequest{
		Path: patch.Path,
		Old:  patch.Old,
		New:  patch.New,
	}
	if write {
		result, err = space.ReplaceText(ctx, request)
	} else {
		result, err = space.PreviewReplaceText(ctx, request)
	}
	if !errors.Is(err, workspace.ErrTextNotFound) {
		return result, err
	}
	return replaceLooseWholeFilePatch(ctx, space, patch, write)
}

func overwriteTextWithContent(ctx context.Context, space workspace.Workspace, patch PatchRequest, write bool) (workspace.ReplaceTextResult, error) {
	current, err := space.ReadText(ctx, workspace.ReadTextRequest{
		Path:     patch.Path,
		MaxBytes: maxContentPatchExistingReadBytes,
	})
	if err != nil {
		return workspace.ReplaceTextResult{}, err
	}
	if current.Truncated {
		return workspace.ReplaceTextResult{}, fmt.Errorf("existing file too large for full-file content patch")
	}
	if write {
		if _, err := space.WriteText(ctx, workspace.WriteTextRequest{
			Path:    current.Path,
			Content: patch.Content,
		}); err != nil {
			return workspace.ReplaceTextResult{}, err
		}
	}
	return workspace.ReplaceTextResult{
		Path: current.Path,
		Diff: fmt.Sprintf("--- %s\n+++ %s\n-%s\n+%s\n", current.Path, current.Path, current.Content, patch.Content),
		Old:  current.Content,
		New:  patch.Content,
	}, nil
}

func replaceLooseWholeFilePatch(ctx context.Context, space workspace.Workspace, patch PatchRequest, write bool) (workspace.ReplaceTextResult, error) {
	if strings.TrimSpace(patch.Old) == "" || strings.TrimSpace(patch.New) == "" {
		return workspace.ReplaceTextResult{}, workspace.ErrTextNotFound
	}
	current, err := space.ReadText(ctx, workspace.ReadTextRequest{
		Path:     patch.Path,
		MaxBytes: maxContentPatchExistingReadBytes,
	})
	if err != nil {
		return workspace.ReplaceTextResult{}, err
	}
	if current.Truncated {
		return workspace.ReplaceTextResult{}, fmt.Errorf("existing file too large for loose full-file patch")
	}
	if !looseWholeFilePatchMatch(current.Content, patch.Old) {
		return workspace.ReplaceTextResult{}, workspace.ErrTextNotFound
	}
	if write {
		if _, err := space.WriteText(ctx, workspace.WriteTextRequest{
			Path:    current.Path,
			Content: patch.New,
		}); err != nil {
			return workspace.ReplaceTextResult{}, err
		}
	}
	return workspace.ReplaceTextResult{
		Path: current.Path,
		Diff: fmt.Sprintf("--- %s\n+++ %s\n-%s\n+%s\n", current.Path, current.Path, current.Content, patch.New),
		Old:  current.Content,
		New:  patch.New,
	}, nil
}

func looseWholeFilePatchMatch(current string, old string) bool {
	return normalizeWholeFilePatchText(current) == normalizeWholeFilePatchText(old)
}

func normalizeWholeFilePatchText(text string) string {
	clean := strings.TrimSpace(text)
	clean = strings.TrimSuffix(clean, ";")
	clean = strings.TrimSpace(clean)
	return strings.Join(strings.Fields(clean), " ")
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
