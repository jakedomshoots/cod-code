package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

func runConfigDoctor(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildConfigCheckReport(ctx, opts)
	if err != nil {
		return err
	}
	return writeConfigDoctorReport(out, report, opts.reportFormat)
}

func writeConfigDoctorReport(out io.Writer, report configCheckReport, format reportFormat) error {
	if format != reportFormatText {
		return writeConfigCheckReport(out, report, format)
	}
	if _, err := io.WriteString(out, renderConfigDoctorText(report)); err != nil {
		return fmt.Errorf("write text config doctor report: %w", err)
	}
	return nil
}

func renderConfigDoctorText(report configCheckReport) string {
	var builder strings.Builder
	builder.WriteString("Config doctor: ")
	builder.WriteString(configDoctorStatus(report))
	builder.WriteString("\nConfig: ")
	builder.WriteString(report.ConfigPath)
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Providers: %d (%d HTTP)\n", report.ProviderCount, report.ProviderHTTPCount))
	builder.WriteString(fmt.Sprintf("Provider env: %d/%d set", report.ProviderEnvVarPresentCount, report.ProviderEnvVarCount))
	if report.ProviderEnvVarMissingCount > 0 {
		builder.WriteString(" missing=")
		builder.WriteString(strings.Join(report.ProviderEnvVarMissingNames, ","))
	}
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Checks: configured=%v required=%v\n", report.CheckCommandPresent, report.RequireChecks))
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

func configDoctorStatus(report configCheckReport) string {
	if report.ProviderEnvVarMissingCount > 0 {
		return "needs setup"
	}
	if report.RequireChecks && !report.CheckCommandPresent {
		return "needs setup"
	}
	return "pass"
}
