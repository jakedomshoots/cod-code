package eval

import (
	"flag"
	"path/filepath"
)

func registerLocalAgentFlags(flags *flag.FlagSet, opts *cliOptions) {
	flags.BoolVar(&opts.localAgentSuite, "local-agent-suite", false, "run safe installed local-agent readiness comparison")
	flags.BoolVar(&opts.localAgentBenchmark, "local-agent-benchmark", false, "run installed local agents against one scored benchmark task")
	flags.StringVar(&opts.localAgents, "local-agents", "", "comma-separated local agents: ceo_harness,codex_cli,claude_code,aider,opencode,goose,pi,oh_my_pi")
	flags.StringVar(&opts.localAgentTask, "local-agent-task", "readiness", "local-agent suite task: readiness or edit-file")
	flags.StringVar(&opts.localAgentBenchmarkTask, "local-agent-benchmark-task", defaultLocalAgentBenchmarkID, "benchmark task id for --local-agent-benchmark")
	flags.IntVar(&opts.localAgentBenchmarkRepeat, "local-agent-benchmark-repeat", 1, "repeat count for each local-agent benchmark task")
	flags.IntVar(&opts.localAgentBenchmarkConcurrency, "local-agent-benchmark-concurrency", 1, "max parallel local-agent benchmark runs")
	flags.IntVar(&opts.localAgentBenchmarkTimeoutRetries, "local-agent-benchmark-timeout-retries", 0, "retry timed-out local-agent benchmark runs this many times")
	flags.IntVar(&opts.localAgentBenchmarkResultRetries, "local-agent-benchmark-result-retries", 0, "retry partial or failed local-agent benchmark runs this many times")
	flags.StringVar(&opts.localAgentBenchmarkAgentTimeouts, "local-agent-benchmark-agent-timeouts", "", "comma-separated per-agent timeout overrides, for example opencode=600,pi=360")
	flags.StringVar(&opts.localAgentBenchmarkAgentModels, "local-agent-benchmark-agent-models", "", "comma-separated per-agent model overrides, for example opencode=openai/gpt-5.4-mini,claude_code=sonnet")
	flags.StringVar(&opts.ceoHarnessBinary, "ceo-binary", filepath.Join(".", "bin", "ceo-packet"), "Cod Code binary for local-agent suite")
	flags.StringVar(&opts.ceoBenchmarkMode, "ceo-benchmark-mode", ceoBenchmarkModeSynthetic, "Cod Code benchmark mode: synthetic, model-command, or http-provider")
	flags.StringVar(&opts.ceoBenchmarkModelCommand, "ceo-benchmark-model-command-json", "", "JSON argv array for Cod Code model-command benchmark mode")
	flags.StringVar(&opts.ceoBenchmarkProviderName, "ceo-benchmark-provider-name", "main", "provider name for Cod Code http-provider benchmark mode")
	flags.StringVar(&opts.ceoBenchmarkProviderPreset, "ceo-benchmark-provider-preset", "openrouter", "provider preset for Cod Code http-provider benchmark mode: openrouter, kimi-code, minimax, openai, or moonshot")
	flags.StringVar(&opts.ceoBenchmarkProviderModel, "ceo-benchmark-provider-model", "", "provider model for Cod Code http-provider benchmark mode")
	flags.StringVar(&opts.ceoBenchmarkProviderAPIKeyEnv, "ceo-benchmark-provider-api-key-env", "", "API key env var for Cod Code http-provider benchmark mode")
	flags.IntVar(&opts.ceoBenchmarkProviderMaxOutputToks, "ceo-benchmark-provider-max-output-tokens", 2048, "max output tokens for Cod Code http-provider benchmark mode")
}
