package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	annotated := annotateProductionActions(raw.Actions, filepath.Dir(status.FinalizerNextActions.JSONPath))
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

func annotateProductionActions(actions []map[string]any, sourceDir string) []map[string]any {
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
		annotateReleaseProof(next)
		annotateCompetitorSetup(next, sourceDir)
		annotated = append(annotated, next)
	}
	return annotated
}

func annotateReleaseProof(action map[string]any) {
	if actionString(action, "kind") != "release_proof" {
		return
	}
	evidencePath := actionString(action, "evidence")
	if evidencePath == "" {
		return
	}
	summaryPath := filepath.Join(filepath.Dir(evidencePath), "summary.json")
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		action["release_summary_error"] = err.Error()
		return
	}
	var summary struct {
		Status                   string   `json:"status"`
		PublicReleaseReady       bool     `json:"public_release_ready"`
		ReleaseArtifactsVerified bool     `json:"release_artifacts_verified"`
		PreflightStatus          string   `json:"preflight_status"`
		BlockedCount             int      `json:"blocked_count"`
		BlockedChecks            []string `json:"blocked_checks"`
		OriginRemoteConfigured   bool     `json:"origin_remote_configured"`
		GitHubAuthStatus         string   `json:"github_auth_status"`
	}
	if err := json.Unmarshal(content, &summary); err != nil {
		action["release_summary_error"] = err.Error()
		return
	}
	action["release_summary"] = map[string]any{
		"status":                     summary.Status,
		"public_release_ready":       summary.PublicReleaseReady,
		"release_artifacts_verified": summary.ReleaseArtifactsVerified,
		"preflight_status":           summary.PreflightStatus,
		"blocked_count":              summary.BlockedCount,
		"blocked_checks":             summary.BlockedChecks,
		"origin_remote_configured":   summary.OriginRemoteConfigured,
		"github_auth_status":         summary.GitHubAuthStatus,
	}
}

func annotateCompetitorSetup(action map[string]any, sourceDir string) {
	if actionString(action, "kind") != "competitor_setup" {
		return
	}
	inspectPath := actionString(action, "inspect")
	if inspectPath == "" {
		return
	}
	if !filepath.IsAbs(inspectPath) {
		inspectPath = filepath.Join(sourceDir, inspectPath)
	}
	content, err := os.ReadFile(inspectPath)
	if err != nil {
		action["competitor_summary_error"] = err.Error()
		return
	}
	var summary struct {
		Competitors  int `json:"competitors"`
		SmokePassed  int `json:"smoke_passed"`
		SmokeFailed  int `json:"smoke_failed"`
		SetupBlocked int `json:"setup_blocked"`
		Skipped      int `json:"skipped"`
		Results      []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Status    string `json:"status"`
			SetupHint string `json:"setup_hint"`
			Note      string `json:"note"`
			Version   struct {
				Error string `json:"error"`
			} `json:"version"`
			DryRun struct {
				Error string `json:"error"`
			} `json:"dry_run"`
		} `json:"results"`
	}
	if err := json.Unmarshal(content, &summary); err != nil {
		action["competitor_summary_error"] = err.Error()
		return
	}
	blockers := []map[string]string{}
	for _, result := range summary.Results {
		if result.Status == "smoke_pass" {
			continue
		}
		reason := result.SetupHint
		if reason == "" {
			reason = result.Note
		}
		if reason == "" {
			reason = result.DryRun.Error
		}
		if reason == "" {
			reason = result.Version.Error
		}
		blockers = append(blockers, map[string]string{
			"id":     result.ID,
			"name":   result.Name,
			"status": result.Status,
			"reason": reason,
		})
	}
	action["competitor_summary"] = map[string]any{
		"competitors":   summary.Competitors,
		"smoke_passed":  summary.SmokePassed,
		"smoke_failed":  summary.SmokeFailed,
		"setup_blocked": summary.SetupBlocked,
		"skipped":       summary.Skipped,
		"blockers":      blockers,
	}
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
		writeReleaseProofText(&builder, action)
		writeCompetitorSetupText(&builder, action)
	}
	return builder.String()
}

func writeReleaseProofText(builder *strings.Builder, action map[string]any) {
	summary, _ := action["release_summary"].(map[string]any)
	if summary == nil {
		return
	}
	fmt.Fprintf(
		builder,
		"  Release readiness: %s, public_ready=%t, artifacts_verified=%t, blocked=%d\n",
		stringValue(summary["status"]),
		boolValue(summary["public_release_ready"]),
		boolValue(summary["release_artifacts_verified"]),
		int(numberValue(summary["blocked_count"])),
	)
	checks, _ := summary["blocked_checks"].([]string)
	if len(checks) > 0 {
		fmt.Fprintf(builder, "  Blocked checks: %s\n", strings.Join(checks, ", "))
	}
}

func writeCompetitorSetupText(builder *strings.Builder, action map[string]any) {
	summary, _ := action["competitor_summary"].(map[string]any)
	if summary == nil {
		return
	}
	fmt.Fprintf(
		builder,
		"  Competitor setup: %.0f pass, %.0f blocked, %.0f skipped, %.0f failed\n",
		numberValue(summary["smoke_passed"]),
		numberValue(summary["setup_blocked"]),
		numberValue(summary["skipped"]),
		numberValue(summary["smoke_failed"]),
	)
	blockers, _ := summary["blockers"].([]map[string]string)
	for _, blocker := range blockers {
		reason := blocker["reason"]
		if reason != "" {
			fmt.Fprintf(builder, "  - %s: %s - %s\n", blocker["id"], blocker["status"], reason)
		} else {
			fmt.Fprintf(builder, "  - %s: %s\n", blocker["id"], blocker["status"])
		}
	}
}

func numberValue(value any) float64 {
	switch typed := value.(type) {
	case int:
		return float64(typed)
	case float64:
		return typed
	default:
		return 0
	}
}

func stringValue(value any) string {
	typed, _ := value.(string)
	return typed
}

func boolValue(value any) bool {
	typed, _ := value.(bool)
	return typed
}
