package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type productionActionsReport struct {
	Path                string            `json:"path"`
	Status              string            `json:"status"`
	RequiredActionCount int               `json:"required_action_count"`
	Filter              map[string]string `json:"filter,omitempty"`
	Actions             []map[string]any  `json:"actions"`
}

func runProductionActions(out io.Writer, opts options) error {
	report, err := buildProductionActionsReport(opts)
	if err != nil {
		return err
	}
	return writeProductionActionsReport(out, report, opts.reportFormat)
}

func buildProductionActionsReport(opts options) (productionActionsReport, error) {
	status, err := buildProductionStatusReport(opts.workspaceDir)
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
	actions := filterProductionActions(raw.Actions, opts.productionActionKind, opts.productionActionProvider)
	return productionActionsReport{
		Path:                status.FinalizerNextActions.JSONPath,
		Status:              raw.Status,
		RequiredActionCount: len(actions),
		Filter:              productionActionFilter(opts.productionActionKind, opts.productionActionProvider),
		Actions:             actions,
	}, nil
}

func filterProductionActions(actions []map[string]any, kind string, provider string) []map[string]any {
	if kind == "" && provider == "" {
		return actions
	}
	filtered := make([]map[string]any, 0, len(actions))
	for _, action := range actions {
		if kind != "" && actionString(action, "kind") != kind {
			continue
		}
		if provider != "" && actionString(action, "provider") != provider {
			continue
		}
		filtered = append(filtered, action)
	}
	return filtered
}

func productionActionFilter(kind string, provider string) map[string]string {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if provider != "" {
		filter["provider"] = provider
	}
	if len(filter) == 0 {
		return nil
	}
	return filter
}

func actionString(action map[string]any, key string) string {
	value, _ := action[key].(string)
	return value
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
	if len(report.Filter) > 0 {
		parts := []string{}
		if report.Filter["kind"] != "" {
			parts = append(parts, "kind="+report.Filter["kind"])
		}
		if report.Filter["provider"] != "" {
			parts = append(parts, "provider="+report.Filter["provider"])
		}
		fmt.Fprintf(&builder, "Filter: %s\n", strings.Join(parts, " "))
	}
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
