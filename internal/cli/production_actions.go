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
	EnvReadyActionCount int               `json:"env_ready_action_count"`
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
	annotated := annotateProductionActions(raw.Actions)
	actions := filterProductionActions(annotated, opts.productionActionKind, opts.productionActionProvider, opts.productionActionsEnvReadyOnly)
	return productionActionsReport{
		Path:                status.FinalizerNextActions.JSONPath,
		Status:              raw.Status,
		RequiredActionCount: len(actions),
		EnvReadyActionCount: countEnvReadyProductionActions(actions),
		Filter:              productionActionFilter(opts.productionActionKind, opts.productionActionProvider, opts.productionActionsEnvReadyOnly),
		Actions:             actions,
	}, nil
}

func annotateProductionActions(actions []map[string]any) []map[string]any {
	annotated := make([]map[string]any, 0, len(actions))
	for _, action := range actions {
		next := make(map[string]any, len(action)+3)
		for key, value := range action {
			next[key] = value
		}
		requiredEnv := actionString(next, "required_env")
		envReady := true
		if requiredEnv != "" {
			envReady = os.Getenv(requiredEnv) != ""
			next["required_env_set"] = envReady
			if !envReady {
				next["missing_required_env"] = requiredEnv
			}
		}
		next["env_ready"] = envReady
		annotated = append(annotated, next)
	}
	return annotated
}

func countEnvReadyProductionActions(actions []map[string]any) int {
	count := 0
	for _, action := range actions {
		if envReady, _ := action["env_ready"].(bool); envReady {
			count++
		}
	}
	return count
}

func filterProductionActions(actions []map[string]any, kind string, provider string, envReadyOnly bool) []map[string]any {
	if kind == "" && provider == "" && !envReadyOnly {
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
		if envReadyOnly {
			envReady, _ := action["env_ready"].(bool)
			if !envReady {
				continue
			}
		}
		filtered = append(filtered, action)
	}
	return filtered
}

func productionActionFilter(kind string, provider string, envReadyOnly bool) map[string]string {
	filter := map[string]string{}
	if kind != "" {
		filter["kind"] = kind
	}
	if provider != "" {
		filter["provider"] = provider
	}
	if envReadyOnly {
		filter["env_ready"] = "true"
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
	fmt.Fprintf(&builder, "Env ready: %d\n", report.EnvReadyActionCount)
	if len(report.Filter) > 0 {
		parts := []string{}
		if report.Filter["kind"] != "" {
			parts = append(parts, "kind="+report.Filter["kind"])
		}
		if report.Filter["provider"] != "" {
			parts = append(parts, "provider="+report.Filter["provider"])
		}
		if report.Filter["env_ready"] != "" {
			parts = append(parts, "env_ready="+report.Filter["env_ready"])
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
		suffix := ""
		if missingEnv, _ := action["missing_required_env"].(string); missingEnv != "" {
			suffix = " (missing env: " + missingEnv + ")"
		}
		if kind != "" {
			fmt.Fprintf(&builder, "- %s [%s]: %s%s\n", id, kind, text, suffix)
		} else {
			fmt.Fprintf(&builder, "- %s: %s%s\n", id, text, suffix)
		}
	}
	return builder.String()
}
