package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_LocalAgentBenchmarkTasks_market_parity_core_covers_required_categories(t *testing.T) {
	// Given
	want := []string{
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
	}

	// When
	got := requestedLocalAgentBenchmarkTaskIDs("market-parity-core")

	// Then
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("market-parity-core = %v, want %v", got, want)
	}
}

func Test_WriteLocalAgentComparisonReport_includes_artifact_derived_details(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "comparison-report.md")
	summary := LocalAgentBenchmarkSummary{
		Mode:               "local_agent_benchmark",
		TaskCount:          1,
		Concurrency:        2,
		RunCount:           2,
		Passed:             1,
		Partial:            1,
		IncompleteEvidence: 1,
		Results: []LocalAgentBenchmarkResult{
			{
				ID:             "ceo_harness",
				Name:           "CEO Harness",
				TaskID:         "docs-roadmap-cli-first",
				Attempt:        1,
				Status:         localAgentStatusPass,
				EvidenceStatus: localAgentEvidenceComplete,
				PassedChecks:   5,
				TotalChecks:    5,
				ChangedFiles:   []string{"docs/ROADMAP.md"},
				DurationMS:     42,
				TimingPath:     "run-01/timing.txt",
				ScorePath:      "run-01/score.json",
				ReportPath:     "run-01/report.json",
			},
			{
				ID:             "ceo_harness",
				Name:           "CEO Harness",
				TaskID:         "docs-roadmap-cli-first",
				Attempt:        2,
				Status:         localAgentStatusPartial,
				EvidenceStatus: localAgentEvidenceIncomplete,
				PassedChecks:   4,
				TotalChecks:    5,
				FailedScoreChecks: []CheckResult{
					{Name: "artifact:.omo/evidence/docs-roadmap-cli-first.md", Status: "fail"},
				},
				ChangedFiles:      []string{"docs/ROADMAP.md"},
				ExtraChangedFiles: []string{"tmp.txt"},
				DurationMS:        55,
				TimingPath:        "run-02/timing.txt",
				ScorePath:         "run-02/score.json",
				ReportPath:        "run-02/report.json",
			},
		},
	}

	// When
	if err := writeLocalAgentComparisonReport(path, summary); err != nil {
		t.Fatalf("writeLocalAgentComparisonReport returned error: %v", err)
	}

	// Then
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"## Aggregate Counts",
		"## Readiness Decision",
		"Overall comparison: blocked",
		"CEO Harness result: needs attention",
		"Concurrency: 2",
		"Passed: 1",
		"Partial: 1",
		"## Per-Agent Status",
		"CEO Harness",
		"## Run Artifact Detail",
		"artifact:.omo/evidence/docs-roadmap-cli-first.md",
		"docs/ROADMAP.md",
		"tmp.txt",
		"run-02/timing.txt",
		"run-02/report.json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("comparison report missing %q:\n%s", want, text)
		}
	}
}

func Test_WriteLocalAgentComparisonReport_separatesCleanCEOFromExternalBlockers(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "comparison-report.md")
	summary := LocalAgentBenchmarkSummary{
		Mode:         "local_agent_benchmark",
		TaskCount:    1,
		Concurrency:  2,
		RunCount:     2,
		Passed:       1,
		SetupBlocked: 1,
		Results: []LocalAgentBenchmarkResult{
			{
				ID:             "ceo_harness",
				Name:           "CEO Harness",
				TaskID:         "docs-roadmap-cli-first",
				Attempt:        1,
				Status:         localAgentStatusPass,
				EvidenceStatus: localAgentEvidenceComplete,
				PassedChecks:   5,
				TotalChecks:    5,
			},
			{
				ID:             "opencode",
				Name:           "OpenCode",
				TaskID:         "docs-roadmap-cli-first",
				Attempt:        1,
				Status:         localAgentStatusSetupBlocked,
				EvidenceStatus: localAgentEvidenceComplete,
				PassedChecks:   0,
				TotalChecks:    5,
			},
		},
	}

	// When
	if err := writeLocalAgentComparisonReport(path, summary); err != nil {
		t.Fatalf("writeLocalAgentComparisonReport returned error: %v", err)
	}

	// Then
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	text := string(content)
	for _, want := range []string{
		"Overall comparison: blocked",
		"CEO Harness result: clean",
		"External blockers: OpenCode partial=0 fail=0 timeout=0 setup_blocked=1 incomplete=0",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("comparison report missing %q:\n%s", want, text)
		}
	}
}

func Test_WriteLocalAgentBenchmarkMarkdown_includes_concurrency(t *testing.T) {
	// Given
	path := filepath.Join(t.TempDir(), "summary.md")
	summary := LocalAgentBenchmarkSummary{
		Mode:        localAgentBenchmarkMode,
		TaskCount:   2,
		RepeatCount: 1,
		Concurrency: 4,
		RunCount:    2,
		Passed:      2,
		Results: []LocalAgentBenchmarkResult{
			{
				Name:           "CEO Harness",
				TaskID:         "docs-one",
				Attempt:        1,
				Status:         localAgentStatusPass,
				PassedChecks:   5,
				TotalChecks:    5,
				ChangedFiles:   []string{"docs/ONE.md"},
				EvidenceStatus: localAgentEvidenceComplete,
				ScorePath:      "docs-one/score.json",
			},
		},
	}

	// When
	if err := writeLocalAgentBenchmarkMarkdown(path, summary); err != nil {
		t.Fatalf("writeLocalAgentBenchmarkMarkdown returned error: %v", err)
	}

	// Then
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if !strings.Contains(string(content), "Concurrency: 4") {
		t.Fatalf("summary missing concurrency:\n%s", string(content))
	}
}
