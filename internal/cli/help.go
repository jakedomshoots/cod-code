package cli

import (
	"fmt"
	"io"
)

const helpText = `ceo-packet — Cod Code CLI

Alpha Cod delegates bounded swimmers, inspects catches, and owns the final verdict.

Primary flow:
  start <path>                    Guided setup/check/doctor start flow
  run [flags] <task>              Run the Alpha Cod packet loop
  gauntlet [flags]                Run market or production benchmark gauntlets
  doctor [flags]                  Run harness health checks
  inbox [flags]                   Review queue alias with text details
  status [flags]                  Print summary job history
  oauth list|doctor|init          Setup CLI-login model providers
  browser doctor|manifest|read    Inspect local web pages through the browser tool gate
  computer doctor|manifest|snapshot Inspect configured desktop accessibility tool gate
  tools manifest                  Print built-in tool/MCP capability manifest
  production-status [flags]       Print local/public production readiness
  production-actions [flags]      Print remaining production action checklist
  production-finalize [flags]     Run guarded final production evidence
  resume <job> --answer <text>    Resume a needs_input job
  retry <job>                     Rerun a saved job with current config
  rollback <report>               Roll back supported patches from a saved JSON report
  explain-failure <job>           Explain a failed job in plain language

Usage:
  ceo-packet start <path>
  ceo-packet run [flags] <task>
  ceo-packet gauntlet [flags]
  ceo-packet gauntlet --suite production-core [flags]
  ceo-packet gauntlet --suite production-core --concurrency 4 [flags]
  ceo-packet doctor [flags]
  ceo-packet inbox [flags]
  ceo-packet status [flags]
  ceo-packet oauth list
  ceo-packet oauth doctor [provider]
  ceo-packet oauth init <provider> --workspace <path>
  ceo-packet browser doctor
  ceo-packet browser read http://localhost:3000 --format text
  ceo-packet computer doctor
  ceo-packet tools manifest
  ceo-packet production-status [flags]
  ceo-packet production-actions [flags]
  ceo-packet production-finalize --dry-run [flags]
  ceo-packet resume <job> --answer <text>
  ceo-packet retry <job>
  ceo-packet rollback <report>
  ceo-packet explain-failure <job>
  ceo-packet [flags] <task>

Common flags:
  --workspace <path>              Use a workspace directory
  --quickstart <path>             Create example config and run doctor
  --start <path>                  Guided setup/check/doctor start flow
  --doctor                       Run harness health checks
  --demo                         Run the built-in golden coding demo
  --inbox                        Review queue alias with text details
  --init-demo-repo <path>         Create a tiny golden demo repo
  --check <command...> --         Add a verification command
  --plan-only                     Preview packet/routes/checks without model calls
  --write-policy <policy>         observe, preview, dry-run, approved-write, or trusted-local
  --dry-run                       Preview patch writes without workspace artifacts/history; write intent previews by default
  --provider-wizard <preset>      Create a main provider for openai, openrouter, kimi-code, moonshot, or minimax
  --adapter <name>                Use external worker adapter: codex, kimi, claude, opencode, aider, goose
  OAuth CLI: ceo-packet oauth init kimi --workspace .
  --format <json|text>            Print JSON or compact text
  --browser-policy <policy>        deny, ask, allow-localhost, or allow
  --computer-policy <policy>       deny, ask, or allow
  --version                       Print version
  --help, -h                      Print this compact help
  --help-advanced                 Print all commands, provider flags, model flags, and history tools

Recommended first run:
  1. ceo-packet oauth doctor --format text
  2. ceo-packet oauth init kimi --workspace . --format text
  3. ceo-packet run --workspace . --check go test ./... -- "Fix one real task"
  4. ceo-packet production-status --workspace . --format text

Examples:
  ceo-packet oauth doctor --format text
  ceo-packet oauth init kimi --workspace . --format text
  ceo-packet tools manifest --format json
  ceo-packet browser read http://localhost:3000 --format text
  ceo-packet run --workspace . --check go test ./... -- "Fix one real task"
  ceo-packet production-status --workspace . --format text
  ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-gauntlet
  ceo-packet production-actions --workspace . --format text
  ceo-packet production-actions --workspace . --format text --action-kind provider_proof
  ceo-packet production-actions --workspace . --format text --env-ready-only
  ceo-packet production-finalize --workspace . --dry-run
  ceo-packet retry latest --workspace .
  ceo-packet rollback .ceo-harness/history/job-000001.json --workspace .
  ceo-packet config check --workspace .
  ceo-packet --init-demo-repo /tmp/ceo-demo

Advanced: run ceo-packet --help-advanced for the full reference.
`

func runHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, helpText)
	return err
}

func runAdvancedHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, advancedHelpText)
	return err
}
