package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test_RunLocalAgentBenchmark_runs_selected_tasks_for_each_repeat(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), multiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, multiTaskSpec())

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one,docs-two",
		RepeatCount:     2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.TaskCount != 2 || summary.RepeatCount != 2 || summary.RunCount != 4 {
		t.Fatalf("summary = %+v, want 2 tasks repeated twice", summary)
	}
	if summary.Passed != 4 || len(summary.Results) != 4 {
		t.Fatalf("summary = %+v, want four passing results", summary)
	}
	wantAttempts := map[string]int{"docs-one/1": 0, "docs-one/2": 0, "docs-two/1": 0, "docs-two/2": 0}
	for _, result := range summary.Results {
		wantAttempts[fmt.Sprintf("%s/%d", result.TaskID, result.Attempt)]++
		requireFile(t, result.ScorePath)
	}
	for key, count := range wantAttempts {
		if count != 1 {
			t.Fatalf("attempt %s count = %d, want 1", key, count)
		}
	}
}

func Test_RunLocalAgentBenchmark_runs_jobs_in_parallel_when_concurrency_is_set(t *testing.T) {
	// Given
	binDir := t.TempDir()
	barrierDir := filepath.Join(t.TempDir(), "barrier")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), concurrentMultiTaskAgentScript())
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BENCHMARK_CONCURRENCY_DIR", barrierDir)
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, multiTaskSpec())

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       filepath.Join(root, "benchmark"),
		TimeoutSeconds:  10,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-one,docs-two",
		Concurrency:     2,
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.Concurrency != 2 || summary.Passed != 2 || len(summary.Results) != 2 {
		t.Fatalf("summary = %+v, want two passing concurrent results", summary)
	}
	if summary.Results[0].TaskID != "docs-one" || summary.Results[1].TaskID != "docs-two" {
		t.Fatalf("result order = %s, %s; want planned task order", summary.Results[0].TaskID, summary.Results[1].TaskID)
	}
}

func concurrentMultiTaskAgentScript() string {
	return `#!/bin/sh
set -eu
barrier="${BENCHMARK_CONCURRENCY_DIR:?}"
mkdir -p "$barrier"
: > "$barrier/started-$$"
deadline=$(( $(date +%s) + 5 ))
while [ "$(find "$barrier" -name 'started-*' -type f | wc -l | tr -d ' ')" -lt 2 ]; do
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "concurrency barrier timed out" >&2
    exit 2
  fi
  sleep 0.1
done
if [ -f docs/ONE.md ]; then
  printf '# Benchmark Fixture\n\none-term\n' > docs/ONE.md
  mkdir -p .omo/evidence
  printf 'agent evidence\n' > .omo/evidence/docs-one.md
fi
if [ -f docs/TWO.md ]; then
  printf '# Benchmark Fixture\n\ntwo-term\n' > docs/TWO.md
  mkdir -p .omo/evidence
  printf 'agent evidence\n' > .omo/evidence/docs-two.md
fi
printf 'done\n'
`
}
