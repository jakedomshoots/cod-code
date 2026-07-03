package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func writeQuickstartReport(out io.Writer, report quickstartReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write quickstart report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderQuickstartText(report)); err != nil {
			return fmt.Errorf("write text quickstart report: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderQuickstartText(report quickstartReport) string {
	var builder strings.Builder
	builder.WriteString("Quickstart: ")
	builder.WriteString(report.Status)
	builder.WriteString("\nWorkspace: ")
	builder.WriteString(report.Workspace)
	builder.WriteString("\nConfig: ")
	builder.WriteString(report.ConfigInit.ConfigPath)
	builder.WriteString("\nDoctor: ")
	builder.WriteString(report.Doctor.Status)
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

func quickstartNextSteps(workspaceDir string, providerSteps []string) []string {
	workspace := workspaceArg(workspaceDir)
	doctorStep := "ceo-packet --workspace " + workspace + " --doctor --format text"
	smokeStep := "ceo-packet --workspace " + workspace + " --plan-only " + strconv.Quote("Smoke provider routing")
	steps := []string{"ceo-packet --workspace " + workspace + " --config-check --format text"}
	seen := map[string]struct{}{steps[0]: {}}
	for _, step := range providerSteps {
		if step == smokeStep {
			continue
		}
		if _, ok := seen[step]; ok {
			continue
		}
		seen[step] = struct{}{}
		steps = append(steps, step)
	}
	steps = append(steps, doctorStep, smokeStep)
	return steps
}
