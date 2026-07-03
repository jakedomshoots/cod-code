package eval

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_RunLocalAgentBenchmark_isolates_ceo_artifacts_outside_scored_workspace(t *testing.T) {
	binDir := t.TempDir()
	writeExecutableContent(t, filepath.Join(binDir, "ceo-packet"), `#!/bin/sh
artifact_root=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --artifact-root) artifact_root="$2"; shift 2 ;;
    *) shift ;;
  esac
done
test -n "$artifact_root" || exit 17
mkdir -p "$artifact_root/ceo-artifacts"
printf '{}\n' > "$artifact_root/ceo-artifacts/jobs.jsonl"
mkdir -p .omo/evidence
printf 'agent evidence\n' > .omo/evidence/docs-roadmap-cli-first.md
cat > docs/ROADMAP.md <<'EOF'
# Roadmap

CLI-first dogfood and recovery come before GUI work.
EOF
printf '{"verdict":"pass"}\n'
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := t.TempDir()
	tasksDir := filepath.Join(root, "tasks")
	writeTaskSpec(t, tasksDir, docsRoadmapTaskSpec())

	summary, err := RunLocalAgentBenchmark(context.Background(), LocalAgentBenchmarkRequest{
		TasksDir:         tasksDir,
		OutputDir:        filepath.Join(root, "benchmark"),
		TimeoutSeconds:   5,
		Agents:           []string{"ceo_harness"},
		CEOHarnessBinary: filepath.Join(binDir, "ceo-packet"),
		BenchmarkTaskID:  "docs-roadmap-cli-first",
	})
	if err != nil {
		t.Fatalf("RunLocalAgentBenchmark returned error: %v", err)
	}
	result := summary.Results[0]
	if result.Status != localAgentStatusPass || len(result.ExtraChangedFiles) != 0 {
		t.Fatalf("result = %+v, want pass with no extra changed files", result)
	}
	artifactRoot := commandFlagValue(result.Command, "--artifact-root")
	if artifactRoot == "" {
		t.Fatalf("command = %+v, want --artifact-root", result.Command)
	}
	requireFile(t, filepath.Join(artifactRoot, "ceo-artifacts", "jobs.jsonl"))
	if _, statErr := os.Stat(filepath.Join(result.WorkspaceDir, "ceo-artifacts")); !os.IsNotExist(statErr) {
		t.Fatalf("workspace ceo-artifacts should not exist: %v", statErr)
	}
}

func Test_BuildLocalAgentBenchmarkSpec_uses_model_command_mode(t *testing.T) {
	task := Task{
		ID:                   "docs-roadmap-cli-first",
		Title:                "Keep roadmap CLI-first",
		Objective:            "Refresh roadmap wording.",
		RequiredChangedFiles: []string{"docs/ROADMAP.md"},
		RequiredDiffTerms:    []string{"CLI-first"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
	}

	spec, err := buildLocalAgentBenchmarkSpec("ceo_harness", LocalAgentBenchmarkRequest{
		CEOHarnessBinary:         "/tmp/ceo-packet",
		CEOBenchmarkMode:         ceoBenchmarkModeModelCommand,
		CEOBenchmarkModelCommand: []string{"go", "run", "model.go"},
	}, task)
	if err != nil {
		t.Fatalf("buildLocalAgentBenchmarkSpec returned error: %v", err)
	}
	if spec.binary != "/tmp/ceo-packet" {
		t.Fatalf("binary = %q, want custom CEO binary", spec.binary)
	}
	if !slices.Contains(spec.args, "--apply-model-patches") {
		t.Fatalf("args = %+v, want --apply-model-patches", spec.args)
	}
	if got := commandFlagValue(spec.args, "--subagent-attempts"); got != "2" {
		t.Fatalf("args = %+v, want --subagent-attempts 2, got %q", spec.args, got)
	}
	if got := commandFlagValue(spec.args, "--no-progress-stop"); got != "2" {
		t.Fatalf("args = %+v, want --no-progress-stop 2, got %q", spec.args, got)
	}
	if slices.Contains(spec.args, "--replace") {
		t.Fatalf("args = %+v, model-command mode must not use synthetic --replace", spec.args)
	}
	modelCommandIndex := slices.Index(spec.args, "--model-command")
	if modelCommandIndex < 0 {
		t.Fatalf("args = %+v, want --model-command", spec.args)
	}
	if got := spec.args[modelCommandIndex+1 : modelCommandIndex+4]; !slices.Equal(got, []string{"go", "run", "model.go"}) {
		t.Fatalf("model command args = %+v, want go run model.go", got)
	}
	ceoModelCommandIndex := slices.Index(spec.args, "--ceo-model-command")
	if ceoModelCommandIndex < 0 {
		t.Fatalf("args = %+v, want --ceo-model-command for real CEO delegation/review", spec.args)
	}
	if got := spec.args[ceoModelCommandIndex+1 : ceoModelCommandIndex+4]; !slices.Equal(got, []string{"go", "run", "model.go"}) {
		t.Fatalf("CEO model command args = %+v, want go run model.go", got)
	}
	checkIndex := slices.Index(spec.args, "--check")
	if checkIndex < 0 {
		t.Fatalf("args = %+v, want --check from required benchmark command", spec.args)
	}
	if got := spec.args[checkIndex+1 : checkIndex+4]; !slices.Equal(got, []string{"sh", "-c", "go test ./internal/cli -count=1"}) {
		t.Fatalf("check args = %+v, want shell-wrapped benchmark command", got)
	}
	if !strings.Contains(spec.args[len(spec.args)-1], task.ID) {
		t.Fatalf("prompt = %q, want benchmark task id", spec.args[len(spec.args)-1])
	}
}

func Test_BuildLocalAgentBenchmarkSpec_modelCommandModeRequiresCommand(t *testing.T) {
	_, err := buildLocalAgentBenchmarkSpec("ceo_harness", LocalAgentBenchmarkRequest{
		CEOBenchmarkMode: ceoBenchmarkModeModelCommand,
	}, Task{ID: "docs-roadmap-cli-first"})

	if err == nil || !strings.Contains(err.Error(), "requires --ceo-benchmark-model-command-json") {
		t.Fatalf("error = %v, want missing model command error", err)
	}
}

func Test_BuildLocalAgentBenchmarkSpec_syntheticModePassesRequiredCheck(t *testing.T) {
	task := Task{
		ID:                   "docs-roadmap-cli-first",
		Title:                "Keep roadmap CLI-first",
		Objective:            "Refresh roadmap wording.",
		RequiredChangedFiles: []string{"docs/ROADMAP.md"},
		RequiredDiffTerms:    []string{"CLI-first"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
	}

	spec, err := buildLocalAgentBenchmarkSpec("ceo_harness", LocalAgentBenchmarkRequest{}, task)
	if err != nil {
		t.Fatalf("buildLocalAgentBenchmarkSpec returned error: %v", err)
	}
	checkIndex := slices.Index(spec.args, "--check")
	if checkIndex < 0 {
		t.Fatalf("args = %+v, want --check from required benchmark command", spec.args)
	}
	if checkIndex == 0 || spec.args[checkIndex-1] == "--" {
		t.Fatalf("args = %+v, synthetic mode should not terminate a prior delimited flag before --check", spec.args)
	}
}

func Test_LocalAgentBenchmarkPrompt_names_required_artifact_contract(t *testing.T) {
	task := Task{
		ID:                   "docs-roadmap-cli-first",
		Title:                "Keep roadmap CLI-first",
		Objective:            "Refresh roadmap wording.",
		RequiredChangedFiles: []string{"docs/ROADMAP.md"},
		RequiredArtifacts:    []string{".omo/evidence/docs-roadmap-cli-first.md"},
		RequiredDiffTerms:    []string{"CLI-first"},
		RequiredCommands:     []string{"go test ./internal/cli -count=1"},
	}

	prompt := localAgentBenchmarkPrompt(task)

	for _, want := range []string{
		"Required evidence artifacts: .omo/evidence/docs-roadmap-cli-first.md.",
		"Create every required evidence artifact as a non-empty markdown file inside the workspace.",
		"must summarize the change, commands run, and verification result",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want %q", prompt, want)
		}
	}
}

func Test_BuildLocalAgentBenchmarkSpec_httpProviderModeWritesWorkspaceConfig(t *testing.T) {
	task := Task{
		ID:                   "safety-policy-path-escape",
		Title:                "Reject path escape writes",
		Objective:            "Ensure patch/create requests cannot write outside the workspace root.",
		RequiredChangedFiles: []string{"internal/workspace/workspace.go"},
		RequiredDiffTerms:    []string{"path escapes workspace"},
		RequiredCommands:     []string{"go test ./internal/workspace -run Test_.*[Pp]ath -count=1"},
	}

	spec, err := buildLocalAgentBenchmarkSpec("ceo_harness", LocalAgentBenchmarkRequest{
		CEOHarnessBinary:                  "/tmp/ceo-packet",
		CEOBenchmarkMode:                  ceoBenchmarkModeHTTPProvider,
		CEOBenchmarkProviderName:          "main",
		CEOBenchmarkProviderPreset:        "openrouter",
		CEOBenchmarkProviderModel:         "openai/gpt-5-mini",
		CEOBenchmarkProviderAPIKeyEnv:     "OPENROUTER_API_KEY",
		CEOBenchmarkProviderMaxOutputToks: 2048,
	}, task)
	if err != nil {
		t.Fatalf("buildLocalAgentBenchmarkSpec returned error: %v", err)
	}
	if slices.Contains(spec.args, "--model-command") || slices.Contains(spec.args, "--replace") {
		t.Fatalf("args = %+v, http-provider mode should not use model-command or synthetic replace", spec.args)
	}
	if !slices.Contains(spec.args, "--apply-model-patches") || !slices.Contains(spec.args, "--check") {
		t.Fatalf("args = %+v, want apply-model-patches and required check", spec.args)
	}
	var cfg config.Config
	if err := json.Unmarshal(spec.workspaceConfig, &cfg); err != nil {
		t.Fatalf("workspace config must decode: %v", err)
	}
	provider := cfg.Providers["main"].HTTP
	if cfg.CEOProvider != "main" || cfg.ProviderPolicy.DefaultProvider != "main" {
		t.Fatalf("config = %#v, want main routed as CEO and default provider", cfg)
	}
	if provider.Model != "openai/gpt-5-mini" || provider.APIKeyEnv != "OPENROUTER_API_KEY" || provider.ResponseFormat != "json_object" {
		t.Fatalf("http provider = %#v, want OpenRouter JSON provider", provider)
	}
	if provider.MaxOutputTokens != 2048 {
		t.Fatalf("max output tokens = %d, want 2048", provider.MaxOutputTokens)
	}
}

func commandFlagValue(command []string, flag string) string {
	index := slices.Index(command, flag)
	if index < 0 || index+1 >= len(command) {
		return ""
	}
	return command[index+1]
}
