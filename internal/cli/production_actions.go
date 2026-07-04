package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type productionActionsReport struct {
	Path                string           `json:"path"`
	Status              string           `json:"status"`
	RequiredActionCount int              `json:"required_action_count"`
	Actions             []map[string]any `json:"actions"`
}

func runProductionActions(out io.Writer, opts options) error {
	report, err := buildProductionActionsReport(opts.workspaceDir)
	if err != nil {
		return err
	}
	return writeProductionActionsReport(out, report, opts.reportFormat)
}

func buildProductionActionsReport(workspaceDir string) (productionActionsReport, error) {
	status, err := buildProductionStatusReport(workspaceDir)
	if err != nil {
		return productionActionsReport{}, err
	}
	if status.FinalizerNextActions == nil || status.FinalizerNextActions.JSONPath == "" {
		return productionActionsReport{
			Status: "missing",
			Path:   "",
		}, nil
	}
	content, err := os.ReadFile(status.FinalizerNextActions.JSONPath)
	if err != nil {
		return productionActionsReport{}, fmt.Errorf("read production actions: %w", err)
	}
	var raw struct {
		Status              string           `json:"status"`
		RequiredActionCount int              `json:"required_action_count"`
		Actions             []map[string]any `json:"actions"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return productionActionsReport{}, fmt.Errorf("decode production actions: %w", err)
	}
	return productionActionsReport{
		Path:                status.FinalizerNextActions.JSONPath,
		Status:              raw.Status,
		RequiredActionCount: raw.RequiredActionCount,
		Actions:             raw.Actions,
	}, nil
}

func writeProductionActionsReport(out io.Writer, report productionActionsReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := fmt.Fprint(out, renderProductionActionsText(report))
		return err
	case reportFormatEvents:
		encoder := json.NewEncoder(out)
		return encoder.Encode(struct {
			Kind    string                  `json:"kind"`
			Actions productionActionsReport `json:"actions"`
		}{Kind: "production_actions", Actions: report})
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderProductionActionsText(report productionActionsReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Production actions: %s\n", report.Status)
	fmt.Fprintf(&builder, "Required actions: %d\n", report.RequiredActionCount)
	if report.Path != "" {
		fmt.Fprintf(&builder, "Source: %s\n", report.Path)
	}
	for _, action := range report.Actions {
		id, _ := action["id"].(string)
		kind, _ := action["kind"].(string)
		text, _ := action["text"].(string)
		if kind != "" {
			fmt.Fprintf(&builder, "- %s [%s]: %s\n", id, kind, text)
		} else {
			fmt.Fprintf(&builder, "- %s: %s\n", id, text)
		}
	}
	return builder.String()
}
