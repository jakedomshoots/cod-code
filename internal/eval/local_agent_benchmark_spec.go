package eval

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"ceoharness/internal/config"
)

const (
	ceoBenchmarkModeSynthetic    = "synthetic"
	ceoBenchmarkModeModelCommand = "model-command"
	ceoBenchmarkModeHTTPProvider = "http-provider"
)

func buildLocalAgentBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, task Task) (localAgentSpec, error) {
	prompt := localAgentBenchmarkPrompt(task)
	switch id {
	case "ceo_harness":
		return buildCEOBenchmarkSpec(id, req, task, prompt)
	case "codex_cli":
		args := []string{"exec", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--sandbox", "workspace-write", "--skip-git-repo-check"}
		args = appendAgentModelArgs(args, req.AgentModels, id)
		args = append(args, prompt)
		return localAgentSpec{
			id:        id,
			name:      "OpenAI Codex CLI",
			binary:    "codex",
			args:      args,
			setupHint: "Install and authenticate Codex CLI before benchmark runs.",
		}, nil
	case "opencode":
		args := []string{"run", "--print-logs", "--log-level", "INFO", "--pure", "--auto", "--format", "json"}
		args = appendAgentModelArgs(args, req.AgentModels, id)
		args = append(args, prompt)
		return localAgentSpec{
			id:        id,
			name:      "OpenCode",
			binary:    "opencode",
			args:      args,
			setupHint: "Install and authenticate OpenCode before benchmark runs.",
		}, nil
	case "pi":
		args := []string{"--no-session", "--approve"}
		args = appendAgentModelArgs(args, req.AgentModels, id)
		args = append(args, "-p", prompt)
		return localAgentSpec{
			id:        id,
			name:      "Pi",
			binary:    "pi",
			args:      args,
			setupHint: "Install Pi and configure a provider key before benchmark runs.",
		}, nil
	default:
		return localAgentSpec{}, fmt.Errorf("%w: unknown local agent %q", ErrInvalidCompetitor, id)
	}
}

func appendAgentModelArgs(args []string, models map[string]string, agentID string) []string {
	if len(models) == 0 {
		return args
	}
	model := strings.TrimSpace(models[agentID])
	if model == "" {
		return args
	}
	return append(args, "--model", model)
}

func buildCEOBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, task Task, prompt string) (localAgentSpec, error) {
	binary := strings.TrimSpace(req.CEOHarnessBinary)
	if binary == "" {
		binary = filepath.Join(".", "bin", "ceo-packet")
	}
	mode := strings.TrimSpace(req.CEOBenchmarkMode)
	if mode == "" {
		mode = ceoBenchmarkModeSynthetic
	}
	if mode == ceoBenchmarkModeModelCommand {
		if len(req.CEOBenchmarkModelCommand) == 0 {
			return localAgentSpec{}, fmt.Errorf("%w: --ceo-benchmark-mode model-command requires --ceo-benchmark-model-command-json", ErrInvalidCompetitor)
		}
		args := []string{"--write-policy", "trusted-local", "--apply-model-patches", "--check-fix-attempts", "2", "--subagent-attempts", "2", "--no-progress-stop", "2", "--format", "json", "--model-command"}
		args = append(args, req.CEOBenchmarkModelCommand...)
		args = append(args, "--")
		args = append(args, "--ceo-model-command")
		args = append(args, req.CEOBenchmarkModelCommand...)
		args = append(args, "--")
		args = appendCEORequiredCheckArgs(args, task)
		args = append(args, prompt)
		return localAgentSpec{
			id:        id,
			name:      "CEO Harness",
			binary:    binary,
			args:      args,
			setupHint: "Build CEO Harness with `make build` before benchmark runs.",
		}, nil
	}
	if mode == ceoBenchmarkModeHTTPProvider {
		workspaceConfig, err := ceoBenchmarkHTTPProviderConfig(req)
		if err != nil {
			return localAgentSpec{}, err
		}
		args := appendCEORequiredCheckArgs([]string{"--write-policy", "trusted-local", "--apply-model-patches", "--check-fix-attempts", "2", "--max-subagents", "1", "--subagent-attempts", "2", "--no-progress-stop", "2", "--format", "json"}, task, prompt)
		return localAgentSpec{
			id:              id,
			name:            "CEO Harness",
			binary:          binary,
			args:            args,
			workspaceConfig: workspaceConfig,
			setupHint:       "Build CEO Harness with `make build` and configure the provider API key before benchmark runs.",
		}, nil
	}
	if mode != ceoBenchmarkModeSynthetic {
		return localAgentSpec{}, fmt.Errorf("%w: unknown CEO benchmark mode %q", ErrInvalidCompetitor, mode)
	}
	args := appendCEOBenchmarkSyntheticReplaceArgs([]string{"--write-policy", "trusted-local"}, task)
	args = append(args, "--format", "json")
	args = appendCEORequiredCheckArgs(args, task, prompt)
	return localAgentSpec{
		id:                       id,
		name:                     "CEO Harness",
		binary:                   binary,
		args:                     args,
		benchmarkWritesArtifacts: true,
		setupHint:                "Build CEO Harness with `make build` before benchmark runs.",
	}, nil
}

func appendCEOBenchmarkSyntheticReplaceArgs(args []string, task Task) []string {
	targetPaths := task.RequiredChangedFiles
	if len(targetPaths) == 0 {
		targetPaths = []string{"README.md"}
	}
	for _, targetPath := range targetPaths {
		args = append(args,
			"--replace",
			targetPath,
			benchmarkBaselineText(task, targetPath),
			benchmarkExpectedText(task, targetPath),
		)
	}
	return args
}

func ceoBenchmarkHTTPProviderConfig(req LocalAgentBenchmarkRequest) ([]byte, error) {
	name := strings.TrimSpace(req.CEOBenchmarkProviderName)
	if name == "" {
		name = "main"
	}
	model := strings.TrimSpace(req.CEOBenchmarkProviderModel)
	if model == "" {
		return nil, fmt.Errorf("%w: --ceo-benchmark-mode http-provider requires --ceo-benchmark-provider-model", ErrInvalidCompetitor)
	}
	preset, err := ceoBenchmarkHTTPProviderPreset(req.CEOBenchmarkProviderPreset)
	if err != nil {
		return nil, err
	}
	apiKeyEnv := strings.TrimSpace(req.CEOBenchmarkProviderAPIKeyEnv)
	if apiKeyEnv == "" {
		apiKeyEnv = preset.apiKeyEnv
	}
	responseFormat := preset.responseFormat
	if responseFormat == "" && !preset.allowUnstructured {
		responseFormat = "json_object"
	}
	cfg := config.Config{
		Providers: map[string]config.Provider{
			name: {
				HTTP: config.HTTPProvider{
					URL:             preset.url,
					Model:           model,
					APIKeyEnv:       apiKeyEnv,
					MaxOutputTokens: req.CEOBenchmarkProviderMaxOutputToks,
					ResponseFormat:  responseFormat,
					DisableThinking: preset.disableThinking,
				},
			},
		},
		CEOProvider: name,
		ProviderPolicy: config.ProviderPolicy{
			DefaultProvider: name,
		},
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("build benchmark provider config: %w", err)
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode benchmark provider config: %w", err)
	}
	return append(payload, '\n'), nil
}

type ceoBenchmarkProviderPreset struct {
	url               string
	apiKeyEnv         string
	responseFormat    string
	allowUnstructured bool
	disableThinking   bool
}

func ceoBenchmarkHTTPProviderPreset(raw string) (ceoBenchmarkProviderPreset, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "openrouter":
		return ceoBenchmarkProviderPreset{
			url:       "https://openrouter.ai/api/v1/chat/completions",
			apiKeyEnv: "OPENROUTER_API_KEY",
		}, nil
	case "openai":
		return ceoBenchmarkProviderPreset{
			url:       "https://api.openai.com/v1/chat/completions",
			apiKeyEnv: "OPENAI_API_KEY",
		}, nil
	case "kimi", "kimi-code", "kimicode":
		return ceoBenchmarkProviderPreset{
			url:               "https://api.kimi.com/coding/v1/chat/completions",
			apiKeyEnv:         "KIMI_CODE_API_KEY",
			allowUnstructured: true,
		}, nil
	case "moonshot":
		return ceoBenchmarkProviderPreset{
			url:       "https://api.moonshot.ai/v1/chat/completions",
			apiKeyEnv: "MOONSHOT_API_KEY",
		}, nil
	case "minimax":
		return ceoBenchmarkProviderPreset{
			url:             "https://api.minimax.io/v1/chat/completions",
			apiKeyEnv:       "MINIMAX_API_KEY",
			disableThinking: true,
		}, nil
	default:
		return ceoBenchmarkProviderPreset{}, fmt.Errorf("%w: unknown --ceo-benchmark-provider-preset %q", ErrInvalidCompetitor, raw)
	}
}
