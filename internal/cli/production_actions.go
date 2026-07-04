package cli

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type productionActionsReport struct {
	Path                          string            `json:"path"`
	Status                        string            `json:"status"`
	RequiredActionCount           int               `json:"required_action_count"`
	EnvReadyActionCount           int               `json:"env_ready_action_count"`
	ProviderEnvReadyActionCount   int               `json:"provider_env_ready_action_count"`
	ProviderEnvBlockedActionCount int               `json:"provider_env_blocked_action_count"`
	ReadyActionCount              int               `json:"ready_action_count"`
	RunnableCommandCount          int               `json:"runnable_command_count"`
	BlockedCommandCount           int               `json:"blocked_command_count"`
	EvidenceDeclaredMatchCount    int               `json:"evidence_declared_match_count"`
	EvidenceDeclaredMismatchCount int               `json:"evidence_declared_mismatch_count"`
	ActionStateCounts             map[string]int    `json:"action_state_counts,omitempty"`
	NextOnly                      bool              `json:"next_only,omitempty"`
	CommandsOnly                  bool              `json:"commands_only,omitempty"`
	Filter                        map[string]string `json:"filter,omitempty"`
	NextBlockedAction             map[string]any    `json:"next_blocked_action,omitempty"`
	Actions                       []map[string]any  `json:"actions"`
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
	filteredCandidates := filterProductionActions(annotated, opts.productionActionID, opts.productionActionKind, opts.productionActionProvider, opts.productionActionState, opts.productionActionsEnvReadyOnly, opts.productionActionsReadyOnly, false)
	actions := filteredCandidates
	var nextBlockedAction map[string]any
	if opts.productionActionsNextOnly {
		actions = firstReadyAction(filteredCandidates)
		if len(actions) == 0 {
			nextBlockedAction = firstBlockedAction(filteredCandidates)
		}
	}
	evidenceMatches, evidenceMismatches := countProductionEvidenceDeclaredMatches(actions)
	return productionActionsReport{
		Path:                          status.FinalizerNextActions.JSONPath,
		Status:                        raw.Status,
		RequiredActionCount:           len(actions),
		EnvReadyActionCount:           countEnvReadyProductionActions(actions),
		ProviderEnvReadyActionCount:   countProviderEnvProductionActions(actions, true),
		ProviderEnvBlockedActionCount: countProviderEnvProductionActions(actions, false),
		ReadyActionCount:              countReadyProductionActions(actions),
		RunnableCommandCount:          countRunnableProductionActionCommands(actions),
		BlockedCommandCount:           countBlockedProductionActionCommands(actions),
		EvidenceDeclaredMatchCount:    evidenceMatches,
		EvidenceDeclaredMismatchCount: evidenceMismatches,
		ActionStateCounts:             countProductionActionStates(actions),
		NextOnly:                      opts.productionActionsNextOnly,
		CommandsOnly:                  opts.productionActionsCommandsOnly,
		Filter:                        productionActionFilter(opts.productionActionID, opts.productionActionKind, opts.productionActionProvider, opts.productionActionState, opts.productionActionsEnvReadyOnly, opts.productionActionsReadyOnly, opts.productionActionsNextOnly),
		NextBlockedAction:             nextBlockedAction,
		Actions:                       actions,
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
		annotateDeclaredEvidence(next, sourceDir)
		annotateReleaseProof(next)
		annotateProviderProof(next)
		markPassedProviderProofSatisfied(next)
		annotateCompetitorSetup(next, sourceDir)
		annotated = append(annotated, next)
	}
	annotateProductionActionDependencies(annotated)
	annotateProductionActionStates(annotated)
	return annotated
}

func annotateDeclaredEvidence(action map[string]any, sourceDir string) {
	declared := declaredEvidenceFilesByField(action["declared_evidence_files"])
	refs := []struct {
		field string
		path  string
	}{
		{field: "evidence", path: actionString(action, "evidence")},
		{field: "inspect", path: actionString(action, "inspect")},
	}
	files := []map[string]any{}
	for _, ref := range refs {
		if strings.TrimSpace(ref.path) == "" {
			continue
		}
		path := ref.path
		if !filepath.IsAbs(path) && sourceDir != "" {
			path = filepath.Join(sourceDir, path)
		}
		entry := map[string]any{
			"field": ref.field,
			"path":  path,
		}
		if declaredEntry := declared[ref.field]; declaredEntry != nil {
			if declaredPath := stringValue(declaredEntry["path"]); declaredPath != "" {
				entry["declared_path"] = declaredPath
			}
			if declaredSHA := stringValue(declaredEntry["sha256"]); declaredSHA != "" {
				entry["declared_sha256"] = declaredSHA
			}
			if declaredSize := numberValue(declaredEntry["size_bytes"]); declaredSize > 0 {
				entry["declared_size_bytes"] = int(declaredSize)
			}
			if declaredExists, ok := declaredEntry["exists"].(bool); ok {
				entry["declared_exists"] = declaredExists
			}
		}
		content, err := os.ReadFile(path)
		if err != nil {
			entry["exists"] = false
			entry["error"] = err.Error()
			if _, ok := entry["declared_sha256"]; ok {
				entry["matches_declared"] = false
			}
			files = append(files, entry)
			continue
		}
		sum := sha256.Sum256(content)
		currentSHA := fmt.Sprintf("%x", sum[:])
		entry["exists"] = true
		entry["size_bytes"] = len(content)
		entry["sha256"] = currentSHA
		if declaredSHA := stringValue(entry["declared_sha256"]); declaredSHA != "" {
			entry["matches_declared"] = declaredSHA == currentSHA
		}
		files = append(files, entry)
	}
	if len(files) > 0 {
		action["evidence_files"] = files
	}
}

func annotateProductionActionStates(actions []map[string]any) {
	for _, action := range actions {
		action["action_state"] = productionActionState(action)
		action["action_reason"] = productionActionReason(action)
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

func productionActionReason(action map[string]any) string {
	if emptyEnv := actionString(action, "empty_required_env"); emptyEnv != "" {
		return "required env is set but empty: " + emptyEnv
	}
	if missingEnv := actionString(action, "missing_required_env"); missingEnv != "" {
		return "missing required env: " + missingEnv
	}
	if summary, _ := action["release_summary"].(map[string]any); summary != nil {
		if checks, _ := summary["blocked_checks"].([]string); len(checks) > 0 {
			return "release setup blocked: " + strings.Join(checks, ", ")
		}
		if numberValue(summary["blocked_count"]) > 0 || !boolValue(summary["public_release_ready"]) {
			return "release setup blocked"
		}
	}
	if summary, _ := action["competitor_summary"].(map[string]any); summary != nil {
		if blockers, _ := summary["blockers"].([]map[string]string); len(blockers) > 0 {
			parts := make([]string, 0, len(blockers))
			for _, blocker := range blockers {
				if id := blocker["id"]; id != "" {
					parts = append(parts, id)
				}
			}
			if len(parts) > 0 {
				return "competitor setup blocked: " + strings.Join(parts, ", ")
			}
		}
		if numberValue(summary["setup_blocked"]) > 0 || numberValue(summary["skipped"]) > 0 || numberValue(summary["smoke_failed"]) > 0 {
			return "competitor setup blocked"
		}
	}
	if providerProofPassed(action) {
		return "provider proof already passed"
	}
	if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
		return "waiting on: " + strings.Join(blockedBy, ", ")
	}
	if status := actionString(action, "status"); status != "" && status != "pass" {
		return "status is " + status
	}
	return "ready to run"
}

func validProductionActionState(state string) bool {
	switch state {
	case "ready", "missing_env", "empty_env", "setup_blocked", "waiting":
		return true
	default:
		return false
	}
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
	if providerProofPassed(action) {
		return false
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
		SetupActionCount         int      `json:"setup_action_count"`
		SetupActionsSHA256       string   `json:"setup_actions_sha256"`
		SetupCommands            string   `json:"setup_commands"`
		SetupCommandsSHA256      string   `json:"setup_commands_sha256"`
		SetupCommandPolicy       string   `json:"setup_command_policy"`
		PublishActionsPerformed  bool     `json:"publish_actions_performed"`
		SecretValueSaved         bool     `json:"secret_value_saved"`
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
		"setup_action_count":         summary.SetupActionCount,
		"setup_actions_sha256":       summary.SetupActionsSHA256,
		"setup_commands":             summary.SetupCommands,
		"setup_commands_sha256":      summary.SetupCommandsSHA256,
		"setup_command_policy":       summary.SetupCommandPolicy,
		"publish_actions_performed":  summary.PublishActionsPerformed,
		"secret_value_saved":         summary.SecretValueSaved,
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
	if summary.SetupCommands != "" {
		releaseSummary := action["release_summary"].(map[string]any)
		releaseSummary["setup_commands_path"] = filepath.Join(filepath.Dir(summaryPath), summary.SetupCommands)
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
		Status                    string            `json:"status"`
		Provider                  string            `json:"provider"`
		ProviderMode              string            `json:"provider_mode"`
		HTTPPreset                string            `json:"http_preset"`
		HTTPModel                 string            `json:"http_model"`
		APIKeyEnv                 string            `json:"api_key_env"`
		BlockedReason             string            `json:"blocked_reason"`
		SecretValueSaved          bool              `json:"secret_value_saved"`
		CommandScriptSecretPolicy string            `json:"command_script_secret_policy"`
		SetupChecklistItemCount   int               `json:"setup_checklist_item_count"`
		SetupArtifactsSHA256      map[string]string `json:"setup_artifacts_sha256"`
		Artifacts                 struct {
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
	if summary.CommandScriptSecretPolicy != "" {
		providerSummary["command_script_secret_policy"] = summary.CommandScriptSecretPolicy
	}
	if summary.SetupChecklistItemCount > 0 {
		providerSummary["setup_checklist_item_count"] = summary.SetupChecklistItemCount
	}
	if len(summary.SetupArtifactsSHA256) > 0 {
		providerSummary["setup_artifacts_sha256"] = summary.SetupArtifactsSHA256
	}
	if summary.Artifacts.Checklist != "" {
		checklistPath := filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.Checklist)
		providerSummary["checklist_path"] = checklistPath
		if checklistItems, err := readNumberedChecklistItems(checklistPath); err == nil && len(checklistItems) > 0 {
			providerSummary["checklist_items"] = checklistItems
		}
	}
	if summary.Artifacts.Commands != "" {
		providerSummary["commands_path"] = filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.Commands)
	}
	if summary.Artifacts.EnvTemplate != "" {
		providerSummary["env_template_path"] = filepath.Join(filepath.Dir(summaryPath), summary.Artifacts.EnvTemplate)
	}
	action["provider_summary"] = providerSummary
}

func markPassedProviderProofSatisfied(action map[string]any) {
	if !providerProofPassed(action) {
		return
	}
	if requiredEnv := actionString(action, "required_env"); requiredEnv != "" {
		action["required_env_satisfied_by_evidence"] = true
	}
	action["env_ready"] = true
	delete(action, "missing_required_env")
	delete(action, "empty_required_env")
}

func providerProofPassed(action map[string]any) bool {
	if actionString(action, "kind") != "provider_proof" {
		return false
	}
	summary, _ := action["provider_summary"].(map[string]any)
	return summary != nil && actionString(summary, "status") == "pass"
}

func readNumberedChecklistItems(path string) ([]map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	items := []map[string]string{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		dot := strings.Index(line, ".")
		if dot <= 0 {
			continue
		}
		rawNumber := line[:dot]
		for _, r := range rawNumber {
			if r < '0' || r > '9' {
				rawNumber = ""
				break
			}
		}
		if rawNumber == "" {
			continue
		}
		text := strings.TrimSpace(line[dot+1:])
		if text == "" {
			continue
		}
		items = append(items, map[string]string{
			"step": rawNumber,
			"text": text,
		})
	}
	return items, nil
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

func countProviderEnvProductionActions(actions []map[string]any, ready bool) int {
	count := 0
	for _, action := range actions {
		if actionString(action, "required_env") == "" {
			continue
		}
		envReady, _ := action["env_ready"].(bool)
		if envReady == ready {
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

func countRunnableProductionActionCommands(actions []map[string]any) int {
	count := 0
	for _, action := range actions {
		if len(stringSlice(action["command"])) > 0 && actionReady(action) {
			count++
		}
	}
	return count
}

func countBlockedProductionActionCommands(actions []map[string]any) int {
	count := 0
	for _, action := range actions {
		if len(stringSlice(action["command"])) > 0 && !actionReady(action) {
			count++
		}
	}
	return count
}

func countProductionActionStates(actions []map[string]any) map[string]int {
	counts := map[string]int{}
	for _, action := range actions {
		state := actionString(action, "action_state")
		if state == "" {
			state = "unknown"
		}
		counts[state]++
	}
	if len(counts) == 0 {
		return nil
	}
	return counts
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

func firstBlockedAction(actions []map[string]any) map[string]any {
	for _, action := range actions {
		if !actionReady(action) {
			return action
		}
	}
	return nil
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
	fmt.Fprintf(&builder, "Provider env ready: %d\n", report.ProviderEnvReadyActionCount)
	fmt.Fprintf(&builder, "Provider env blocked: %d\n", report.ProviderEnvBlockedActionCount)
	fmt.Fprintf(&builder, "Ready now: %d\n", report.ReadyActionCount)
	fmt.Fprintf(&builder, "Runnable commands: %d\n", report.RunnableCommandCount)
	fmt.Fprintf(&builder, "Blocked commands: %d\n", report.BlockedCommandCount)
	fmt.Fprintf(&builder, "Evidence matches: declared=%d mismatched=%d\n", report.EvidenceDeclaredMatchCount, report.EvidenceDeclaredMismatchCount)
	if len(report.ActionStateCounts) > 0 {
		fmt.Fprintf(&builder, "States: %s\n", renderActionStateCounts(report.ActionStateCounts))
	}
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
	if len(report.Actions) == 0 && report.NextBlockedAction != nil {
		writeNextBlockedActionText(&builder, report.NextBlockedAction)
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
		if reason := actionString(action, "action_reason"); reason != "" {
			fmt.Fprintf(&builder, "  Reason: %s\n", reason)
		}
		writeReleaseProofText(&builder, action)
		writeProviderProofText(&builder, action)
		writeCompetitorSetupText(&builder, action)
		writeEvidenceFilesText(&builder, action)
		writeBlockedByText(&builder, action)
		writeActionCommandText(&builder, action)
	}
	return builder.String()
}

func writeNextBlockedActionText(builder *strings.Builder, action map[string]any) {
	id := actionString(action, "id")
	kind := actionString(action, "kind")
	state := actionString(action, "action_state")
	reason := actionString(action, "action_reason")
	if state == "" {
		state = "unknown"
	}
	if kind != "" {
		fmt.Fprintf(builder, "Next blocked action: %s [%s] state=%s\n", id, kind, state)
	} else {
		fmt.Fprintf(builder, "Next blocked action: %s state=%s\n", id, state)
	}
	if reason != "" {
		fmt.Fprintf(builder, "Blocked reason: %s\n", reason)
	}
	if missingEnv := actionString(action, "missing_required_env"); missingEnv != "" {
		fmt.Fprintf(builder, "Missing env: %s\n", missingEnv)
	} else if emptyEnv := actionString(action, "empty_required_env"); emptyEnv != "" {
		fmt.Fprintf(builder, "Empty env: %s\n", emptyEnv)
	}
	if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
		fmt.Fprintf(builder, "Waiting on: %s\n", strings.Join(blockedBy, ", "))
	}
}

func renderActionStateCounts(counts map[string]int) string {
	order := []string{"ready", "missing_env", "empty_env", "setup_blocked", "waiting", "unknown"}
	parts := []string{}
	seen := map[string]bool{}
	for _, state := range order {
		if count := counts[state]; count > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", state, count))
			seen[state] = true
		}
	}
	for state, count := range counts {
		if seen[state] || count <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%d", state, count))
	}
	return strings.Join(parts, " ")
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
			if reason := actionString(action, "action_reason"); reason != "" {
				fmt.Fprintf(&builder, " reason: %s", reason)
			}
			if blockedBy := stringSlice(action["blocked_by"]); len(blockedBy) > 0 {
				fmt.Fprintf(&builder, " waiting on: %s", strings.Join(blockedBy, ", "))
			}
			builder.WriteString("\n")
		}
		writeActionSetupComments(&builder, action)
		commandLine := shellCommandLine(command)
		if actionReady(action) {
			fmt.Fprintf(&builder, "%s\n", commandLine)
		} else {
			fmt.Fprintf(&builder, "# blocked command: %s\n", commandLine)
		}
	}
	return builder.String()
}

func writeActionSetupComments(builder *strings.Builder, action map[string]any) {
	if summary, _ := action["release_summary"].(map[string]any); summary != nil {
		if items, _ := summary["setup_action_items"].([]map[string]string); len(items) > 0 {
			fmt.Fprintf(builder, "# setup actions:\n")
			for _, item := range items {
				fmt.Fprintf(builder, "# - %s: %s\n", item["check"], item["text"])
			}
		}
	}
	if summary, _ := action["provider_summary"].(map[string]any); summary != nil {
		if items, _ := summary["checklist_items"].([]map[string]string); len(items) > 0 {
			fmt.Fprintf(builder, "# setup checklist:\n")
			for _, item := range items {
				fmt.Fprintf(builder, "# %s. %s\n", item["step"], item["text"])
			}
		}
	}
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
	if setupCommands := stringValue(summary["setup_commands_path"]); setupCommands != "" {
		fmt.Fprintf(builder, "  Setup command file: %s\n", setupCommands)
	}
	if count := int(numberValue(summary["setup_action_count"])); count > 0 {
		fmt.Fprintf(builder, "  Setup action count: %d\n", count)
	}
	if sha := stringValue(summary["setup_actions_sha256"]); sha != "" {
		fmt.Fprintf(builder, "  Setup actions sha256: %s\n", sha)
	}
	if sha := stringValue(summary["setup_commands_sha256"]); sha != "" {
		fmt.Fprintf(builder, "  Setup commands sha256: %s\n", sha)
	}
	if policy := stringValue(summary["setup_command_policy"]); policy != "" {
		fmt.Fprintf(builder, "  Setup command policy: %s\n", policy)
	}
	if publishActions, ok := summary["publish_actions_performed"].(bool); ok {
		fmt.Fprintf(builder, "  Publish actions performed: %t\n", publishActions)
	}
	if secretSaved, ok := summary["secret_value_saved"].(bool); ok {
		fmt.Fprintf(builder, "  Secret value saved: %t\n", secretSaved)
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
	if policy := stringValue(summary["command_script_secret_policy"]); policy != "" {
		fmt.Fprintf(builder, "  Command secret policy: %s\n", policy)
	}
	if checklist := stringValue(summary["checklist_path"]); checklist != "" {
		fmt.Fprintf(builder, "  Setup checklist: %s\n", checklist)
	}
	if count := int(numberValue(summary["setup_checklist_item_count"])); count > 0 {
		fmt.Fprintf(builder, "  Setup checklist count: %d\n", count)
	}
	if items, _ := summary["checklist_items"].([]map[string]string); len(items) > 0 {
		fmt.Fprintf(builder, "  Setup checklist items:\n")
		for _, item := range items {
			fmt.Fprintf(builder, "  %s. %s\n", item["step"], item["text"])
		}
	}
	if hashes := stringStringMap(summary["setup_artifacts_sha256"]); len(hashes) > 0 {
		fmt.Fprintf(builder, "  Setup artifact hashes: %s\n", renderStringMap(hashes))
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

func writeEvidenceFilesText(builder *strings.Builder, action map[string]any) {
	files := evidenceFileEntries(action["evidence_files"])
	for _, file := range files {
		field := stringValue(file["field"])
		path := stringValue(file["path"])
		if path == "" {
			continue
		}
		label := "Evidence file"
		if field == "inspect" {
			label = "Inspect file"
		}
		if exists, _ := file["exists"].(bool); exists {
			fmt.Fprintf(builder, "  %s: %s sha256=%s size=%d", label, path, stringValue(file["sha256"]), int(numberValue(file["size_bytes"])))
			if match, ok := file["matches_declared"].(bool); ok {
				fmt.Fprintf(builder, " declared_match=%t", match)
			}
			builder.WriteString("\n")
			continue
		}
		if err := stringValue(file["error"]); err != "" {
			fmt.Fprintf(builder, "  %s: %s missing (%s)\n", label, path, err)
		} else {
			fmt.Fprintf(builder, "  %s: %s missing\n", label, path)
		}
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

func stringStringMap(value any) map[string]string {
	switch typed := value.(type) {
	case map[string]string:
		return typed
	case map[string]any:
		result := map[string]string{}
		for key, value := range typed {
			if text := stringValue(value); text != "" {
				result[key] = text
			}
		}
		return result
	default:
		return nil
	}
}

func renderStringMap(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	return strings.Join(parts, " ")
}

func evidenceFileEntries(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if entry, ok := item.(map[string]any); ok {
				result = append(result, entry)
			}
		}
		return result
	default:
		return nil
	}
}

func declaredEvidenceFilesByField(value any) map[string]map[string]any {
	result := map[string]map[string]any{}
	for _, entry := range evidenceFileEntries(value) {
		field := stringValue(entry["field"])
		if field == "" {
			continue
		}
		result[field] = entry
	}
	return result
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
