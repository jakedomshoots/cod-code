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
		"cod — Cod Code terminal",
		"Simple commands:",
		"cod                              Open the Cod Code TUI",
		"cod chat                         Open the Cod Code TUI",
		"cod dev                          Open the Cod Code TUI",
		"cod run <task>",
		"cod start .",
		"cod doctor",
		"cod status",
		"cod inbox",
		"cod oauth list",
		"cod production-status [flags]",
		"--workspace <path>",
		"--check <command...> --",
		"--write-policy <policy>",
		"--adapter <name>",
		"--version",
		"--format <json|text>",
		"--help-advanced",
		"cod run --workspace . --check go test ./... -- \"Fix one real task\"",
		"cod production-status --workspace . --format text",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("help output missing %q:\n%s", want, body)
		}
	}
	for _, legacy := range []string{
		"Provider quick start:",
		"Codex CLI: ceo-packet config init --adapter codex",
		"Kimi CLI: ceo-packet config init --adapter kimi",
		"OpenRouter: use --provider-wizard openrouter; missing OPENROUTER_API_KEY is blocked setup, not a failed benchmark",
	} {
		if strings.Contains(body, legacy) {
			t.Fatalf("compact help must omit legacy provider quick-start %q:\n%s", legacy, body)
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

func Test_Run_compact_help_includes_cod_code_identity(t *testing.T) {
	var out bytes.Buffer

	if err := Run(context.Background(), &out, []string{"--help"}); err != nil {
		t.Fatalf("Run --help: %v", err)
	}

	body := out.String()
	for _, want := range []string{
		"cod — Cod Code terminal",
		"local agentic coding cockpit",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("compact help missing identity marker %q:\n%s", want, body)
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
		"OAuth CLI setup:",
		"CEO_RESEARCH_COMMAND_JSON",
		"cod production-actions --workspace . --format text",
		"cod production-actions --workspace . --action-state empty_env --commands-only",
		`create patch JSON: {"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("advanced help missing %q:\n%s", want, body)
		}
	}
}

func Test_Run_advanced_help_includes_cod_code_identity(t *testing.T) {
	var out bytes.Buffer

	if err := Run(context.Background(), &out, []string{"--help-advanced"}); err != nil {
		t.Fatalf("Run --help-advanced: %v", err)
	}

	body := out.String()
	for _, want := range []string{
		"cod — Cod Code terminal",
		"local agentic coding cockpit",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("advanced help missing identity marker %q:\n%s", want, body)
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
