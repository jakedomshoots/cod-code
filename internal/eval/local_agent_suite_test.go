package eval

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_RunLocalAgentSuite_records_installed_agents_when_commands_match(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "ceo-packet"), "#!/bin/sh\nprintf '{\"mode\":\"plan_only\"}\\n'\n")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), "#!/bin/sh\nprintf 'CEO_HARNESS_EVAL_OK\\n'\n")
	writeExecutableContent(t, filepath.Join(binDir, "opencode"), "#!/bin/sh\nprintf 'CEO_HARNESS_EVAL_OK\\n'\n")
	writeExecutableContent(t, filepath.Join(binDir, "pi"), "#!/bin/sh\nprintf 'CEO_HARNESS_EVAL_OK\\n'\n")
	t.Setenv("PATH", binDir)
	outputDir := filepath.Join(t.TempDir(), "suite")

	// When
	summary, err := RunLocalAgentSuite(context.Background(), LocalAgentSuiteRequest{
		OutputDir:        outputDir,
		TimeoutSeconds:   5,
		Agents:           []string{"ceo_harness", "codex_cli", "opencode", "pi"},
		CEOHarnessBinary: filepath.Join(binDir, "ceo-packet"),
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentSuite returned error: %v", err)
	}
	if summary.AgentCount != 4 || summary.Passed != 4 || summary.Failed != 0 {
		t.Fatalf("summary = %+v, want 4 passing agents", summary)
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
	requireFile(t, filepath.Join(outputDir, "summary.md"))
	requireFile(t, filepath.Join(outputDir, "iteration-backlog.md"))
	requireFile(t, filepath.Join(outputDir, "codex_cli", "stdout.log"))
}

func Test_RunLocalAgentSuite_times_out_hung_agent_without_blocking_suite(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "pi"), "#!/bin/sh\n/bin/sleep 5\n")
	t.Setenv("PATH", binDir)
	outputDir := filepath.Join(t.TempDir(), "suite")

	// When
	summary, err := RunLocalAgentSuite(context.Background(), LocalAgentSuiteRequest{
		OutputDir:      outputDir,
		TimeoutSeconds: 1,
		Agents:         []string{"pi"},
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentSuite returned error: %v", err)
	}
	if summary.TimedOut != 1 || summary.Failed != 0 {
		t.Fatalf("summary = %+v, want one timeout and no ordinary failure", summary)
	}
	result := summary.Results[0]
	if result.Status != localAgentStatusTimeout {
		t.Fatalf("Status = %q, want timeout", result.Status)
	}
	requireFile(t, filepath.Join(outputDir, "pi", "stderr.log"))
}

func Test_RunLocalAgentSuite_scores_edit_file_task_when_agent_changes_fixture(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "ceo-packet"), "#!/bin/sh\nprintf 'hello new\\n' > app.txt\nprintf '{\"verdict\":\"pass\"}\\n'\n")
	writeExecutableContent(t, filepath.Join(binDir, "codex"), "#!/bin/sh\nprintf 'hello new\\n' > app.txt\nprintf 'done\\n'\n")
	t.Setenv("PATH", binDir)
	outputDir := filepath.Join(t.TempDir(), "suite")

	// When
	summary, err := RunLocalAgentSuite(context.Background(), LocalAgentSuiteRequest{
		OutputDir:        outputDir,
		TimeoutSeconds:   5,
		Agents:           []string{"ceo_harness", "codex_cli"},
		CEOHarnessBinary: filepath.Join(binDir, "ceo-packet"),
		Task:             "edit-file",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentSuite returned error: %v", err)
	}
	if summary.Passed != 2 || summary.Failed != 0 {
		t.Fatalf("summary = %+v, want both edit-file agents passing", summary)
	}
	for _, result := range summary.Results {
		if !result.FileMatched {
			t.Fatalf("%s FileMatched = false, want true", result.ID)
		}
		requireFile(t, result.AppAfterPath)
	}
}

func Test_RunCLI_runs_local_agent_suite_when_flag_is_set(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "ceo-packet"), "#!/bin/sh\nprintf '{\"mode\":\"plan_only\"}\\n'\n")
	t.Setenv("PATH", binDir)
	outputDir := filepath.Join(t.TempDir(), "suite")

	// When
	err := RunCLI(context.Background(), os.Stdout, os.Stderr, []string{
		"--local-agent-suite",
		"--local-agents", "ceo_harness",
		"--ceo-binary", filepath.Join(binDir, "ceo-packet"),
		"--output-dir", outputDir,
		"--timeout-seconds", "5",
	})
	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v", err)
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
}

func Test_BuildLocalAgentIterations_reports_file_mismatch_when_content_is_not_exact(t *testing.T) {
	// Given
	result := LocalAgentResult{
		ID:            "opencode",
		Name:          "OpenCode",
		Status:        localAgentStatusFail,
		OutputMatched: true,
		FileMatched:   false,
		CommandPath:   "evidence/opencode/command.json",
	}

	// When
	iterations := buildLocalAgentIterations(localAgentTaskEditFile, []LocalAgentResult{result})

	// Then
	if len(iterations) != 2 {
		t.Fatalf("iterations length = %d, want 2", len(iterations))
	}
	if !strings.Contains(iterations[1].Finding, "exact expected file content") {
		t.Fatalf("Finding = %q, want file-content mismatch", iterations[1].Finding)
	}
	if !strings.Contains(iterations[1].NextStep, "file diffs") {
		t.Fatalf("NextStep = %q, want file-diff guidance", iterations[1].NextStep)
	}
}

func Test_RunLocalAgentBenchmark_scores_real_task_when_agent_changes_fixture(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), "#!/bin/sh\nmkdir -p .omo/evidence\nprintf 'agent evidence\\n' > .omo/evidence/docs-roadmap-cli-first.md\ncat > docs/ROADMAP.md <<'EOF'\n# Roadmap\n\nCLI-first dogfood and recovery come before GUI work.\nEOF\nprintf 'done\\n'\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, docsRoadmapTaskSpec())
	outputDir := filepath.Join(root, "benchmark")

	// When
	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:        tasksDir,
		OutputDir:       outputDir,
		TimeoutSeconds:  5,
		Agents:          []string{"codex_cli"},
		BenchmarkTaskID: "docs-roadmap-cli-first",
	})
	// Then
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	if summary.AgentCount != 1 || summary.Passed != 1 || summary.Failed != 0 {
		t.Fatalf("summary = %+v, want one passing benchmark agent", summary)
	}
	result := summary.Results[0]
	if result.ScoreVerdict != "pass" || result.PassedChecks != result.TotalChecks {
		t.Fatalf("result = %+v, want passing score", result)
	}
	requireFile(t, result.ReportPath)
	requireFile(t, result.ScorePath)
	requireFile(t, result.DiffPath)
	requireFile(t, filepath.Join(result.WorkspaceDir, ".omo/evidence/docs-roadmap-cli-first.md"))
}

func Test_RunCLI_runs_local_agent_benchmark_when_flag_is_set(t *testing.T) {
	// Given
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "codex"), "#!/bin/sh\nmkdir -p .omo/evidence\nprintf 'agent evidence\\n' > .omo/evidence/docs-roadmap-cli-first.md\nprintf 'CLI-first\\n' >> docs/ROADMAP.md\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, docsRoadmapTaskSpec())
	outputDir := filepath.Join(root, "benchmark")

	// When
	err := RunCLI(context.Background(), os.Stdout, os.Stderr, []string{
		"--local-agent-benchmark",
		"--local-agents", "codex_cli",
		"--local-agent-benchmark-task", "docs-roadmap-cli-first",
		"--tasks", tasksDir,
		"--output-dir", outputDir,
		"--timeout-seconds", "5",
	})
	// Then
	if err != nil {
		t.Fatalf("RunCLI returned error: %v", err)
	}
	requireFile(t, filepath.Join(outputDir, "summary.json"))
	requireFile(t, filepath.Join(outputDir, "codex_cli", "score.json"))
}

func Test_BuildLocalAgentBenchmarkIterations_reports_extra_changed_files(t *testing.T) {
	// Given
	result := LocalAgentBenchmarkResult{
		ID:                "ceo_harness",
		Name:              "CEO Harness",
		Status:            localAgentStatusPass,
		ChangedFilesPath:  "evidence/ceo/changed-files.txt",
		ExtraChangedFiles: []string{"ceo-artifacts/jobs.jsonl"},
	}

	// When
	iterations := buildLocalAgentBenchmarkIterations([]LocalAgentBenchmarkResult{result})

	// Then
	if len(iterations) != 2 {
		t.Fatalf("iterations length = %d, want 2", len(iterations))
	}
	if !strings.Contains(iterations[1].Finding, "extra file") {
		t.Fatalf("Finding = %q, want extra-file footprint finding", iterations[1].Finding)
	}
	if !strings.Contains(iterations[1].NextStep, "isolate") {
		t.Fatalf("NextStep = %q, want artifact isolation next step", iterations[1].NextStep)
	}
}

func docsRoadmapTaskSpec() string {
	return `{
		"id":"docs-roadmap-cli-first",
		"category":"docs",
		"title":"Keep roadmap CLI-first",
		"objective":"Refresh roadmap wording so dogfood and recovery come before GUI work.",
		"required_changed_files":["docs/ROADMAP.md"],
		"required_commands":["go test ./internal/cli -count=1"],
		"required_artifacts":[".omo/evidence/docs-roadmap-cli-first.md"],
		"required_diff_terms":["CLI-first"],
		"required_report_fields":["changed_files"]
	}`
}

func writeExecutableContent(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}
