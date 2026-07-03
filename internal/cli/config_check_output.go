package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func writeConfigCheckReport(out io.Writer, report configCheckReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write config check report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderConfigCheckText(report)); err != nil {
			return fmt.Errorf("write text config check report: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderConfigCheckText(report configCheckReport) string {
	var builder strings.Builder
	builder.WriteString("Config: ")
	builder.WriteString(report.ConfigPath)
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Providers: %d (%d HTTP)\n", report.ProviderCount, report.ProviderHTTPCount))
	builder.WriteString(fmt.Sprintf(
		"Provider env: %d/%d set",
		report.ProviderEnvVarPresentCount,
		report.ProviderEnvVarCount,
	))
	if report.ProviderEnvVarMissingCount > 0 {
		builder.WriteString(" missing=")
		builder.WriteString(strings.Join(report.ProviderEnvVarMissingNames, ","))
	}
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Checks: configured=%v required=%v\n", report.CheckCommandPresent, report.RequireChecks))
	if len(report.AdapterCapabilities) > 0 {
		builder.WriteString("Adapters:\n")
		for _, adapterReport := range report.AdapterCapabilities {
			builder.WriteString("- ")
			builder.WriteString(adapterReport.Tool)
			builder.WriteString(": ")
			builder.WriteString(string(adapterReport.Status))
			if adapterReport.ErrorKind != "" {
				builder.WriteString(" ")
				builder.WriteString(adapterReport.ErrorKind)
			}
			builder.WriteString("\n")
		}
	}
	if len(report.ProviderSetupSteps) > 0 {
		builder.WriteString("Next:\n")
		for _, step := range report.ProviderSetupSteps {
			builder.WriteString("- ")
			builder.WriteString(step)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func providerSetupSteps(selection modelCommandSelection, workspaceDir string) []string {
	steps := make([]string, 0, len(selection.providerEnvMissingNames)+len(selection.providerConfigs)+1)
	for _, name := range selection.providerEnvMissingNames {
		steps = append(steps, "export "+name+"=...")
	}
	workspace := workspaceArg(workspaceDir)
	for _, providerName := range sortedDoctorProviderNames(selection.providerConfigs) {
		steps = append(steps, "ceo-packet --workspace "+workspace+" --doctor-provider "+strconv.Quote(providerName)+" --format text")
	}
	if selection.providerCount > 0 {
		steps = append(steps, "ceo-packet --workspace "+workspace+" --plan-only "+strconv.Quote("Smoke provider routing"))
	}
	return steps
}

func workspaceArg(workspaceDir string) string {
	cleanWorkspace := strings.TrimSpace(workspaceDir)
	if cleanWorkspace == "" {
		return strconv.Quote(".")
	}
	return strconv.Quote(cleanWorkspace)
}
