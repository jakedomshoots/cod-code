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
		"--doctor",
		"--demo",
		"--quickstart <path>",
		"--start <path>",
		"--inbox",
		"--provider-wizard",
		"--init-demo-repo",
		"--write-policy <policy>",
		"write intent previews by default",
		"--adapter <name>",
		"--version",
		"--format <json|text>",
		"--help-advanced",
		"Provider quick start:",
		"Codex CLI: ceo-packet config init --adapter codex",
		"Kimi CLI: ceo-packet config init --adapter kimi",
		"OpenRouter: use --provider-wizard openrouter; missing OPENROUTER_API_KEY is blocked setup, not a failed benchmark",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("help output missing %q:\n%s", want, body)
		}
	}
	for _, advanced := range []string{
		"Advanced flags:",
		"HTTP provider setup:",
		"--http-input-cost-per-million",
		"--provider-health-watch-cost-per-attempt-microusd",
		`create patch JSON: {"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`,
	} {
		if strings.Contains(body, advanced) {
			t.Fatalf("compact help should not include %q:\n%s", advanced, body)
		}
	}
}

func Test_Run_prints_advanced_help_when_requested(t *testing.T) {
	var out bytes.Buffer

	if err := Run(context.Background(), &out, []string{"--help-advanced"}); err != nil {
		t.Fatalf("Run --help-advanced: %v", err)
	}

	body := out.String()
	for _, want := range []string{
		"Advanced flags:",
		"--init-config",
		"--provider-health",
		"--context-trace <id>",
		"--human-verdict <accept|reject>",
		"HTTP provider setup:",
		"--http-input-cost-per-million",
		"--provider-health-watch-cost-per-attempt-microusd",
		"CEO_RESEARCH_COMMAND_JSON",
		"ceo-packet production-actions --workspace . --format text",
		"ceo-packet production-actions --workspace . --action-state empty_env --commands-only",
		`create patch JSON: {"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("advanced help missing %q:\n%s", want, body)
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
