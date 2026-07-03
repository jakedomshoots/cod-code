package ceo

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"ceoharness/internal/workspace"
)

const ceoPlanArtifactPath = "ceo-artifacts/ceo-plan.md"

type runtimeArtifactStore struct {
	Space         workspace.Workspace
	Root          string
	Enabled       bool
	ReportChanges bool
	ReportPrefix  string
}

func openRuntimeArtifactStore(req JobRequest, space workspace.Workspace, hasWorkspace bool) (runtimeArtifactStore, error) {
	root := strings.TrimSpace(req.ArtifactRoot)
	if !hasWorkspace {
		if root != "" {
			return runtimeArtifactStore{}, fmt.Errorf("artifact root requires workspace")
		}
		return runtimeArtifactStore{}, nil
	}
	if root == "" {
		return runtimeArtifactStore{Space: space, Root: req.WorkspaceDir, Enabled: true, ReportChanges: true}, nil
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return runtimeArtifactStore{}, fmt.Errorf("create artifact root: %w", err)
	}
	artifactSpace, err := workspace.New(root)
	if err != nil {
		return runtimeArtifactStore{}, fmt.Errorf("open artifact root: %w", err)
	}
	prefix, reportChanges, err := artifactReportPrefix(req.WorkspaceDir, root)
	if err != nil {
		return runtimeArtifactStore{}, err
	}
	return runtimeArtifactStore{
		Space:         artifactSpace,
		Root:          root,
		Enabled:       true,
		ReportChanges: reportChanges,
		ReportPrefix:  prefix,
	}, nil
}

func artifactReportPrefix(workspaceRoot string, artifactRoot string) (string, bool, error) {
	workspaceAbs, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", false, fmt.Errorf("resolve workspace root: %w", err)
	}
	artifactAbs, err := filepath.Abs(artifactRoot)
	if err != nil {
		return "", false, fmt.Errorf("resolve artifact root: %w", err)
	}
	relative, err := filepath.Rel(workspaceAbs, artifactAbs)
	if err != nil {
		return "", false, nil
	}
	if relative == "." {
		return "", true, nil
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", false, nil
	}
	return filepath.ToSlash(relative), true, nil
}

func (s runtimeArtifactStore) changedPath(artifactPath string) (string, bool) {
	if !s.ReportChanges {
		return "", false
	}
	clean := filepath.ToSlash(artifactPath)
	if s.ReportPrefix == "" {
		return clean, true
	}
	return path.Join(s.ReportPrefix, clean), true
}

func (s runtimeArtifactStore) changedPaths(artifactPaths []string) []string {
	changed := make([]string, 0, len(artifactPaths))
	for _, artifactPath := range artifactPaths {
		path, ok := s.changedPath(artifactPath)
		if ok {
			changed = append(changed, path)
		}
	}
	return changed
}

func writeExecutionPlanArtifact(ctx context.Context, store runtimeArtifactStore, plan ExecutionPlan) ([]string, error) {
	if !store.Enabled {
		return nil, nil
	}
	writeResult, err := store.Space.WriteText(ctx, workspace.WriteTextRequest{
		Path:    ceoPlanArtifactPath,
		Content: renderExecutionPlan(plan),
	})
	if err != nil {
		return nil, fmt.Errorf("write CEO plan: %w", err)
	}
	path, ok := store.changedPath(writeResult.Path)
	if !ok {
		return nil, nil
	}
	return []string{path}, nil
}

func persistRuntimeReport(ctx context.Context, store runtimeArtifactStore, dryRun bool, report Report) (Report, error) {
	if !store.Enabled || dryRun {
		return report, nil
	}
	return persistWorkspaceReport(ctx, store.Root, report)
}
