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

func Test_ProviderProofScript_dryRunWritesKimiCodeHTTPPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "kimi-code",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof kimi-code dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Provider: kimi-code",
		"- Provider mode: http-provider",
		"- HTTP preset: kimi-code",
		"- HTTP model: kimi-for-coding",
		"- API key env: KIMI_CODE_API_KEY",
		"| cross-language-js-state-reducer | planned |",
		"| cross-language-python-retry-policy | planned |",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
}

func Test_ProviderProofScript_dryRunWritesMiniMaxHTTPPlan(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-proof")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-proof.sh"),
		"--dry-run",
		"--provider", "minimax",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider proof minimax dry-run failed: %v\n%s", err, string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"# Provider Proof Evidence",
		"- Provider: minimax",
		"- Provider mode: http-provider",
		"- HTTP preset: minimax",
		"- HTTP model: MiniMax-M3",
		"- API key env: MINIMAX_API_KEY",
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
		`"setup_checklist_item_count": 6`,
		`"setup_artifacts_sha256": {`,
		`"blocked.md": "`,
		`"commands.sh": "`,
		`"env.template": "`,
		`"setup-checklist.md": "`,
		`"command_script_secret_policy": "no_secret_assignment"`,
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
	if !strings.Contains(commands, "scripts/provider-setup-preflight.sh --providers openrouter") {
		t.Fatalf("commands.sh missing provider setup preflight:\n%s", commands)
	}
	for _, want := range []string{
		"${OPENROUTER_API_KEY+x}",
		"${OPENROUTER_API_KEY}",
		"provider setup: OPENROUTER_API_KEY is not set",
		"provider setup: OPENROUTER_API_KEY is empty",
	} {
		if !strings.Contains(commands, want) {
			t.Fatalf("commands.sh missing env guard %q:\n%s", want, commands)
		}
	}
	if strings.Contains(commands, "OPENROUTER_API_KEY=") || strings.Contains(commands, "<redacted>") {
		t.Fatalf("commands.sh should not include key assignment, even redacted:\n%s", commands)
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
		"--provider", "minimax",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(withoutEnv(os.Environ(), "MINIMAX_API_KEY"), "MINIMAX_API_KEY=")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider proof unexpectedly passed with empty MINIMAX_API_KEY:\n%s", string(output))
	}

	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	if !strings.Contains(index, "| provider_setup | blocked_empty_key | blocked.md |") {
		t.Fatalf("index.md missing empty key blocker:\n%s", index)
	}
	blocked := readTextFile(t, filepath.Join(outputDir, "blocked.md"))
	if !strings.Contains(blocked, "has `MINIMAX_API_KEY` set, but it is empty") {
		t.Fatalf("blocked.md missing empty key guidance:\n%s", blocked)
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"blocked_reason": "empty_api_key_env"`,
		`"setup_result_status": "blocked_empty_key"`,
		`"setup_checklist_item_count": 6`,
		`"setup_artifacts_sha256": {`,
		`"command_script_secret_policy": "no_secret_assignment"`,
		`"secret_value_saved": false`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
}

func Test_ProviderSetupPreflightScript_writesBlockedEvidenceWithoutSecretValues(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-setup")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-setup-preflight.sh"),
		"--providers", "openrouter,kimi-code,minimax",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(
		withoutEnv(withoutEnv(withoutEnv(os.Environ(), "OPENROUTER_API_KEY"), "KIMI_CODE_API_KEY"), "MINIMAX_API_KEY"),
		"OPENROUTER_API_KEY=",
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("provider setup preflight unexpectedly passed:\n%s", string(output))
	}

	for _, path := range []string{
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "commands.sh"),
		filepath.Join(outputDir, "blocked-providers.txt"),
	} {
		requireTextFile(t, path)
	}
	index := readTextFile(t, filepath.Join(outputDir, "index.md"))
	for _, want := range []string{
		"Status: blocked",
		"| Provider | Status | Env | Model |\n| --- | --- | --- | --- |",
		"| openrouter | empty_env | `OPENROUTER_API_KEY` |",
		"| kimi-code | missing_env | `KIMI_CODE_API_KEY` |",
		"| minimax | missing_env | `MINIMAX_API_KEY` |",
		"Secret values were not printed or saved.",
	} {
		if !strings.Contains(index, want) {
			t.Fatalf("index.md missing %q:\n%s", want, index)
		}
	}
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "blocked"`,
		`"provider_count": 3`,
		`"ready_count": 0`,
		`"blocked_count": 3`,
		`"command_script_secret_policy": "no_secret_assignment"`,
		`"secret_value_saved": false`,
		`"commands_sha256": "`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
	if strings.Contains(summary, "secret-openrouter") || strings.Contains(index, "secret-openrouter") {
		t.Fatalf("provider setup preflight leaked secret-like value")
	}

	commands := readTextFile(t, filepath.Join(outputDir, "commands.sh"))
	for _, want := range []string{
		"# blocked command: sh scripts/provider-proof.sh --provider openrouter",
		"# reason: OPENROUTER_API_KEY is missing or empty; export it before running provider proof.",
		"# blocked command: sh scripts/provider-proof.sh --provider kimi-code",
		"# reason: KIMI_CODE_API_KEY is missing or empty; export it before running provider proof.",
		"# blocked command: sh scripts/provider-proof.sh --provider minimax",
		"# reason: MINIMAX_API_KEY is missing or empty; export it before running provider proof.",
	} {
		if !strings.Contains(commands, want) {
			t.Fatalf("commands.sh missing %q:\n%s", want, commands)
		}
	}
	for _, blockedRunnable := range []string{
		"\nsh scripts/provider-proof.sh --provider openrouter",
		"\nsh scripts/provider-proof.sh --provider kimi-code",
		"\nsh scripts/provider-proof.sh --provider minimax",
	} {
		if strings.Contains(commands, blockedRunnable) {
			t.Fatalf("commands.sh should not contain runnable blocked provider command %q:\n%s", blockedRunnable, commands)
		}
	}
}

func Test_ProviderSetupPreflightScript_passesWhenHTTPKeysArePresent(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	outputDir := filepath.Join(t.TempDir(), "provider-setup")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "provider-setup-preflight.sh"),
		"--providers", "openrouter kimi-code minimax",
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	cmd.Env = append(
		withoutEnv(withoutEnv(withoutEnv(os.Environ(), "OPENROUTER_API_KEY"), "KIMI_CODE_API_KEY"), "MINIMAX_API_KEY"),
		"OPENROUTER_API_KEY=secret-openrouter",
		"KIMI_CODE_API_KEY=secret-kimi",
		"MINIMAX_API_KEY=secret-minimax",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider setup preflight failed with keys present: %v\n%s", err, string(output))
	}

	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	for _, want := range []string{
		`"status": "pass"`,
		`"provider_count": 3`,
		`"ready_count": 3`,
		`"blocked_count": 0`,
		`"secret_value_saved": false`,
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary.json missing %q:\n%s", want, summary)
		}
	}
	if strings.Contains(summary, "secret-openrouter") || strings.Contains(summary, "secret-kimi") || strings.Contains(summary, "secret-minimax") {
		t.Fatalf("provider setup preflight leaked secret values:\n%s", summary)
	}

	commands := readTextFile(t, filepath.Join(outputDir, "commands.sh"))
	for _, want := range []string{
		"sh scripts/provider-proof.sh --provider openrouter --output-dir .omo/evidence/provider-proof-openrouter --timeout-seconds 600",
		"sh scripts/provider-proof.sh --provider kimi-code --output-dir .omo/evidence/provider-proof-kimi-code --timeout-seconds 600",
		"sh scripts/provider-proof.sh --provider minimax --output-dir .omo/evidence/provider-proof-minimax --timeout-seconds 600",
	} {
		if !strings.Contains(commands, want) {
			t.Fatalf("commands.sh missing runnable command %q:\n%s", want, commands)
		}
	}
	if strings.Contains(commands, "# blocked command:") {
		t.Fatalf("commands.sh should not include blocked commands when all providers are ready:\n%s", commands)
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
	summary := readTextFile(t, filepath.Join(outputDir, "summary.json"))
	if !strings.Contains(summary, `"status": "fail"`) || strings.Contains(summary, `"blocked_reason"`) {
		t.Fatalf("summary.json = %s, want fresh fail summary without stale blocked metadata", summary)
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
