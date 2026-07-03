package ceo

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

func openWorkspace(root string) (workspace.Workspace, bool, error) {
	if strings.TrimSpace(root) == "" {
		return workspace.Workspace{}, false, nil
	}
	space, err := workspace.New(root)
	if err != nil {
		return workspace.Workspace{}, false, fmt.Errorf("open workspace: %w", err)
	}
	return space, true, nil
}

func writeSubagentEvidenceFiles(ctx context.Context, space workspace.Workspace, results []subagent.Result) ([]string, error) {
	paths := make([]string, 0, len(results))
	for _, result := range results {
		path, err := writeSubagentEvidenceFile(ctx, space, result.AgentName, result)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func writeSubagentEvidenceFile(ctx context.Context, space workspace.Workspace, artifactName string, result subagent.Result) (string, error) {
	writeResult, err := space.WriteText(ctx, workspace.WriteTextRequest{
		Path:    fmt.Sprintf("ceo-artifacts/%s.md", artifactName),
		Content: renderSubagentEvidence(result),
	})
	if err != nil {
		return "", fmt.Errorf("write %s evidence: %w", artifactName, err)
	}
	return writeResult.Path, nil
}

func renderSubagentEvidence(result subagent.Result) string {
	return fmt.Sprintf("# %s\n\nStatus: %s\nContext: %s\n", result.AgentName, result.Status, result.ContextReceived)
}
