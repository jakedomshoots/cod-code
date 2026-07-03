package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeBenchmarkArtifacts(taskDir string, task Task) error {
	for _, artifact := range task.RequiredArtifacts {
		if err := writeRelativeFile(taskDir, artifact, "deterministic benchmark fixture evidence\n"); err != nil {
			return err
		}
	}
	return nil
}

func benchmarkReportPayload(task Task, status worktreeStatusEvidence) (map[string]any, error) {
	report := map[string]any{
		"task_id":        task.ID,
		"verdict":        "pass",
		"changed_files":  append([]string(nil), task.RequiredChangedFiles...),
		"check_results":  commandResultsForTask(task),
		"patch_results":  patchResultsForTask(task),
		"evidence_paths": append([]string(nil), task.RequiredArtifacts...),
	}
	if task.DirtyWorktreeSensitive {
		report["worktree_status"] = status
	}
	for _, field := range task.RequiredReportFields {
		setSyntheticReportField(report, field)
	}
	return report, nil
}

func commandResultsForTask(task Task) []commandResult {
	results := make([]commandResult, 0, len(task.RequiredCommands))
	for _, command := range task.RequiredCommands {
		results = append(results, commandResult{
			Argv:     strings.Fields(command),
			Status:   "pass",
			ExitCode: 0,
			Stdout:   "ok\n",
		})
	}
	return results
}

func patchResultsForTask(task Task) []patchResult {
	diff := "@@\n+" + strings.Join(task.RequiredDiffTerms, "\n+") + "\n"
	results := make([]patchResult, 0, len(task.RequiredChangedFiles))
	for _, path := range task.RequiredChangedFiles {
		results = append(results, patchResult{Path: path, Diff: diff})
	}
	return results
}

func setSyntheticReportField(report map[string]any, field string) {
	parts := strings.Split(strings.TrimSpace(field), ".")
	if len(parts) == 0 || parts[0] == "" {
		return
	}
	if len(parts) == 1 {
		if _, ok := report[parts[0]]; ok {
			return
		}
		report[parts[0]] = "present"
		return
	}
	current := report
	for _, part := range parts[:len(parts)-1] {
		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	current[parts[len(parts)-1]] = "present"
}

func writeRelativeFile(root string, name string, content string) error {
	clean, ok := cleanRelativeArtifactPath(name)
	if !ok {
		return fmt.Errorf("invalid relative path %q", name)
	}
	path := filepath.Join(root, clean)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent for %s: %w", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}
