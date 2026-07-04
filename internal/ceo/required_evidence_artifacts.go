package ceo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/workspace"
)

const requiredEvidenceArtifactsPrefix = "required evidence artifacts:"
const requiredChangedFilesPrefix = "required changed files:"

func writeRequiredEvidenceArtifacts(ctx context.Context, space workspace.Workspace, packet jobpacket.Packet, changedFiles []string, checks []checkrunner.Result) ([]string, error) {
	paths := requiredEvidenceArtifactPaths(packet.Task)
	written := make([]string, 0, len(paths))
	for _, path := range paths {
		existing, err := space.ReadText(ctx, workspace.ReadTextRequest{Path: path})
		if err == nil && strings.TrimSpace(existing.Content) != "" {
			continue
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			// WriteText will surface path, parent directory, and permission problems with clearer context.
		}
		result, err := space.WriteText(ctx, workspace.WriteTextRequest{
			Path:    path,
			Content: renderRequiredEvidenceArtifact(packet, changedFiles, checks),
		})
		if err != nil {
			return nil, fmt.Errorf("write required evidence artifact %s: %w", path, err)
		}
		written = append(written, result.Path)
	}
	return written, nil
}

func requiredEvidenceArtifactPaths(task string) []string {
	return requiredTaskLinePaths(task, requiredEvidenceArtifactsPrefix)
}

func requiredChangedFilePaths(task string) []string {
	return requiredTaskLinePaths(task, requiredChangedFilesPrefix)
}

func requiredTaskLinePaths(task string, prefix string) []string {
	seen := map[string]struct{}{}
	var paths []string
	for _, line := range strings.Split(task, "\n") {
		clean := strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(clean), prefix) {
			continue
		}
		raw := strings.TrimSpace(clean[len(prefix):])
		for _, item := range strings.Split(raw, ",") {
			path := strings.TrimSpace(item)
			path = strings.TrimRight(path, ". ")
			if path == "" || strings.EqualFold(path, "none") {
				continue
			}
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func renderRequiredEvidenceArtifact(packet jobpacket.Packet, changedFiles []string, checks []checkrunner.Result) string {
	var builder strings.Builder
	builder.WriteString("# Task Evidence\n\n")
	builder.WriteString("Task: ")
	builder.WriteString(firstTaskLine(packet.Task))
	builder.WriteString("\n\n")
	builder.WriteString("## Changed Files\n\n")
	if len(changedFiles) == 0 {
		builder.WriteString("- none recorded\n")
	} else {
		for _, path := range changedFiles {
			builder.WriteString("- ")
			builder.WriteString(path)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n## Verification\n\n")
	if len(checks) == 0 {
		builder.WriteString("- no checks configured\n")
	} else {
		for _, check := range checks {
			builder.WriteString("- ")
			builder.WriteString(check.Status)
			builder.WriteString(": `")
			builder.WriteString(strings.Join(check.Argv, " "))
			builder.WriteString("`\n")
		}
	}
	return builder.String()
}

func firstTaskLine(task string) string {
	for _, line := range strings.Split(task, "\n") {
		clean := strings.TrimSpace(line)
		if clean != "" {
			return clean
		}
	}
	return "unspecified"
}
