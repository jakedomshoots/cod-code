package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type providerWizardReport struct {
	Preset     string           `json:"preset"`
	Provider   string           `json:"provider"`
	Model      string           `json:"model"`
	ConfigInit configInitReport `json:"config_init"`
	NextSteps  []string         `json:"next_steps,omitempty"`
}

func runProviderWizard(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildProviderWizardReport(ctx, opts)
	if err != nil {
		return err
	}
	return writeProviderWizardReport(out, report, opts.reportFormat)
}

func buildProviderWizardReport(ctx context.Context, opts options) (providerWizardReport, error) {
	preset := strings.ToLower(strings.TrimSpace(opts.providerWizardPreset))
	loose := providerWizardLooseHTTPInit(opts)
	model := strings.TrimSpace(loose.providerModel)
	if model == "" {
		model = providerWizardDefaultModel(preset)
	}
	if _, err := resolveHTTPProviderPreset(preset); err != nil {
		return providerWizardReport{}, err
	}
	opts.initConfig = true
	opts.httpInit = httpInitOptions{}
	opts.httpInits = []httpInitOptions{{
		providerName:         "main",
		presetName:           preset,
		providerModel:        model,
		apiKeyEnv:            loose.apiKeyEnv,
		timeoutMS:            loose.timeoutMS,
		maxOutputTokens:      loose.maxOutputTokens,
		responseFormat:       loose.responseFormat,
		inputCostPerMillion:  loose.inputCostPerMillion,
		outputCostPerMillion: loose.outputCostPerMillion,
	}}
	opts.ceoProvider = "main"
	opts.defaultProvider = "main"
	opts.fallbackProvider = "main"
	configReport, err := buildConfigInitReport(ctx, opts)
	if err != nil {
		return providerWizardReport{}, err
	}
	configCheck, err := buildConfigCheckReport(ctx, opts)
	if err != nil {
		return providerWizardReport{}, err
	}
	return providerWizardReport{
		Preset:     preset,
		Provider:   "main",
		Model:      model,
		ConfigInit: configReport,
		NextSteps:  quickstartNextSteps(opts.workspaceDir, configCheck.ProviderSetupSteps),
	}, nil
}

func providerWizardLooseHTTPInit(opts options) httpInitOptions {
	merged := httpInitOptions{}
	for _, init := range opts.httpInits {
		if strings.TrimSpace(init.providerModel) != "" {
			merged.providerModel = init.providerModel
		}
		if strings.TrimSpace(init.apiKeyEnv) != "" {
			merged.apiKeyEnv = init.apiKeyEnv
		}
		if init.timeoutMS > 0 {
			merged.timeoutMS = init.timeoutMS
		}
		if init.maxOutputTokens > 0 {
			merged.maxOutputTokens = init.maxOutputTokens
		}
		if strings.TrimSpace(init.responseFormat) != "" {
			merged.responseFormat = init.responseFormat
		}
		if init.inputCostPerMillion > 0 {
			merged.inputCostPerMillion = init.inputCostPerMillion
		}
		if init.outputCostPerMillion > 0 {
			merged.outputCostPerMillion = init.outputCostPerMillion
		}
	}
	return merged
}

func providerWizardDefaultModel(preset string) string {
	switch preset {
	case "openai":
		return "gpt-5"
	case "openrouter":
		return "openai/gpt-5"
	case "kimi", "kimi-code", "kimicode":
		return "kimi-for-coding"
	case "moonshot":
		return "moonshot-v1-128k"
	case "minimax":
		return "MiniMax-M3"
	default:
		return ""
	}
}

func writeProviderWizardReport(out io.Writer, report providerWizardReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := io.WriteString(out, renderProviderWizardText(report))
		return err
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderProviderWizardText(report providerWizardReport) string {
	var builder strings.Builder
	builder.WriteString("Provider wizard: ")
	builder.WriteString(report.Preset)
	builder.WriteString("\nConfig: ")
	builder.WriteString(report.ConfigInit.ConfigPath)
	builder.WriteString("\nProvider: ")
	builder.WriteString(report.Provider)
	builder.WriteString("\nModel: ")
	builder.WriteString(report.Model)
	builder.WriteString("\n")
	if len(report.NextSteps) > 0 {
		builder.WriteString("Next:\n")
		for _, step := range report.NextSteps {
			builder.WriteString("- ")
			builder.WriteString(step)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
