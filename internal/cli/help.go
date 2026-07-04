package cli

import (
	"fmt"
	"io"
)

const helpText = `ceo-packet

Primary flow:
  start <path>                    Guided setup/check/doctor start flow
  run [flags] <task>              Run the CEO packet loop
  gauntlet [flags]                Run market or production benchmark gauntlets
  doctor [flags]                  Run harness health checks
  inbox [flags]                   Review queue alias with text details
  status [flags]                  Print summary job history
  oauth list|doctor|init          Setup CLI-login model providers
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
  --version                       Print version
  --help, -h                      Print this compact help
  --help-advanced                 Print all commands, provider flags, model flags, and history tools

Provider quick start:
  OAuth providers: ceo-packet oauth list
  Kimi OAuth: ceo-packet oauth init kimi --workspace .
  Codex OAuth: ceo-packet oauth init codex --workspace .
  Claude OAuth: ceo-packet oauth init claude --workspace .
  OpenCode OAuth: ceo-packet oauth init opencode --workspace .
  Goose OAuth: ceo-packet oauth init goose --workspace .
  Codex CLI: ceo-packet config init --adapter codex
  Kimi CLI: ceo-packet config init --adapter kimi
  OpenRouter: use --provider-wizard openrouter; missing OPENROUTER_API_KEY is blocked setup, not a failed benchmark
  Kimi Code API: use --provider-wizard kimi-code with KIMI_CODE_API_KEY, or use --adapter kimi for OAuth CLI
  MiniMax API: use --provider-wizard minimax with MINIMAX_API_KEY

Examples:
  ceo-packet start .
  ceo-packet run --workspace . --check go test ./... -- "Fix retry bug"
  ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-gauntlet
  ceo-packet inbox --workspace .
  ceo-packet production-status --workspace .
  ceo-packet production-actions --workspace . --format text
  ceo-packet production-actions --workspace . --format text --action-kind provider_proof
  ceo-packet production-actions --workspace . --format text --env-ready-only
  ceo-packet production-finalize --workspace . --dry-run
  ceo-packet oauth doctor --format text
  ceo-packet retry latest --workspace .
  ceo-packet rollback .ceo-harness/history/job-000001.json --workspace .
  ceo-packet config check --workspace .
  ceo-packet --init-demo-repo /tmp/ceo-demo

Advanced: run ceo-packet --help-advanced for the full reference.
`

const advancedHelpText = `ceo-packet

Primary flow:
  start <path>                    Guided setup/check/doctor start flow
  run [flags] <task>              Run the CEO packet loop
  gauntlet [flags]                Run market or production benchmark gauntlets
  doctor [flags]                  Run harness health checks
  inbox [flags]                   Review queue alias with text details
  status [flags]                  Print summary job history
  oauth list|doctor|init          Setup CLI-login model providers
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
  ceo-packet production-status [flags]
  ceo-packet production-actions [flags]
  ceo-packet production-finalize --dry-run [flags]
  ceo-packet resume <job> --answer <text>
  ceo-packet retry <job>
  ceo-packet rollback <report>
  ceo-packet explain-failure <job>
  ceo-packet [flags] <task>

Advanced commands:
  review [flags]                  Print review queue with details
  context [flags] <job>            Print per-agent context trace
  oauth list|doctor|init          Setup CLI-login model providers
  config check|doctor|explain     Print config health and first-run help
  config completions [flags]      Write zsh, bash, or fish completion file
  config init [flags]             Create workspace config
  eval [flags]                    Run eval catalog, rubric, scoring, and benchmark tools
  tui [flags]                     Open stdin-driven operator dashboard

Common flags:
	  --workspace <path>              Use a workspace directory
	  --artifact-root <path>          Write runtime artifacts/history outside the workspace
	  --quickstart <path>             Create example config and run doctor
	  --start <path>                  Guided setup/check/doctor start flow
	  --doctor                       Run harness health checks
	  --demo                         Run the built-in golden coding demo
	  --init-demo-repo <path>         Create a tiny golden demo repo
	  --plan-only                     Preview packet/routes/checks without model calls
	  --write-policy <policy>         observe, preview, dry-run, approved-write, or trusted-local
	  --dry-run                       Preview patch writes without workspace artifacts/history; write intent previews by default
	  --job-timeout-ms <n>            Timeout the whole job run
	  --check <command...> --         Add a verification command
	  --require-checks                Refuse run/plan when no check is configured
	  --check-set <name>              Run a named check set from config
	  --check-attempts <n>            Retry failing checks before verdict
	  --check-backoff-ms <n>          Wait between check retries
	  --tool-command-timeout-ms <n>   Timeout check/research commands
	  --repair-preset <standard>      Set standard repair attempts unless explicit flags override
	  --check-fix-attempts <n>        Let coder patch after failed checks
	  --ceo-revision-attempts <n>     Let CEO send failed review feedback to coder
	  --max-ceo-iterations <n>        Cap initial run plus CEO correction passes
	  --max-subagents <n>             Override the default 3-agent delegation budget
	  --subagent-concurrency <n>      Cap parallel subagents per stage
	  --subagent-attempts <n>         Retry failing subagent calls
	  --subagent-backoff-ms <n>       Wait between subagent retries
	  --max-tool-requests <n>         Cap tool requests per subagent
	  --no-progress-stop <n>          Stop repeated weak subagent attempts
	  --replace <path> <old> <new>    Apply a bounded text patch
	  --approve-preview <digest>      Require a matching dry-run patch digest
	  --rollback-report <path>        Roll back supported patches from a saved JSON report
	  --apply-model-patches           Apply coder JSON patch proposals
	  --preview-model-patches         Preview coder JSON patch proposals
	  --max-model-patches <n>         Cap coder JSON patches per proposal
	  --max-context-bytes <n>         Cap lean task-packet context bytes
	  --max-subagent-output-bytes <n> Cap saved subagent summary/evidence text
	  --min-subagent-confidence <n>   Fail or fallback below confidence 0..1
	  --workspace-brief-max-files <n> Cap workspace brief file count
	  --workspace-brief-exclude <glob> Omit path/glob from workspace brief
	  --format <json|text|events>     Print JSON, compact text, or JSONL events
	  --interactive                   Ask and resume when subagents need input
	  --tui                           Open stdin-driven operator dashboard
	  --snapshot                      Print deterministic TUI/dashboard text for CI
	  --with-job-context <id>         Add compact previous-job context to task
	  --version                       Print version

Advanced flags:
  --init-config                   Create workspace config
  --init-example-adapters         Use bundled example model/CEO/research commands
  --adapter <name>                Use external worker adapter: codex, kimi, claude, opencode, aider, goose
  --config-check                  Print config health summary
  --config-doctor                 Print compact config doctor report
  --config-explain                Print compact first-run checklist
  --config-completions            Write a shell completion file
  --shell <zsh|bash|fish>         Completion shell
  --output-dir <path>             Completion output directory
  --history                       Print recent job history
  --production-status             Print latest production-readiness status from evidence
  --production-actions            Print latest production finalizer action checklist
  --action-id <id>                Filter production-actions by exact action id
  --action-kind <kind>            Filter production-actions by kind
  --action-provider <name>        Filter production-actions by provider
  --action-state <state>          Filter production-actions by ready, missing_env, empty_env, setup_blocked, or waiting
  --env-ready-only                Show production actions whose required env is present
  --ready-only                    Show production actions with env and dependencies ready
  --next                          Show the first ready production action after filters
  --commands-only                 Print a paste-safe production action command script
  --production-finalize           Run guarded final production evidence sequence
  --run-comparison                Include the expensive 29-task all-agent comparison
  --review-queue                  Print jobs needing human attention
  --review-details                Include compact context in review queue
  --inbox                         Review queue alias with text details
  --snapshot                      Print deterministic TUI/dashboard text for CI
  --provider-health               Summarize provider health from history
  --doctor-provider <name>         Run one provider health check
  --provider <name>               Filter provider-health output
  --ceo-provider <name>           Route CEO delegation/review to provider
  --default-provider <name>       Route unassigned subagents to provider
  --fallback-provider <name>      Fallback provider after route failure
  --risk-provider <r=p>           Route low/medium/high task risk to provider
  --kind-provider <k=p>           Route coding/planning/research/mixed to provider
  --risk-area-provider <a=p>      Route risk specialist area to provider
  --recommendation <label>        Filter provider-health by recommendation
  --task <text>                   Filter history/provider-health by task text
  --verdict <value>               Filter history by verdict
  --limit <n>                     Limit history rows
  --since <time>                  Filter history from timestamp
  --until <time>                  Filter history until timestamp
  --summary-only                  Print history/provider-health counts without rows
  --top-providers <n>             Limit provider-health to worst N rows
  --job <id>                      Print one history job
  --job-context <id>              Print compact resume context for one job
  --context-trace <id>             Print per-agent packet budget trace
  --job-report <id>               Print the saved full CEO report for one job
  --job-events <id>               Print saved run events as JSONL
  --judge-job <id>                Read or write a human final judgment
  --human-verdict <accept|reject> Human judgment to write for --judge-job
  --judgment-note <text>          Optional human judgment note
  --rerun <id>                    Rerun a history job task with current config
  --continue-job <id>             Reuse passed subagents from a saved job
  --resume <id>                   Resume a needs_input job with --answer
  --answer <text>                 Answer a resumed job question
  job id aliases: latest, last
  --help, -h                      Print this help

Provider quick start:
  OAuth providers: ceo-packet oauth list
  Kimi OAuth: ceo-packet oauth init kimi --workspace .
  Codex OAuth: ceo-packet oauth init codex --workspace .
  Claude OAuth: ceo-packet oauth init claude --workspace .
  OpenCode OAuth: ceo-packet oauth init opencode --workspace .
  Goose OAuth: ceo-packet oauth init goose --workspace .
  Codex CLI: ceo-packet config init --adapter codex
  Kimi CLI: ceo-packet config init --adapter kimi
  OpenRouter: use --provider-wizard openrouter; missing OPENROUTER_API_KEY is blocked setup, not a failed benchmark
  Kimi Code API: use --provider-wizard kimi-code with KIMI_CODE_API_KEY, or use --adapter kimi for OAuth CLI
  MiniMax API: use --provider-wizard minimax with MINIMAX_API_KEY

Model:
	  --model-command <command...> -- Run a local model command
	  --ceo-model-command <command...> -- Run final CEO model review
	  --research-command <command...> -- Run network_research tool requests
	  --model-command-timeout-ms <n> Timeout local model commands
	  env: CEO_MODEL_COMMAND_JSON, CEO_REVIEW_MODEL_COMMAND_JSON, CEO_RESEARCH_COMMAND_JSON
	  config: model_command, ceo_model_command, research_command, model_command_timeout_ms, tool_command_timeout_ms
	  coder patch JSON: {"patches":[{"path":"app.txt","old":"old","new":"new"}]}
	  create patch JSON: {"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}

OAuth CLI setup:
  ceo-packet oauth list
  ceo-packet oauth doctor --format text
  ceo-packet oauth init kimi --workspace . --format text
  ceo-packet oauth init codex --workspace . --format text
  ceo-packet oauth init claude --workspace . --format text
  ceo-packet oauth init opencode --workspace . --format text
  ceo-packet oauth init goose --workspace . --format text
  Built-in OAuth init stores no tokens. It creates a command provider that uses the local CLI login.

HTTP provider setup:
  --provider-wizard <preset>      Create a main provider for openai, openrouter, kimi-code, moonshot, or minimax
  --http-provider <name>          Provider name for --init-config
  --http-preset <name>            openai, openrouter, kimi-code, moonshot, or minimax
  --http-url <url>                Chat completions endpoint
  --http-model <model>            Provider model name
  --http-api-key-env <name>       Env var containing API key
  --http-agent <name>             Agent role using the provider
  --http-timeout-ms <n>           Request timeout
  --http-max-output-tokens <n>    Max generated tokens
  --http-response-format <format> text or json_object
  --http-input-cost-per-million <n>  Input token price
  --http-output-cost-per-million <n> Output token price
  --provider-health-avoid-failure-rate <n>  Avoid threshold from 0 to 1
  --provider-health-watch-failure-rate <n>  Watch threshold from 0 to 1
  --provider-health-watch-cost-per-attempt-microusd <n>  Watch cost threshold

Examples:
  ceo-packet "Add tests for checkout"
  ceo-packet doctor
  ceo-packet --demo
  ceo-packet --quickstart .
  ceo-packet start .
  ceo-packet gauntlet --agents ceo_harness --output-dir .omo/evidence/gauntlet
  ceo-packet gauntlet --suite production-core --agents ceo_harness --output-dir .omo/evidence/production-gauntlet
  ceo-packet inbox --workspace .
  ceo-packet retry latest --workspace .
  ceo-packet rollback .ceo-harness/history/job-000001.json --workspace .
  ceo-packet review --workspace .
  ceo-packet config check --workspace .
  ceo-packet config doctor --workspace . --format text
  ceo-packet config explain --workspace . --format text
  ceo-packet config completions --shell zsh --output-dir /tmp/ceo-completions
  ceo-packet production-actions --workspace . --format text
  ceo-packet production-actions --workspace . --action-state empty_env --commands-only
  ceo-packet oauth list
  ceo-packet oauth init kimi --workspace . --format text
  ceo-packet --workspace . --provider-wizard openai --http-model gpt-5
  ceo-packet --init-demo-repo /tmp/ceo-demo
  ceo-packet run --workspace . --check go test ./... -- "Fix retry bug"
  ceo-packet config init --workspace . --model-command llm run --
`

func runHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, helpText)
	return err
}

func runAdvancedHelp(out io.Writer) error {
	_, err := fmt.Fprint(out, advancedHelpText)
	return err
}
