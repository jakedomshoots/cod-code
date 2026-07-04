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
	ReadyActionCount    int               `json:"ready_action_count"`
	NextOnly            bool              `json:"next_only,omitempty"`
	CommandsOnly        bool              `json:"commands_only,omitempty"`
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
	actions := filterProductionActions(annotated, opts.productionActionID, opts.productionActionKind, opts.productionActionProvider, opts.productionActionState, opts.productionActionsEnvReadyOnly, opts.productionActionsReadyOnly, opts.productionActionsNextOnly)
	return productionActionsReport{
		Path:                status.FinalizerNextActions.JSONPath,
		Status:              raw.Status,
		RequiredActionCount: len(actions),
		EnvReadyActionCount: countEnvReadyProductionActions(actions),
		ReadyActionCount:    countReadyProductionActions(actions),
		NextOnly:            opts.productionActionsNextOnly,
		CommandsOnly:        opts.productionActionsCommandsOnly,
		Filter:              productionActionFilter(opts.productionActionID, opts.productionActionKind, opts.productionActionProvider, opts.productionActionState, opts.productionActionsEnvReadyOnly, opts.productionActionsReadyOnly, opts.productionActionsNextOnly),
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
			value, present := os.LookupEnv(requiredEnv)
			envReady = present && value != ""
			next["required_env_set"] = present
			if !envReady {
				if present {
					next["empty_required_env"] = requiredEnv
				} else {
					next["missing_required_env"] = requiredEnv
				}
			}
		}
		next["env_ready"] = envReady
		annotateReleaseProof(next)
		annotateProviderProof(next)
		annotateCompetitorSetup(next, sourceDir)
		annotated = append(annotated, next)
	}
	annotateProductionActionDependencies(annotated)
	annotateProductionActionStates(annotated)
	return annotated
}

func annotateProductionActionStates(actions []map[string]any) {
	for _, action := range actions {
		action["action_state"] = productionActionState(action)
	}
}

func productionActionState(action map[string]any) string {
	if emptyEnv := actionString(action, "empty_required_env"); emptyEnv != "" {
		return "empty_env"
	}
	if missingEnv := actionString(action, "missing_required_env"); missingEnv != "" {
		return "missing_env"
	}
	if hasSetupBlocker(action) {
		return "setup_blocked"
	}
	if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
		return "waiting"
	}
	return "ready"
}

func hasSetupBlocker(action map[string]any) bool {
	if summary, _ := action["release_summary"].(map[string]any); summary != nil {
		return numberValue(summary["blocked_count"]) > 0 || !boolValue(summary["public_release_ready"])
	}
	if summary, _ := action["competitor_summary"].(map[string]any); summary != nil {
		blockers, _ := summary["blockers"].([]map[string]string)
		return len(blockers) > 0 || numberValue(summary["setup_blocked"]) > 0 || numberValue(summary["skipped"]) > 0 || numberValue(summary["smoke_failed"]) > 0
	}
	return false
}

func annotateProductionActionDependencies(actions []map[string]any) {
	byID := map[string]map[string]any{}
	for _, action := range actions {
		if id := actionString(action, "id"); id != "" {
			byID[id] = action
		}
	}
	if comparison := byID["all-agent-29-comparison"]; comparison != nil {
		if actionHasOpenBlocker(byID["competitor-smoke"]) {
			addBlockedBy(comparison, "competitor-smoke")
		}
	}
	if final := byID["production-readiness"]; final != nil {
		for _, action := range actions {
			id := actionString(action, "id")
			if id == "" || id == "production-readiness" {
				continue
			}
			if actionHasOpenBlocker(action) {
				addBlockedBy(final, id)
			}
		}
	}
}

func actionHasOpenBlocker(action map[string]any) bool {
	if action == nil {
		return false
	}
	if emptyEnv := actionString(action, "empty_required_env"); emptyEnv != "" {
		return true
	}
	if missingEnv := actionString(action, "missing_required_env"); missingEnv != "" {
		return true
	}
	if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
		return true
	}
	if hasSetupBlocker(action) {
		return true
	}
	status := actionString(action, "status")
	return status != "" && status != "pass"
}

func addBlockedBy(action map[string]any, id string) {
	if id == "" {
		return
	}
	current := stringSlice(action["blocked_by"])
	for _, existing := range current {
		if existing == id {
			return
		}
	}
	action["blocked_by"] = append(current, id)
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
		SetupActions             string   `json:"setup_actions"`
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
		"setup_actions":              summary.SetupActions,
		"origin_remote_configured":   summary.OriginRemoteConfigured,
		"github_auth_status":         summary.GitHubAuthStatus,
	}
	if summary.SetupActions != "" {
		setupActionsPath := filepath.Join(filepath.Dir(summaryPath), summary.SetupActions)
		releaseSummary := action["release_summary"].(map[string]any)
		releaseSummary["setup_actions_path"] = setupActionsPath
		if setupActionItems, err := readReleaseSetupActionItems(setupActionsPath); err == nil && len(setupActionItems) > 0 {
			releaseSummary["setup_action_items"] = setupActionItems
		}
	}
}

func readReleaseSetupActionItems(path string) ([]map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	items := []map[string]string{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(line, "- "))
		check, text, ok := strings.Cut(body, ":")
		if !ok {
			continue
		}
		check = strings.TrimSpace(strings.Trim(check, "`"))
		text = strings.TrimSpace(text)
		if check == "" || text == "" {
			continue
		}
		items = append(items, map[string]string{
			"check": check,
			"text":  text,
		})
	}
	return items, nil
}

func annotateProviderProof(action map[string]any) {
	if actionString(action, "kind") != "provider_proof" {
		return
	}
	evidencePath := actionString(action, "evidence")
	if evidencePath == "" {
		return
	}
	summaryPath := filepath.Join(filepath.Dir(evidencePath), "summary.json")
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		action["provider_summary_error"] = err.Error()
		return
	}
	var summary struct {
		Status           string `json:"status"`
		Provider         string `json:"provider"`
		ProviderMode     string `json:"provider_mode"`
		HTTPPreset       string `json:"http_preset"`
		HTTPModel        string `json:"http_model"`
		APIKeyEnv        string `json:"api_key_env"`
		BlockedReason    string `json:"blocked_reason"`
		SecretValueSaved bool   `json:"secret_value_saved"`
		Artifacts        struct {
			Checklist   string `json:"checklist"`
			Commands    string `json:"commands"`
			EnvTemplate string `json:"env_template"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(content, &summary); err != nil {
		action["provider_summary_error"] = err.Error()
		return
	}
	providerSummary := map[string]any{
		"status":             summary.Status,
		"provider":           summary.Provider,
		"provider_mode":      summary.ProviderMode,
		"http_preset":        summary.HTTPPreset,
		"http_model":         summary.HTTPModel,
		"api_key_env":        summary.APIKeyEnv,
		"blocked_reason":     summary.BlockedReason,
		"secret_value_saved": summary.SecretValueSaved,
	}
	if summary.Artifacts.Checklist != "" {
		providerSummary["checklist_path"] = filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.Checklist)
	}
	if summary.Artifacts.Commands != "" {
		providerSummary["commands_path"] = filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.Commands)
	}
	if summary.Artifacts.EnvTemplate != "" {
		providerSummary["env_template_path"] = filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.EnvTemplate)
	}
	action["provider_summary"] = providerSummary
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
		Competitors  int    `json:"competitors"`
		SmokePassed  int    `json:"smoke_passed"`
		SmokeFailed  int    `json:"smoke_failed"`
		SetupBlocked int    `json:"setup_blocked"`
		Skipped      int    `json:"skipped"`
		SetupActions string `json:"setup_actions"`
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
		"setup_actions": summary.SetupActions,
		"blockers":      blockers,
	}
	if summary.SetupActions != "" {
		action["competitor_summary"].(map[string]any)["setup_actions_path"] = filepath.Join(filepath.Dir(inspectPath), summary.SetupActions)
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

func countReadyProductionActions(actions []map[string]any) int {
	count := 0
	for _, action := range actions {
		if actionReady(action) {
			count++
		}
	}
	return count
}

func actionReady(action map[string]any) bool {
	envReady, _ := action["env_ready"].(bool)
	return envReady && !hasSetupBlocker(action) && len(stringSlice(action["blocked_by"])) == 0
}

func filterProductionActions(actions []map[string]any, id string, kind string, provider string, state string, envReadyOnly bool, readyOnly bool, nextOnly bool) []map[string]any {
	if id == "" && kind == "" && provider == "" && state == "" && !envReadyOnly && !readyOnly && !nextOnly {
		return actions
	}
	filtered := make([]map[string]any, 0, len(actions))
	for _, action := range actions {
		if id != "" && actionString(action, "id") != id {
			continue
		}
		if kind != "" && actionString(action, "kind") != kind {
			continue
		}
		if provider != "" && actionString(action, "provider") != provider {
			continue
		}
		if state != "" && actionString(action, "action_state") != state {
			continue
		}
		if envReadyOnly {
			envReady, _ := action["env_ready"].(bool)
			if !envReady {
				continue
			}
		}
		if readyOnly && !actionReady(action) {
			continue
		}
		filtered = append(filtered, action)
	}
	if nextOnly {
		return firstReadyAction(filtered)
	}
	return filtered
}

func firstReadyAction(actions []map[string]any) []map[string]any {
	for _, action := range actions {
		if actionReady(action) {
			return []map[string]any{action}
		}
	}
	return []map[string]any{}
}

func productionActionFilter(id string, kind string, provider string, state string, envReadyOnly bool, readyOnly bool, nextOnly bool) map[string]string {
	filter := map[string]string{}
	if id != "" {
		filter["id"] = id
	}
	if kind != "" {
		filter["kind"] = kind
	}
	if provider != "" {
		filter["provider"] = provider
	}
	if state != "" {
		filter["state"] = state
	}
	if envReadyOnly {
		filter["env_ready"] = "true"
	}
	if readyOnly {
		filter["ready"] = "true"
	}
	if nextOnly {
		filter["next"] = "true"
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
	if report.CommandsOnly {
		_, err := fmt.Fprint(out, renderProductionActionCommandsOnly(report))
		return err
	}
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
	if report.CommandsOnly {
		return renderProductionActionCommandsOnly(report)
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "Production actions: %s\n", report.Status)
	fmt.Fprintf(&builder, "Required actions: %d\n", report.RequiredActionCount)
	fmt.Fprintf(&builder, "Env ready: %d\n", report.EnvReadyActionCount)
	fmt.Fprintf(&builder, "Ready now: %d\n", report.ReadyActionCount)
	if len(report.Filter) > 0 {
		parts := []string{}
		if report.Filter["id"] != "" {
			parts = append(parts, "id="+report.Filter["id"])
		}
		if report.Filter["kind"] != "" {
			parts = append(parts, "kind="+report.Filter["kind"])
		}
		if report.Filter["provider"] != "" {
			parts = append(parts, "provider="+report.Filter["provider"])
		}
		if report.Filter["state"] != "" {
			parts = append(parts, "state="+report.Filter["state"])
		}
		if report.Filter["env_ready"] != "" {
			parts = append(parts, "env_ready="+report.Filter["env_ready"])
		}
		if report.Filter["ready"] != "" {
			parts = append(parts, "ready="+report.Filter["ready"])
		}
		if report.Filter["next"] != "" {
			parts = append(parts, "next="+report.Filter["next"])
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
		if emptyEnv, _ := action["empty_required_env"].(string); emptyEnv != "" {
			suffix = " (empty env: " + emptyEnv + ")"
		} else if missingEnv, _ := action["missing_required_env"].(string); missingEnv != "" {
			suffix = " (missing env: " + missingEnv + ")"
		}
		if kind != "" {
			fmt.Fprintf(&builder, "- %s [%s]: %s%s\n", id, kind, text, suffix)
		} else {
			fmt.Fprintf(&builder, "- %s: %s%s\n", id, text, suffix)
		}
		writeReleaseProofText(&builder, action)
		writeProviderProofText(&builder, action)
		writeCompetitorSetupText(&builder, action)
		writeBlockedByText(&builder, action)
		writeActionCommandText(&builder, action)
	}
	return builder.String()
}

func renderProductionActionCommandsOnly(report productionActionsReport) string {
	var builder strings.Builder
	for _, action := range report.Actions {
		command := stringSlice(action["command"])
		if len(command) == 0 {
			continue
		}
		id := actionString(action, "id")
		kind := actionString(action, "kind")
		if id != "" || kind != "" {
			fmt.Fprintf(&builder, "# %s", id)
			if kind != "" {
				fmt.Fprintf(&builder, " [%s]", kind)
			}
			if requiredEnv := actionString(action, "required_env"); requiredEnv != "" {
				if actionString(action, "empty_required_env") != "" {
					fmt.Fprintf(&builder, " empty env: %s", requiredEnv)
				} else if actionString(action, "missing_required_env") != "" {
					fmt.Fprintf(&builder, " missing env: %s", requiredEnv)
				} else {
					fmt.Fprintf(&builder, " requires env: %s", requiredEnv)
				}
			}
			if state := actionString(action, "action_state"); state != "" {
				fmt.Fprintf(&builder, " state: %s", state)
			}
			if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
				fmt.Fprintf(&builder, " waiting on: %s", strings.Join(blockedBy, ", "))
			}
			builder.WriteString("\n")
		}
		fmt.Fprintf(&builder, "%s\n", shellCommandLine(command))
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
	if setupActions := stringValue(summary["setup_actions_path"]); setupActions != "" {
		fmt.Fprintf(builder, "  Setup actions: %s\n", setupActions)
	}
	if items, _ := summary["setup_action_items"].([]map[string]string); len(items) > 0 {
		fmt.Fprintf(builder, "  Setup action items:\n")
		for _, item := range items {
			fmt.Fprintf(builder, "  - %s: %s\n", item["check"], item["text"])
		}
	}
}

func writeProviderProofText(builder *strings.Builder, action map[string]any) {
	summary, _ := action["provider_summary"].(map[string]any)
	if summary == nil {
		return
	}
	if blockedReason := stringValue(summary["blocked_reason"]); blockedReason != "" {
		fmt.Fprintf(builder, "  Provider blocker: %s\n", blockedReason)
	}
	if model := stringValue(summary["http_model"]); model != "" {
		fmt.Fprintf(builder, "  Provider model: %s\n", model)
	}
	if checklist := stringValue(summary["checklist_path"]); checklist != "" {
		fmt.Fprintf(builder, "  Setup checklist: %s\n", checklist)
	}
	if commands := stringValue(summary["commands_path"]); commands != "" {
		fmt.Fprintf(builder, "  Setup command file: %s\n", commands)
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
	if setupActions := stringValue(summary["setup_actions_path"]); setupActions != "" {
		fmt.Fprintf(builder, "  Setup actions: %s\n", setupActions)
	}
}

func writeBlockedByText(builder *strings.Builder, action map[string]any) {
	blockedBy := stringSlice(action["blocked_by"])
	if len(blockedBy) == 0 {
		return
	}
	fmt.Fprintf(builder, "  Waiting on: %s\n", strings.Join(blockedBy, ", "))
}

func writeActionCommandText(builder *strings.Builder, action map[string]any) {
	command := stringSlice(action["command"])
	if len(command) == 0 {
		return
	}
	if requiredEnv := actionString(action, "required_env"); requiredEnv != "" {
		fmt.Fprintf(builder, "  Requires env: %s\n", requiredEnv)
	}
	fmt.Fprintf(builder, "  Command: %s\n", shellCommandLine(command))
}

func shellCommandLine(args []string) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func shellQuote(arg string) string {
	if arg == "" {
		return "''"
	}
	if strings.IndexFunc(arg, func(r rune) bool {
		return !(r == '_' || r == '-' || r == '.' || r == '/' || r == ':' || r == '=' || r == '@' || r == '+' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	}) == -1 {
		return arg
	}
	return "'" + strings.ReplaceAll(arg, "'", "'\"'\"'") + "'"
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

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}
