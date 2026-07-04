package eval

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func Test_RunLocalAgentBenchmark_retries_timed_out_agent_run(t *testing.T) {
	// Given
	binDir := t.TempDir()
	markerPath := filepath.Join(t.TempDir(), "first-run-marker")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), retryAfterTimeoutAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BENCHMARK_RETRY_MARKER", markerPath)
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  1,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one",
		TimeoutRetries:  1,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.TimeoutRetries != 1 || summary.RunCount != 1 || summary.Passed != 1 || summary.TimedOut != 0 {
		t.Fatalf("summary = %+v, want one final passing result after retry", summary)
	}
	result := summary.Results[0]
	if result.RunAttempt != 2 || result.Status != localAgentStatusPass {
		t.Fatalf("result = %+v, want second run attempt pass", result)
	}
	if len(result.PriorAttempts) != 1 || result.PriorAttempts[0].Status != localAgentStatusTimeout {
		t.Fatalf("prior attempts = %+v, want first timeout preserved", result.PriorAttempts)
	}
	requireFile(t, filepath.Join(root, "benchmark", "codex_cli", "attempt-01", "timing.txt"))
	requireFile(t, filepath.Join(root, "benchmark", "codex_cli", "attempt-02", "score.json"))
}

func Test_RunLocalAgentBenchmark_retries_partial_agent_run(t *testing.T) {
	// Given
	binDir := t.TempDir()
	markerPath := filepath.Join(t.TempDir(), "first-run-marker")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), retryAfterPartialAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BENCHMARK_RETRY_MARKER", markerPath)
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one",
		ResultRetries:   1,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.ResultRetries != 1 || summary.RunCount != 1 || summary.Passed != 1 || summary.Partial != 0 {
		t.Fatalf("summary = %+v, want one final passing result after partial retry", summary)
	}
	result := summary.Results[0]
	if result.RunAttempt != 2 || result.Status != localAgentStatusPass {
		t.Fatalf("result = %+v, want second run attempt pass", result)
	}
	if len(result.PriorAttempts) != 1 || result.PriorAttempts[0].Status != localAgentStatusPartial {
		t.Fatalf("prior attempts = %+v, want first partial preserved", result.PriorAttempts)
	}
	requireFile(t, filepath.Join(root, "benchmark", "codex_cli", "attempt-01", "score.json"))
	requireFile(t, filepath.Join(root, "benchmark", "codex_cli", "attempt-02", "score.json"))
}

func Test_RunLocalAgentBenchmark_uses_agent_timeout_override(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), slowPassingAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:            tasksDir,
		OutputDir:           filepath.Join(root, "benchmark"),
		TimeoutSeconds:      1,
		Agents:              []string{"codex_cli"},
		BenchmarkTaskID:     "docs-one",
		AgentTimeoutSeconds: map[string]int{"codex_cli": 5},
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Passed != 1 || summary.TimedOut != 0 || summary.AgentTimeouts["codex_cli"] != 5 {
		t.Fatalf("summary = %+v, want agent timeout override to allow pass", summary)
	}
}

func Test_RunLocalAgentBenchmark_records_agent_model_override(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), slowPassingAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, `{
		"id":"docs-one",
		"category":"docs",
		"title":"Update one",
		"objective":"Refresh docs one.",
		"required_changed_files":["docs/ONE.md"],
		"required_artifacts":[".omo/evidence/docs-one.md"],
		"required_diff_terms":["one-term"],
		"required_report_fields":["changed_files"]
	}`)

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one",
		AgentModels:     map[string]string{"codex_cli": "openai/gpt-5.4-mini"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.AgentModels["codex_cli"] != "openai/gpt-5.4-mini" {
		t.Fatalf("agent models = %+v, want codex_cli override", summary.AgentModels)
	}
	if !slices.Contains(summary.Results[0].Command, "--model") || !slices.Contains(summary.Results[0].Command, "openai/gpt-5.4-mini") {
		t.Fatalf("command = %+v, want model override", summary.Results[0].Command)
	}
}

func retryAfterTimeoutAgentScript() string {
	return `#!/bin/sh
set -eu
marker="${BENCHMARK_RETRY_MARKER:?}"
if [ ! -f "$marker" ]; then
  : > "$marker"
  sleep 3
fi
printf '# Benchmark Fixture\n\none-term\n' > docs/ONE.md
mkdir -p .omo/evidence
printf 'agent evidence\n' > .omo/evidence/docs-one.md
printf 'done\n'
`
}

func retryAfterPartialAgentScript() string {
	return `#!/bin/sh
set -eu
marker="${BENCHMARK_RETRY_MARKER:?}"
mkdir -p .omo/evidence
if [ ! -f "$marker" ]; then
  : > "$marker"
  printf '# Benchmark Fixture\n\nmissing required term\n' > docs/ONE.md
  printf 'agent evidence\n' > .omo/evidence/docs-one.md
  printf 'partial\n'
  exit 0
fi
printf '# Benchmark Fixture\n\none-term\n' > docs/ONE.md
printf 'agent evidence\n' > .omo/evidence/docs-one.md
printf 'done\n'
`
}

func slowPassingAgentScript() string {
	return `#!/bin/sh
set -eu
sleep 2
printf '# Benchmark Fixture\n\none-term\n' > docs/ONE.md
mkdir -p .omo/evidence
printf 'agent evidence\n' > .omo/evidence/docs-one.md
printf 'done\n'
`
}
