package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type configExplainReport struct {
	Workspace string   `json:"workspace"`
	Checklist []string `json:"checklist"`
}

func runConfigExplain(out io.Writer, opts options) error {
	report := configExplainReport{
		Workspace: strings.TrimSpace(opts.workspaceDir),
		Checklist: configExplainChecklist(opts.workspaceDir),
	}
	return writeConfigExplainReport(out, report, opts.reportFormat)
}

func writeConfigExplainReport(out io.Writer, report configExplainReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write config explain report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderConfigExplainText(report)); err != nil {
			return fmt.Errorf("write text config explain report: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderConfigExplainText(report configExplainReport) string {
	var builder strings.Builder
	builder.WriteString("First-run checklist\n")
	for _, item := range report.Checklist {
		builder.WriteString("- ")
		builder.WriteString(item)
		builder.WriteString("\n")
	}
	return builder.String()
}

func configExplainChecklist(workspaceDir string) []string {
	workspace := workspaceArg(workspaceDir)
	return []string{
		"provider wizard: cod --workspace " + workspace + " --provider-wizard openai --http-model gpt-5 --format text",
		"dogfood: cod --demo --format text",
		"write policy: start with --write-policy dry-run",
		"first task: cod run --workspace " + workspace + " --check go test ./... -- " + fmt.Sprintf("%q", "Fix one failing test"),
	}
}
