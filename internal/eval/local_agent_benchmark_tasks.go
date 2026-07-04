package eval

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

var localAgentBenchmarkTaskSuites = map[string][]string{
	"market-parity-core": {
		"bugfix-cli-timeout",
		"docs-roadmap-cli-first",
		"refactor-model-selection-split",
		"test-repair-require-checks",
		"provider-config-openai-compatible",
		"safety-policy-observe-no-write",
		"safety-policy-path-escape",
		"recovery-resume-retry",
		"safety-policy-rollback-report",
		"report-quality-evidence-summary",
	},
	"production-core": {
		"bugfix-cli-timeout",
		"bugfix-history-latest",
		"bugfix-provider-health-rollup",
		"bugfix-report-context-truncation",
		"refactor-model-selection-split",
		"refactor-text-report-sections",
		"refactor-check-fix-prompt",
		"refactor-workspace-brief-excludes",
		"multi-file-provider-fallback-reporting",
		"test-repair-require-checks",
		"test-repair-provider-policy",
		"test-repair-run-events",
		"test-repair-smoke-script",
		"docs-verification-record",
		"docs-product-status-weak-spots",
		"docs-roadmap-cli-first",
		"provider-config-openai-compatible",
		"provider-config-budget-metadata",
		"provider-config-health-policy",
		"safety-policy-observe-no-write",
		"safety-policy-approved-digest",
		"safety-policy-path-escape",
		"recovery-resume-retry",
		"safety-policy-rollback-report",
		"multi-file-operator-safety-flow",
		"multi-file-release-readiness-publish-boundary",
		"multi-file-lean-context-autonomy",
		"multi-file-secret-safe-provider-proof",
		"multi-file-finalizer-check-fix",
		"report-quality-evidence-summary",
	},
	"cross-language-core": {
		"cross-language-js-state-reducer",
		"cross-language-python-retry-policy",
	},
}

func localAgentBenchmarkTasks(ctx context.Context, req LocalAgentBenchmarkRequest) ([]Task, error) {
	tasks, err := LoadTasks(ctx, req.TasksDir)
	if err != nil {
		return nil, err
	}
	taskIDs := requestedLocalAgentBenchmarkTaskIDs(req.BenchmarkTaskID)
	if len(taskIDs) == 1 && taskIDs[0] == "all" {
		return tasks, nil
	}
	selected := make([]Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, err := FindTask(tasks, taskID)
		if err != nil {
			return nil, err
		}
		selected = append(selected, task)
	}
	return selected, nil
}

func requestedLocalAgentBenchmarkTaskIDs(raw string) []string {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return []string{defaultLocalAgentBenchmarkID}
	}
	parts := strings.Split(clean, ",")
	taskIDs := make([]string, 0, len(parts))
	for _, part := range parts {
		taskID := strings.TrimSpace(part)
		if taskID != "" {
			taskIDs = append(taskIDs, expandLocalAgentBenchmarkTaskID(taskID)...)
		}
	}
	if len(taskIDs) == 0 {
		return []string{defaultLocalAgentBenchmarkID}
	}
	return taskIDs
}

func expandLocalAgentBenchmarkTaskID(taskID string) []string {
	if suite, ok := localAgentBenchmarkTaskSuites[taskID]; ok {
		return append([]string(nil), suite...)
	}
	return []string{taskID}
}

func normalizeLocalAgentBenchmarkRepeat(repeatCount int) int {
	if repeatCount < 1 {
		return 1
	}
	return repeatCount
}

func localAgentBenchmarkTaskIDs(tasks []Task) []string {
	taskIDs := make([]string, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}
	return taskIDs
}

func localAgentBenchmarkResultDir(outputDir string, task Task, spec localAgentSpec, attempt int, multiRun bool) string {
	if !multiRun {
		return filepath.Join(outputDir, spec.id)
	}
	return filepath.Join(outputDir, task.ID, fmt.Sprintf("run-%02d", attempt), spec.id)
}
