package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func Test_ParseArgs_start_verb_matches_start_flag(t *testing.T) {
	// Given
	flagOpts, flagErr := parseArgs([]string{"--start", "/tmp/workspace"})

	// When
	verbOpts, verbErr := parseArgs([]string{"start", "/tmp/workspace"})

	// Then
	if flagErr != nil {
		t.Fatalf("parse --start: %v", flagErr)
	}
	if verbErr != nil {
		t.Fatalf("parse start verb: %v", verbErr)
	}
	if verbOpts.startDir != flagOpts.startDir {
		t.Fatalf("startDir = %q, want %q", verbOpts.startDir, flagOpts.startDir)
	}
}

func Test_ParseArgs_sets_command_surfaces_when_verbs_are_supplied(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want func(t *testing.T, opts options)
	}{
		{
			name: "inbox",
			args: []string{"inbox", "--workspace", "/tmp/workspace"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showInbox || opts.workspaceDir != "/tmp/workspace" {
					t.Fatalf("opts = %+v, want inbox for workspace", opts)
				}
			},
		},
		{
			name: "doctor",
			args: []string{"doctor"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showDoctor {
					t.Fatalf("showDoctor = false, want true")
				}
			},
		},
		{
			name: "config check",
			args: []string{"config", "check"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showConfigCheck {
					t.Fatalf("showConfigCheck = false, want true")
				}
			},
		},
		{
			name: "config doctor",
			args: []string{"config", "doctor"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showConfigCheck {
					t.Fatalf("showConfigCheck = false, want true")
				}
			},
		},
		{
			name: "review",
			args: []string{"review", "--limit", "5"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showReviewQueue || !opts.reviewDetails || opts.historyLimit != 5 {
					t.Fatalf("opts = %+v, want review queue details limit 5", opts)
				}
			},
		},
		{
			name: "context",
			args: []string{"context", "--workspace", "/tmp/workspace", "latest", "--format", "text"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if opts.contextTraceID != "latest" || opts.workspaceDir != "/tmp/workspace" || opts.reportFormat != reportFormatText {
					t.Fatalf("opts = %+v, want context trace latest text for workspace", opts)
				}
			},
		},
		{
			name: "status",
			args: []string{"status", "--workspace", "/tmp/workspace"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if !opts.showHistory || !opts.historySummaryOnly || opts.workspaceDir != "/tmp/workspace" {
					t.Fatalf("opts = %+v, want summary history status", opts)
				}
			},
		},
		{
			name: "resume",
			args: []string{"resume", "job-000001", "--answer", "Use internal/cli", "--workspace", "/tmp/workspace"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if opts.resumeJobID != "job-000001" || len(opts.resumeAnswers) != 1 || opts.resumeAnswers[0] != "Use internal/cli" || opts.workspaceDir != "/tmp/workspace" {
					t.Fatalf("opts = %+v, want resume job with answer", opts)
				}
			},
		},
		{
			name: "retry",
			args: []string{"retry", "latest", "--workspace", "/tmp/workspace"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if opts.rerunJobID != "latest" || opts.workspaceDir != "/tmp/workspace" {
					t.Fatalf("opts = %+v, want latest rerun retry", opts)
				}
			},
		},
		{
			name: "rollback",
			args: []string{"rollback", "reports/job-000001.json", "--workspace", "/tmp/workspace"},
			want: func(t *testing.T, opts options) {
				t.Helper()
				if opts.rollbackReportPath != "reports/job-000001.json" || opts.workspaceDir != "/tmp/workspace" {
					t.Fatalf("opts = %+v, want rollback report path", opts)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			opts, err := parseArgs(tt.args)

			// Then
			if err != nil {
				t.Fatalf("parseArgs: %v", err)
			}
			tt.want(t, opts)
		})
	}
}

func Test_ParseArgs_run_verb_preserves_task_text(t *testing.T) {
	// Given
	args := []string{"run", "--workspace", "/tmp/workspace", "Fix", "retry", "bug"}

	// When
	opts, err := parseArgs(args)

	// Then
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if opts.task != "Fix retry bug" {
		t.Fatalf("task = %q, want joined task text", opts.task)
	}
	if opts.workspaceDir != "/tmp/workspace" {
		t.Fatalf("workspaceDir = %q, want workspace", opts.workspaceDir)
	}
}

func Test_ParseArgs_preserves_single_token_legacy_task_text(t *testing.T) {
	// Given
	args := []string{"refactor"}

	// When
	opts, err := parseArgs(args)

	// Then
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if opts.task != "refactor" {
		t.Fatalf("task = %q, want legacy task text", opts.task)
	}
}

func Test_Run_preserves_single_token_legacy_task_text(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"refactor"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobPacket struct {
			Task string `json:"task"`
		} `json:"job_packet"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.JobPacket.Task != "refactor" {
		t.Fatalf("task = %q, want single-token legacy task text", body.JobPacket.Task)
	}
}

func Test_Run_eval_verb_lists_tasks(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"eval", "--list", "--tasks", "../../evals/tasks"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "bugfix-cli-timeout") {
		t.Fatalf("output = %q, want eval task ids", out.String())
	}
}

func Test_Run_gauntlet_verb_lists_benchmark_helpful_error_when_agent_is_missing(t *testing.T) {
	// Given
	var out bytes.Buffer
	outputDir := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{
		"gauntlet",
		"--agents", "missing_agent",
		"--output-dir", outputDir,
		"--tasks", "../../evals/tasks",
		"--timeout-seconds", "1",
	})

	// Then
	if err == nil {
		t.Fatal("expected unknown local agent error")
	}
	if !strings.Contains(err.Error(), "unknown local agent") {
		t.Fatalf("error = %q, want unknown local agent guidance", err.Error())
	}
}

func Test_GauntletEvalArgs_defaults_to_market_parity_core_suite(t *testing.T) {
	// Given
	args := []string{"--agents", "ceo_harness", "--output-dir", "/tmp/gauntlet"}

	// When
	normalized := gauntletEvalArgs(args)

	// Then
	body := strings.Join(normalized, " ")
	for _, want := range []string{
		"--local-agent-benchmark",
		"--local-agent-benchmark-task market-parity-core",
		"--local-agents ceo_harness",
		"--output-dir /tmp/gauntlet",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("gauntlet args = %q, want %q", body, want)
		}
	}
}

func Test_GauntletEvalArgs_accepts_production_suite_alias(t *testing.T) {
	// Given
	args := []string{"--suite", "production-core", "--agents", "ceo_harness"}

	// When
	normalized := gauntletEvalArgs(args)

	// Then
	body := strings.Join(normalized, " ")
	for _, want := range []string{
		"--local-agent-benchmark-task production-core",
		"--local-agents ceo_harness",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("gauntlet args = %q, want %q", body, want)
		}
	}
}

func Test_ParseArgs_unknown_config_subcommand_returns_actionable_guidance(t *testing.T) {
	// Given
	args := []string{"config", "nope"}

	// When
	_, err := parseArgs(args)

	// Then
	if err == nil {
		t.Fatal("expected unknown config command error")
	}
	if !strings.Contains(err.Error(), "unknown config command") || !strings.Contains(err.Error(), "--help") {
		t.Fatalf("error = %q, want unknown config command guidance", err.Error())
	}
}
