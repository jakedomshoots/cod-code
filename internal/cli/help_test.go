package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_prints_help_when_help_flag_is_supplied(t *testing.T) {
	var out bytes.Buffer

	if err := Run(context.Background(), &out, []string{"--help"}); err != nil {
		t.Fatalf("Run --help: %v", err)
	}

	body := out.String()
	for _, want := range []string{
		"ceo-packet",
		"Primary flow:",
		"ceo-packet start <path>",
		"ceo-packet run [flags] <task>",
		"ceo-packet gauntlet [flags]",
		"ceo-packet doctor [flags]",
		"ceo-packet inbox [flags]",
		"ceo-packet status [flags]",
		"ceo-packet resume <job> --answer <text>",
		"ceo-packet retry <job>",
		"ceo-packet rollback <report>",
		"ceo-packet explain-failure <job>",
		"gauntlet [flags]                Run market or production benchmark gauntlets",
		"ceo-packet gauntlet --suite production-core [flags]",
		"--init-config",
		"--init-example-adapters",
		"--config-check",
		"--config-doctor",
		"--config-explain",
		"--config-completions",
		"--shell <zsh|bash|fish>",
		"--output-dir <path>",
		"--history",
		"--review-details",
		"--provider-health",
		"--doctor-provider <name>",
		"--doctor",
		"--demo",
		"--quickstart <path>",
		"--start <path>",
		"--inbox",
		"--provider-wizard",
		"--init-demo-repo",
		"--tui",
		"--write-policy <policy>",
		"write intent previews by default",
		"--adapter <name>",
		"--ceo-provider",
		"--risk-provider",
		"--kind-provider",
		"--version",
		"--format <json|text|events>",
		"--dry-run",
		"--approve-preview",
		"--job-timeout-ms <n>",
		"--require-checks",
		"--check-attempts <n>",
		"--check-backoff-ms <n>",
		"--repair-preset <standard>",
		"--max-ceo-iterations <n>",
		"--workspace-brief-max-files <n>",
		"--workspace-brief-exclude <glob>",
		"--interactive",
		"--max-subagents <n>",
		"--subagent-concurrency <n>",
		"--subagent-attempts <n>",
		"--subagent-backoff-ms <n>",
		"--max-tool-requests <n>",
		"--no-progress-stop <n>",
		"--tool-command-timeout-ms",
		"--research-command",
		"--model-command-timeout-ms",
		"--provider <name>",
		"--recommendation <label>",
		"--task <text>",
		"--verdict <value>",
		"--limit <n>",
		"--since <time>",
		"--until <time>",
		"--summary-only",
		"--top-providers <n>",
		"--job-events <id>",
		"--context-trace <id>",
		"--judge-job <id>",
		"--human-verdict <accept|reject>",
		"--judgment-note <text>",
		"--continue-job <id>",
		"Provider quick start:",
		"Codex CLI: ceo-packet config init --adapter codex",
		"Kimi CLI: ceo-packet config init --adapter kimi",
		"OpenRouter: use --provider-wizard openrouter; missing OPENROUTER_API_KEY is blocked setup, not a failed benchmark",
		"--http-provider",
		"--http-max-output-tokens",
		"--http-input-cost-per-million",
		"--http-output-cost-per-million",
		"--provider-health-avoid-failure-rate",
		"--provider-health-watch-failure-rate",
		"--provider-health-watch-cost-per-attempt-microusd",
		"CEO_RESEARCH_COMMAND_JSON",
		`create patch JSON: {"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("help output missing %q:\n%s", want, body)
		}
	}
	if strings.Index(body, "Primary flow:") > strings.Index(body, "Advanced commands:") {
		t.Fatalf("help output should list primary flow before advanced commands:\n%s", body)
	}
	for _, advanced := range []string{"review [flags]", "context [flags] <job>", "eval [flags]"} {
		if strings.Index(body, advanced) < strings.Index(body, "Advanced commands:") {
			t.Fatalf("advanced command %q appeared before advanced section:\n%s", advanced, body)
		}
	}
}

func Test_Run_prints_help_when_short_help_flag_is_supplied(t *testing.T) {
	var out bytes.Buffer

	if err := Run(context.Background(), &out, []string{"-h"}); err != nil {
		t.Fatalf("Run -h: %v", err)
	}

	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("help output missing Usage:\n%s", out.String())
	}
}
