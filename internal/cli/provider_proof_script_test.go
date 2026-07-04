package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProviderProofScript_dryRunWritesKimiPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "kimi",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Mode: dry-run",
		"- Provider: kimi",
		"scripts/kimi-model-command.sh",
		"| cross-language-js-state-reducer | planned |",
		"| cross-language-python-retry-policy | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	requireTextFile(t, filepath.Join(outputDir, "cross-language-js-state-reducer", "plan.md"))
	requireTextFile(t, filepath.Join(outputDir, "cross-language-python-retry-policy", "plan.md"))
}

func Test_ProviderProofScript_dryRunWritesCodexPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "codex",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof codex dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Provider: codex",
		"scripts/codex-model-command.sh",
		"| cross-language-js-state-reducer | planned |",
		"| cross-language-python-retry-policy | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func Test_ProviderProofScript_dryRunWritesOpenAIHTTPPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "openai",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof openai dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Provider: openai",
		"- Provider mode: http-provider",
		"- HTTP preset: openai",
		"- API key env: OPENAI_API_KEY",
		"| cross-language-js-state-reducer | planned |",
		"| cross-language-python-retry-policy | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func Test_ProviderProofScript_liveBlocksWhenHTTPKeyMissing(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--provider", "openrouter",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = withoutEnv(os.Environ(), "OPENROUTER_API_KEY")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider proof unexpectedly passed without OPENROUTER_API_KEY:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"- Provider: openrouter",
		"- Provider mode: http-provider",
		"- API key env: OPENROUTER_API_KEY",
		"| provider_setup | blocked_missing_key | blocked.md |",
		"- Overall: blocked",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}

	blocked := readTextFile(t, filepath.Join(outputDir, "blocked.md"))
	if !strings.Contains(blocked, "OPENROUTER_API_KEY") {
		t.Fatalf("blocked.md missing key guidance:\n%s", blocked)
	}
	for _, path := range []string{
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "env.template"),
		filepath.Join(outputDir, "commands.sh"),
		filepath.Join(outputDir, "setup-checklist.md"),
	} {
		requireTextFile(t, path)
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"provider": "openrouter"`,
		`"api_key_env": "OPENROUTER_API_KEY"`,
		`"blocked_reason": "missing_api_key_env"`,
		`"setup_checklist_item_count": 5`,
		`"setup_artifacts_sha256": {`,
		`"blocked.md": "`,
		`"commands.sh": "`,
		`"env.template": "`,
		`"setup-checklist.md": "`,
		`"secret_value_saved": false`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
	envTemplate := readTextFile(t, filepath.Join(outputDir, "env.template"))
	if !strings.Contains(envTemplate, "OPENROUTER_API_KEY=") {
		t.Fatalf("env.template missing OPENROUTER_API_KEY:\n%s", envTemplate)
	}
	commands := readTextFile(t, filepath.Join(outputDir, "commands.sh"))
	if !strings.Contains(commands, "scripts/provider-proof.sh --provider openrouter") {
		t.Fatalf("commands.sh missing rerun command:\n%s", commands)
	}
}

func Test_ProviderProofScript_liveBlocksWhenHTTPKeyEmpty(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--provider", "moonshot",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(withoutEnv(os.Environ(), "MOONSHOT_API_KEY"), "MOONSHOT_API_KEY=")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider proof unexpectedly passed with empty MOONSHOT_API_KEY:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	if !strings.Contains(index, "| provider_setup | blocked_empty_key | blocked.md |") {
		t.Fatalf("index.md missing empty key blocker:\n%s", index)
	}
	blocked := readTextFile(t, filepath.Join(outputDir, "blocked.md"))
	if !strings.Contains(blocked, "has `MOONSHOT_API_KEY` set, but it is empty") {
		t.Fatalf("blocked.md missing empty key guidance:\n%s", blocked)
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"blocked_reason": "empty_api_key_env"`,
		`"setup_result_status": "blocked_empty_key"`,
		`"setup_checklist_item_count": 5`,
		`"setup_artifacts_sha256": {`,
		`"secret_value_saved": false`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}

func Test_ProviderProofScript_liveFailsWhenBenchmarkSummaryFails(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")
	binDir := t.TempDir()
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), "#!/bin/sh\nexit 0\n")
	writeExecutableScript(t, filepath.Join(binDir, "go"), `#!/bin/sh
if [ "$1" = "build" ]; then
  exit 0
fi
if [ "$1" = "run" ]; then
  output=""
  task=""
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --output-dir)
        shift
        output="$1"
        ;;
      --local-agent-benchmark-task)
        shift
        task="$1"
        ;;
    esac
    shift || true
  done
  mkdir -p "$output"
  cat > "$output/summary.json" <<JSON
{"task_id":"$task","passed":0,"partial":0,"failed":1,"timed_out":0,"incomplete_evidence":1}
JSON
  exit 0
fi
exit 1
`)

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--provider", "kimi",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider proof unexpectedly passed with failing benchmark summary:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	if !strings.Contains(index, "| cross-language-js-state-reducer | fail |") {
		t.Fatalf("index.md missing failing benchmark row:\n%s", index)
	}
}

func withoutEnv(env []string, key string) []string {
	prefix := key + "="
	next := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		next = append(next, entry)
	}
	return next
}
