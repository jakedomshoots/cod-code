package cli

import (
	"fmt"
	"io"
)

const helpText = `cod — Cod Code terminal

Cod Code is the local agentic coding cockpit: start a lane, send a task, review the catch, and keep proof close.

Simple commands:
  cod                              Open the Cod Code TUI
  cod chat                         Open the Cod Code TUI
  cod dev                          Open the Cod Code TUI
  cod run <task>                   Send one coding task
  cod start .                      Guided setup for this repo
  cod doctor                       Check local setup
  cod status                       Show recent runs
  cod inbox                        Review items needing your input

Usage:
  cod
  cod chat
  cod dev
  cod start <path>
  cod run [flags] <task>
  cod doctor [flags]
  cod inbox [flags]
  cod status [flags]
  cod oauth list
  cod oauth doctor [provider]
  cod oauth init <provider> --workspace <path>
  cod browser read http://localhost:3000 --format text
  cod computer snapshot <app>
  cod tools manifest
  cod production-status [flags]
  cod production-actions [flags]
  cod production-finalize --dry-run [flags]
  cod retry <job>
  cod resume <job> --answer <text>
  cod rollback <report>
  cod explain-failure <job>

Core flags:
  --workspace <path>              Use a workspace directory
  --check <command...> --         Require a verification command
  --write-policy <policy>         observe, preview, dry-run, approved-write, or trusted-local
  --provider-wizard <preset>      Create a provider for openai, openrouter, kimi-code, moonshot, or minimax
  --adapter <name>                Use codex, kimi, claude, opencode, aider, or goose
  --format <json|text>            Print JSON or compact text
  --version                       Print version
  --help, -h                      Print this compact help
  --help-advanced                 Print the full reference

First run:
  1. cod
  2. cod start .
  3. cod oauth doctor --format text
  4. cod run --workspace . --check go test ./... -- "Fix one real task"

Examples:
  cod
  cod chat
  cod dev
  cod start .
  cod doctor --format text
  cod oauth init kimi --workspace . --format text
  cod run --workspace . --check go test ./... -- "Fix one real task"
  cod production-status --workspace . --format text
  cod retry latest --workspace .

Advanced: run cod --help-advanced for the full reference.
`

func runHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, helpText)
	return err
}

func runAdvancedHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, advancedHelpText)
	return err
}
