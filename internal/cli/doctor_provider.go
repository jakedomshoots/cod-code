package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"ceoharness/internal/config"
	"ceoharness/internal/subagent"
)

func runNamedProviderDoctor(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildNamedProviderDoctorReport(ctx, opts)
	if writeErr := writeDoctorReport(out, report, opts.reportFormat); writeErr != nil {
		return writeErr
	}
	if report.Status != "pass" {
		if err != nil {
			return err
		}
		return ErrVerdictFailed
	}
	return nil
}

func buildNamedProviderDoctorReport(ctx context.Context, opts options) (doctorReport, error) {
	providerName := strings.TrimSpace(opts.doctorProviderName)
	checkName := "provider." + providerName
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		check := failedDoctorCheck(checkName, err)
		check.Source = "workspace"
		return providerDoctorReport("fail", check), err
	}
	provider, ok := cfg.Providers[providerName]
	if !ok {
		err := fmt.Errorf("provider %q: %w", providerName, config.ErrInvalidConfig)
		check := failedDoctorCheck(checkName, err)
		check.Source = "workspace"
		return providerDoctorReport("fail", check), err
	}
	preflight := providerPreflightDoctorCheck(providerName, provider, true, exec.LookPath)
	if preflight.Status != doctorStatusPass {
		return providerDoctorReport("fail", preflight), ErrVerdictFailed
	}
	check := runProviderDoctorCheck(ctx, providerName, provider, cfg.ModelCommandTimeoutMS)
	check.Requirement = doctorRequired
	if check.Status != "pass" {
		return providerDoctorReport("fail", check), ErrVerdictFailed
	}
	return providerDoctorReport("pass", check), nil
}

func providerDoctorReport(status string, check doctorCheck) doctorReport {
	return doctorReport{
		Status:  status,
		Summary: "provider doctor " + status,
		Version: versionDetails(),
		Checks:  []doctorCheck{check},
	}
}

func runProviderDoctorChecks(ctx context.Context, opts options) []doctorCheck {
	cfg, err := config.LoadWorkspace(ctx, opts.workspaceDir)
	if err != nil {
		return []doctorCheck{failedDoctorCheck("provider", err)}
	}
	if len(cfg.Providers) == 0 {
		return nil
	}
	requiredProviders := requiredDoctorProviders(cfg)
	checks := make([]doctorCheck, 0, len(cfg.Providers))
	for _, providerName := range sortedDoctorProviderNames(cfg.Providers) {
		required := requiredProviders[providerName]
		preflight := providerPreflightDoctorCheck(providerName, cfg.Providers[providerName], required, exec.LookPath)
		if preflight.Status != doctorStatusPass {
			checks = append(checks, preflight)
			continue
		}
		check := runProviderDoctorCheck(ctx, providerName, cfg.Providers[providerName], cfg.ModelCommandTimeoutMS)
		check.Requirement = providerRequirement(required)
		checks = append(checks, check)
	}
	return checks
}

func runProviderDoctorCheck(ctx context.Context, providerName string, provider config.Provider, timeoutMS int) doctorCheck {
	checkName := "provider." + providerName
	client, err := clientForProvider(provider, timeoutMS)
	if err != nil {
		check := failedDoctorCheck(checkName, err)
		check.Source = "workspace"
		return check
	}
	result, err := subagent.NewRunnerWithModel(client).Run(ctx, subagent.TaskPacket{
		Task:            "Doctor provider check",
		AgentName:       "doctor",
		Role:            "provider health",
		ProviderName:    providerName,
		ContextMode:     "lean",
		MaxContextBytes: 1024,
	})
	if err != nil {
		check := failedDoctorCheck(checkName, err)
		check.Source = "workspace"
		return check
	}
	check := doctorCheck{Name: checkName, Status: result.Status, Source: "workspace", Verdict: result.Status}
	if result.Status != "pass" {
		check.Status = "fail"
		check.Error = result.Summary
	}
	return check
}

func requiredDoctorProviders(cfg config.Config) map[string]bool {
	required := map[string]bool{}
	addRequiredProvider(required, cfg.CEOProvider)
	for _, providerName := range cfg.AgentProviders {
		addRequiredProvider(required, providerName)
	}
	addRequiredProvider(required, cfg.ProviderPolicy.DefaultProvider)
	addRequiredProvider(required, cfg.ProviderPolicy.FallbackProvider)
	for _, providerName := range cfg.ProviderPolicy.RiskProviders {
		addRequiredProvider(required, providerName)
	}
	for _, providerName := range cfg.ProviderPolicy.KindProviders {
		addRequiredProvider(required, providerName)
	}
	for _, providerName := range cfg.ProviderPolicy.RiskAreaProviders {
		addRequiredProvider(required, providerName)
	}
	return required
}

func addRequiredProvider(required map[string]bool, providerName string) {
	name := strings.TrimSpace(providerName)
	if name != "" {
		required[name] = true
	}
}

func providerPreflightDoctorCheck(providerName string, provider config.Provider, required bool, lookup doctorLookupFunc) doctorCheck {
	check := doctorCheck{
		Name:        "provider." + providerName,
		Status:      doctorStatusPass,
		Requirement: providerRequirement(required),
		Source:      "workspace",
	}
	if missing := missingProviderEnvVar(provider); missing != "" {
		check.Status = missingProviderStatus(required)
		check.Error = "provider env var " + missing + " is required"
		check.Guidance = "export " + missing + "=..."
		return check
	}
	if missing := missingProviderCommand(provider, lookup); missing != "" {
		check.Status = missingProviderStatus(required)
		check.Error = missing + " not found"
		check.Guidance = "install or reconfigure provider command"
		return check
	}
	return check
}

func missingProviderEnvVar(provider config.Provider) string {
	for _, rawName := range providerEnvVars(provider) {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		value, ok := os.LookupEnv(name)
		if !ok || value == "" {
			return name
		}
	}
	return ""
}

func missingProviderCommand(provider config.Provider, lookup doctorLookupFunc) string {
	if len(provider.ModelCommand) == 0 {
		return ""
	}
	command := strings.TrimSpace(provider.ModelCommand[0])
	if command == "" {
		return "empty provider command"
	}
	if strings.Contains(command, string(os.PathSeparator)) {
		info, err := os.Stat(command)
		if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
			return command
		}
		return ""
	}
	if _, err := lookup(command); err != nil {
		return command
	}
	return ""
}

func missingProviderStatus(required bool) string {
	if required {
		return doctorStatusBlocked
	}
	return doctorStatusSkipped
}

func providerRequirement(required bool) string {
	if required {
		return doctorRequired
	}
	return doctorOptional
}

func sortedDoctorProviderNames(providers map[string]config.Provider) []string {
	names := make([]string, 0, len(providers))
	for rawName := range providers {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
