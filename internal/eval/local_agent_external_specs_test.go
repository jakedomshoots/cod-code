package eval

import (
	"slices"
	"testing"
)

// Test_BuildLocalAgentSpec_externalAgents_emits_noninteractive_flags verifies the
// readiness task contract for each newly supported external local agent: the
// expected binary and a focused set of noninteractive/approval flags must
// appear in args, while each agent's binary must be one of the safe paths.
func Test_BuildLocalAgentSpec_externalAgents_emits_noninteractive_flags(t *testing.T) {
	readiness, err := localAgentTask(localAgentTaskReadiness)
	if err != nil {
		t.Fatalf("localAgentTask readiness: %v", err)
	}
	cases := []struct {
		agent      string
		wantBinary string
		mustHave   []string
		mustReject []string
		wantOutput string
	}{
		{
			agent:      "claude_code",
			wantBinary: "claude",
			mustHave:   []string{"--safe-mode", "--no-session-persistence", "--permission-mode", "plan", "--print"},
			wantOutput: localAgentMarker,
		},
		{
			agent:      "aider",
			wantBinary: "aider",
			mustHave:   []string{"--no-git", "--no-gitignore", "--no-auto-commits", "--yes-always", "--message"},
			wantOutput: localAgentMarker,
		},
		{
			agent:      "goose",
			wantBinary: "goose",
			mustHave:   []string{"run", "--no-session", "--quiet", "--text"},
			wantOutput: localAgentMarker,
		},
		{
			agent:      "oh_my_pi",
			wantBinary: "omp",
			mustHave:   []string{"--no-session", "--no-tools", "--no-rules", "--no-skills", "--max-time", "--print"},
			wantOutput: localAgentMarker,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.agent, func(t *testing.T) {
			spec, err := buildLocalAgentSpec(tc.agent, "./bin/ceo-packet", readiness)
			if err != nil {
				t.Fatalf("buildLocalAgentSpec(%s) error: %v", tc.agent, err)
			}
			if spec.binary != tc.wantBinary {
				t.Fatalf("binary = %q, want %q", spec.binary, tc.wantBinary)
			}
			if spec.id != tc.agent {
				t.Fatalf("id = %q, want %q", spec.id, tc.agent)
			}
			if spec.expectedOutput != tc.wantOutput {
				t.Fatalf("expectedOutput = %q, want %q", spec.expectedOutput, tc.wantOutput)
			}
			for _, want := range tc.mustHave {
				if !slices.Contains(spec.args, want) {
					t.Fatalf("args = %+v, missing %q", spec.args, want)
				}
			}
			for _, reject := range tc.mustReject {
				if slices.Contains(spec.args, reject) {
					t.Fatalf("args = %+v, must not contain %q", spec.args, reject)
				}
			}
			if spec.args[len(spec.args)-1] != readiness.prompt {
				t.Fatalf("last arg = %q, want readiness prompt", spec.args[len(spec.args)-1])
			}
		})
	}
}

// Test_BuildLocalAgentSpec_externalAgents_editFileTaskRaisesApproval verifies
// the edit-file task branch: expected output empties, approval/noninteractive
// flags tighten, and a sentinel "no .git mutation" flag stays put for Aider.
func Test_BuildLocalAgentSpec_externalAgents_editFileTaskRaisesApproval(t *testing.T) {
	edit, err := localAgentTask(localAgentTaskEditFile)
	if err != nil {
		t.Fatalf("localAgentTask edit-file: %v", err)
	}
	cases := []struct {
		agent      string
		wantBinary string
		mustHave   []string
	}{
		{
			agent:      "claude_code",
			wantBinary: "claude",
			mustHave:   []string{"--permission-mode", "bypassPermissions", "--output-format", "json", "--print"},
		},
		{
			agent:      "aider",
			wantBinary: "aider",
			mustHave:   []string{"--no-git", "--no-gitignore", "--no-auto-commits", "--yes-always", "app.txt"},
		},
		{
			agent:      "goose",
			wantBinary: "goose",
			mustHave:   []string{"run", "--no-session", "--quiet", "--text"},
		},
		{
			agent:      "oh_my_pi",
			wantBinary: "omp",
			mustHave:   []string{"--auto-approve", "--approval-mode", "yolo", "--max-time", "240", "--print"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.agent, func(t *testing.T) {
			spec, err := buildLocalAgentSpec(tc.agent, "./bin/ceo-packet", edit)
			if err != nil {
				t.Fatalf("buildLocalAgentSpec(%s) error: %v", tc.agent, err)
			}
			if spec.binary != tc.wantBinary {
				t.Fatalf("binary = %q, want %q", spec.binary, tc.wantBinary)
			}
			if spec.expectedOutput != "" {
				t.Fatalf("expectedOutput = %q, want empty for edit-file task", spec.expectedOutput)
			}
			if spec.expectedFile != edit.expectedFile {
				t.Fatalf("expectedFile = %q, want %q", spec.expectedFile, edit.expectedFile)
			}
			for _, want := range tc.mustHave {
				if !slices.Contains(spec.args, want) {
					t.Fatalf("args = %+v, missing %q", spec.args, want)
				}
			}
		})
	}
}

// Test_BuildLocalAgentBenchmarkSpec_externalAgents_appliesModelOverride verifies
// that per-agent --model overrides forward into the spec args for each of the
// four new external agents.
func Test_BuildLocalAgentBenchmarkSpec_externalAgents_appliesModelOverride(t *testing.T) {
	task := Task{
		ID:                   "docs-roadmap-cli-first",
		Title:                "Keep roadmap CLI-first",
		Objective:            "Refresh roadmap wording.",
		RequiredChangedFiles: []string{"docs/ROADMAP.md"},
		RequiredDiffTerms:    []string{"CLI-first"},
	}
	models := map[string]string{
		"claude_code": "anthropic/claude-3.5-sonnet",
		"aider":       "openai/gpt-5.4-mini",
		"goose":       "anthropic/claude-3.5-sonnet",
		"oh_my_pi":    "openai/gpt-5-mini",
	}
	for _, agent := range []string{"claude_code", "aider", "goose", "oh_my_pi"} {
		spec, err := buildLocalAgentBenchmarkSpec(agent, LocalAgentBenchmarkRequest{AgentModels: models}, task)
		if err != nil {
			t.Fatalf("buildLocalAgentBenchmarkSpec(%s) error: %v", agent, err)
		}
		modelIdx := slices.Index(spec.args, "--model")
		if modelIdx < 0 || modelIdx+1 >= len(spec.args) {
			t.Fatalf("%s args = %+v, want --model override", agent, spec.args)
		}
		if got := spec.args[modelIdx+1]; got != models[agent] {
			t.Fatalf("%s --model = %q, want %q", agent, got, models[agent])
		}
	}
}

// Test_BuildLocalAgentBenchmarkSpec_aiderIncludesRequiredChangedFiles verifies
// that Aider's benchmark spec appends the task's required changed files and
// keeps the safe-Aider flags (git root, no gitignore mutation, no auto-commits).
func Test_BuildLocalAgentBenchmarkSpec_aiderIncludesRequiredChangedFiles(t *testing.T) {
	task := Task{
		ID:                   "safety-policy-path-escape",
		Title:                "Reject path escape writes",
		Objective:            "Ensure patch/create requests cannot write outside the workspace root.",
		RequiredChangedFiles: []string{"internal/workspace/workspace.go"},
		RequiredDiffTerms:    []string{"ErrPathEscapesWorkspace"},
	}
	spec, err := buildLocalAgentBenchmarkSpec("aider", LocalAgentBenchmarkRequest{}, task)
	if err != nil {
		t.Fatalf("buildLocalAgentBenchmarkSpec aider error: %v", err)
	}
	if spec.binary != "aider" {
		t.Fatalf("binary = %q, want aider", spec.binary)
	}
	for _, want := range []string{"--git", "--no-gitignore", "--skip-sanity-check-repo", "--no-auto-commits", "--no-dirty-commits", "--yes-always", "--message"} {
		if !slices.Contains(spec.args, want) {
			t.Fatalf("args = %+v, missing safe-Aider flag %q", spec.args, want)
		}
	}
	if slices.Index(spec.args, "--file") < 0 {
		t.Fatalf("args = %+v, missing --file before required changed files", spec.args)
	}
	for _, want := range task.RequiredChangedFiles {
		if !slices.Contains(spec.args, want) {
			t.Fatalf("args = %+v, missing required changed file %q", spec.args, want)
		}
	}
	msgIdx := slices.Index(spec.args, "--message")
	if msgIdx < 0 || msgIdx+1 >= len(spec.args) {
		t.Fatalf("args = %+v, missing --message <prompt>", spec.args)
	}
	if spec.args[msgIdx+1] != localAgentBenchmarkPrompt(task) {
		t.Fatalf("--message prompt = %q, want benchmark prompt", spec.args[msgIdx+1])
	}
}

// Test_BuildLocalAgentBenchmarkSpec_externalAgentsUseNoninteractiveArgs pins
// the noninteractive edit-capable arg shape per external agent so a refactor
// that quietly drops --safe-mode, --yes-always, --quiet, or --auto-approve
// will fail the contract test.
func Test_BuildLocalAgentBenchmarkSpec_externalAgentsUseNoninteractiveArgs(t *testing.T) {
	task := Task{
		ID:                   "docs-roadmap-cli-first",
		Title:                "Keep roadmap CLI-first",
		Objective:            "Refresh roadmap wording.",
		RequiredChangedFiles: []string{"docs/ROADMAP.md"},
		RequiredDiffTerms:    []string{"CLI-first"},
	}
	cases := []struct {
		agent      string
		wantBinary string
		mustHave   []string
	}{
		{
			agent:      "claude_code",
			wantBinary: "claude",
			mustHave:   []string{"--safe-mode", "--no-session-persistence", "--permission-mode", "bypassPermissions", "--output-format", "json", "--print"},
		},
		{
			agent:      "aider",
			wantBinary: "aider",
			mustHave:   []string{"--git", "--no-gitignore", "--skip-sanity-check-repo", "--no-auto-commits", "--no-dirty-commits", "--yes-always", "--file"},
		},
		{
			agent:      "goose",
			wantBinary: "goose",
			mustHave:   []string{"run", "--no-session", "--quiet", "--max-turns"},
		},
		{
			agent:      "oh_my_pi",
			wantBinary: "omp",
			mustHave:   []string{"--no-session", "--auto-approve", "--approval-mode", "yolo", "--max-time"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.agent, func(t *testing.T) {
			spec, err := buildLocalAgentBenchmarkSpec(tc.agent, LocalAgentBenchmarkRequest{}, task)
			if err != nil {
				t.Fatalf("buildLocalAgentBenchmarkSpec(%s) error: %v", tc.agent, err)
			}
			if spec.binary != tc.wantBinary {
				t.Fatalf("binary = %q, want %q", spec.binary, tc.wantBinary)
			}
			for _, want := range tc.mustHave {
				if !slices.Contains(spec.args, want) {
					t.Fatalf("args = %+v, missing noninteractive flag %q", spec.args, want)
				}
			}
		})
	}
}
