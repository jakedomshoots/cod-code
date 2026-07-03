package ceo

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/workspace"
)

func buildWorkspaceBrief(ctx context.Context, space workspace.Workspace, hasWorkspace bool, maxFiles int, excludes []string) (*workspace.Brief, error) {
	if !hasWorkspace {
		return nil, nil
	}
	brief, err := space.Brief(ctx, workspace.BriefRequest{
		MaxFiles:     maxFiles,
		ExcludePaths: excludes,
	})
	if err != nil {
		return nil, err
	}
	return &brief, nil
}

func renderWorkspaceBrief(brief *workspace.Brief) string {
	if brief == nil {
		return ""
	}
	paths := make([]string, 0, len(brief.Files))
	for _, file := range brief.Files {
		paths = append(paths, fmt.Sprintf("%s(%dB)", file.Path, file.Bytes))
	}
	summary := fmt.Sprintf("files=%d shown=%d", brief.FileCount, len(brief.Files))
	if len(paths) > 0 {
		summary += " paths=" + strings.Join(paths, ", ")
	}
	if brief.Truncated {
		summary += " truncated=true"
	}
	return summary
}
