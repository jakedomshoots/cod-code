package cli

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type productionStatusReport struct {
	Status                 string                        `json:"status"`
	LocalProductionReady   bool                          `json:"local_production_ready"`
	PublicProductionReady  bool                          `json:"public_production_ready"`
	BlockedCount           int                           `json:"blocked_count"`
	BlockedChecks          []string                      `json:"blocked_checks"`
	SummaryPath            string                        `json:"summary_path,omitempty"`
	LaunchChecklist        *productionChecklistStatus    `json:"launch_checklist,omitempty"`
	FinalizerNextActions   *productionChecklistStatus    `json:"finalizer_next_actions,omitempty"`
	ReleaseBootstrap       *productionBootstrapStatus    `json:"release_bootstrap,omitempty"`
	ProviderSetupPreflight *providerSetupPreflightStatus `json:"provider_setup_preflight,omitempty"`
	ExternalSetupRequired  bool                          `json:"external_setup_required"`
	NextAction             string                        `json:"next_action"`
}

type providerSetupPreflightStatus struct {
	Path             string   `json:"path"`
	Status           string   `json:"status"`
	ProviderCount    int      `json:"provider_count"`
	ReadyCount       int      `json:"ready_count"`
	BlockedCount     int      `json:"blocked_count"`
	BlockedProviders []string `json:"blocked_providers,omitempty"`
}

type productionBootstrapStatus struct {
	Path         string `json:"path"`
	HandoffPath  string `json:"handoff_path,omitempty"`
	Status       string `json:"status"`
	BlockedCount int    `json:"blocked_count"`
	Version      string `json:"version,omitempty"`
}

type productionChecklistStatus struct {
	Path                          string         `json:"path"`
	JSONPath                      string         `json:"json_path,omitempty"`
	SetupPath                     string         `json:"setup_path,omitempty"`
	SetupSHA256                   string         `json:"setup_sha256,omitempty"`
	SetupCurrentSHA256            string         `json:"setup_current_sha256,omitempty"`
	SetupMatchesDeclared          *bool          `json:"setup_matches_declared,omitempty"`
	SetupRequiredActionCount      int            `json:"setup_required_action_count,omitempty"`
	SHA256                        string         `json:"sha256,omitempty"`
	CurrentSHA256                 string         `json:"current_sha256,omitempty"`
	MatchesDeclared               *bool          `json:"matches_declared,omitempty"`
	RequiredActionCount           int            `json:"required_action_count"`
	RunnableCommandCount          int            `json:"runnable_command_count"`
	BlockedCommandCount           int            `json:"blocked_command_count"`
	EvidenceDeclaredMatchCount    int            `json:"evidence_declared_match_count"`
	EvidenceDeclaredMismatchCount int            `json:"evidence_declared_mismatch_count"`
	ActionStateCounts             map[string]int `json:"action_state_counts,omitempty"`
	Status                        string         `json:"status"`
}

func runProductionStatus(out io.Writer, opts options) error {
	report, err := buildProductionStatusReport(opts.workspaceDir)
	if err != nil {
		return err
	}
	return writeProductionStatusReport(out, report, opts.reportFormat)
}

func buildProductionStatusReport(workspaceDir string) (productionStatusReport, error) {
	workspace := strings.TrimSpace(workspaceDir)
	if workspace == "" {
		workspace = "."
	}
	evidenceRoot := filepath.Join(workspace, ".omo", "evidence")
	summaryPath, err := latestProductionReadinessSummary(evidenceRoot)
	if err != nil {
		return productionStatusReport{}, err
	}
	if summaryPath == "" {
		return productionStatusReport{
			Status:     "missing",
			NextAction: "run sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness",
		}, nil
	}
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		return productionStatusReport{}, fmt.Errorf("read production summary: %w", err)
	}
	var raw struct {
		Status                string   `json:"status"`
		LocalProductionReady  bool     `json:"local_production_ready"`
		PublicProductionReady bool     `json:"public_production_ready"`
		BlockedCount          int      `json:"blocked_count"`
		BlockedChecks         []string `json:"blocked_checks"`
		LaunchChecklist       *struct {
			Path                string `json:"path"`
			SHA256              string `json:"sha256"`
			RequiredActionCount int    `json:"required_action_count"`
			Status              string `json:"status"`
		} `json:"launch_checklist"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return productionStatusReport{}, fmt.Errorf("decode production summary: %w", err)
	}
	report := productionStatusReport{
		Status:                raw.Status,
		LocalProductionReady:  raw.LocalProductionReady,
		PublicProductionReady: raw.PublicProductionReady,
		BlockedCount:          raw.BlockedCount,
		BlockedChecks:         raw.BlockedChecks,
		SummaryPath:           summaryPath,
		NextAction:            "run sh scripts/production-readiness.sh --dist dist --output-dir .omo/evidence/production-readiness",
	}
	if raw.PublicProductionReady {
		report.NextAction = "public production gate is green"
	} else if raw.LaunchChecklist != nil && raw.LaunchChecklist.Path != "" {
		report.NextAction = "open " + filepath.Join(filepath.Dir(summaryPath), raw.LaunchChecklist.Path)
	}
	if raw.LaunchChecklist != nil {
		checklistPath := filepath.Join(filepath.Dir(summaryPath), raw.LaunchChecklist.Path)
		report.LaunchChecklist = &productionChecklistStatus{
			Path:                raw.LaunchChecklist.Path,
			SHA256:              raw.LaunchChecklist.SHA256,
			RequiredActionCount: raw.LaunchChecklist.RequiredActionCount,
			Status:              raw.LaunchChecklist.Status,
		}
		report.LaunchChecklist.CurrentSHA256, report.LaunchChecklist.MatchesDeclared = checklistFingerprint(checklistPath, raw.LaunchChecklist.SHA256)
	}
	if !raw.PublicProductionReady {
		bootstrap, err := latestReleaseBootstrapStatus(evidenceRoot)
		if err != nil {
			return productionStatusReport{}, err
		}
		report.ReleaseBootstrap = bootstrap
		providerSetup, err := latestProviderSetupPreflightStatus(evidenceRoot)
		if err != nil {
			return productionStatusReport{}, err
		}
		report.ProviderSetupPreflight = providerSetup
		finalizer, err := latestProductionFinalizerNextActions(evidenceRoot)
		if err != nil {
			return productionStatusReport{}, err
		}
		if finalizer != nil {
			report.FinalizerNextActions = finalizer
			report.NextAction = "open " + finalizer.Path
		}
	}
	report.ExternalSetupRequired = productionStatusExternalSetupRequired(report)
	return report, nil
}

func productionStatusExternalSetupRequired(report productionStatusReport) bool {
	if !report.LocalProductionReady || report.PublicProductionReady || report.BlockedCount == 0 {
		return false
	}
	if report.FinalizerNextActions == nil {
		return false
	}
	if report.FinalizerNextActions.RunnableCommandCount != 0 {
		return false
	}
	return report.FinalizerNextActions.BlockedCommandCount > 0
}

func latestProviderSetupPreflightStatus(evidenceRoot string) (*providerSetupPreflightStatus, error) {
	matches, err := filepath.Glob(filepath.Join(evidenceRoot, "provider-setup-preflight*", "summary.json"))
	if err != nil {
		return nil, err
	}
	var latest string
	var latestMod time.Time
	for _, candidate := range matches {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if latest == "" || info.ModTime().After(latestMod) {
			latest = candidate
			latestMod = info.ModTime()
		}
	}
	if latest == "" {
		return nil, nil
	}
	content, err := os.ReadFile(latest)
	if err != nil {
		return nil, fmt.Errorf("read provider setup preflight summary: %w", err)
	}
	var raw struct {
		Status           string   `json:"status"`
		ProviderCount    int      `json:"provider_count"`
		ReadyCount       int      `json:"ready_count"`
		BlockedCount     int      `json:"blocked_count"`
		BlockedProviders []string `json:"blocked_providers"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("decode provider setup preflight summary: %w", err)
	}
	return &providerSetupPreflightStatus{
		Path:             latest,
		Status:           raw.Status,
		ProviderCount:    raw.ProviderCount,
		ReadyCount:       raw.ReadyCount,
		BlockedCount:     raw.BlockedCount,
		BlockedProviders: raw.BlockedProviders,
	}, nil
}

func latestReleaseBootstrapStatus(evidenceRoot string) (*productionBootstrapStatus, error) {
	matches, err := filepath.Glob(filepath.Join(evidenceRoot, "release-bootstrap*", "summary.json"))
	if err != nil {
		return nil, err
	}
	var latest string
	var latestMod time.Time
	for _, candidate := range matches {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if latest == "" || info.ModTime().After(latestMod) {
			latest = candidate
			latestMod = info.ModTime()
		}
	}
	if latest == "" {
		return nil, nil
	}
	content, err := os.ReadFile(latest)
	if err != nil {
		return nil, fmt.Errorf("read release bootstrap summary: %w", err)
	}
	var raw struct {
		Status       string `json:"status"`
		BlockedCount int    `json:"blocked_count"`
		Version      string `json:"version"`
		Artifacts    struct {
			Handoff string `json:"handoff"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("decode release bootstrap summary: %w", err)
	}
	status := &productionBootstrapStatus{
		Path:         latest,
		Status:       raw.Status,
		BlockedCount: raw.BlockedCount,
		Version:      raw.Version,
	}
	if raw.Artifacts.Handoff != "" {
		status.HandoffPath = filepath.Join(filepath.Dir(latest), raw.Artifacts.Handoff)
	}
	return status, nil
}

func latestProductionReadinessSummary(evidenceRoot string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(evidenceRoot, "production-readiness*", "summary.json"))
	if err != nil {
		return "", err
	}
	var latest string
	var latestMod time.Time
	for _, candidate := range matches {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if productionReadinessSummaryHasSkippedChecks(candidate) {
			continue
		}
		if latest == "" || info.ModTime().After(latestMod) {
			latest = candidate
			latestMod = info.ModTime()
		}
	}
	return latest, nil
}

func productionReadinessSummaryHasSkippedChecks(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	var raw struct {
		Checks []struct {
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return true
	}
	for _, check := range raw.Checks {
		if check.Status == "skipped" {
			return true
		}
	}
	return false
}

func latestProductionFinalizerNextActions(evidenceRoot string) (*productionChecklistStatus, error) {
	matches, err := filepath.Glob(filepath.Join(evidenceRoot, "production-finalize*", "summary.json"))
	if err != nil {
		return nil, err
	}
	var latest string
	var latestMod time.Time
	for _, candidate := range matches {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if productionFinalizerSummaryHasSkippedSteps(candidate) {
			continue
		}
		if latest == "" || info.ModTime().After(latestMod) {
			latest = candidate
			latestMod = info.ModTime()
		}
	}
	if latest == "" {
		return nil, nil
	}
	content, err := os.ReadFile(latest)
	if err != nil {
		return nil, fmt.Errorf("read production finalizer summary: %w", err)
	}
	var raw struct {
		NextActions *struct {
			Path                string `json:"path"`
			JSONPath            string `json:"json_path"`
			RequiredActionCount int    `json:"required_action_count"`
		} `json:"next_actions"`
		SetupActions *struct {
			Path                string `json:"path"`
			SHA256              string `json:"sha256"`
			RequiredActionCount int    `json:"required_action_count"`
		} `json:"setup_actions"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("decode production finalizer summary: %w", err)
	}
	if raw.NextActions == nil || raw.NextActions.Path == "" {
		return nil, nil
	}
	finalizerDir := filepath.Dir(latest)
	status := &productionChecklistStatus{
		Path:                filepath.Join(finalizerDir, raw.NextActions.Path),
		RequiredActionCount: raw.NextActions.RequiredActionCount,
		Status:              "pass",
	}
	if raw.NextActions.JSONPath != "" {
		status.JSONPath = filepath.Join(finalizerDir, raw.NextActions.JSONPath)
		if summary, err := productionActionSummaryFromJSON(status.JSONPath); err == nil {
			status.ActionStateCounts = summary.StateCounts
			status.RunnableCommandCount = summary.RunnableCommandCount
			status.BlockedCommandCount = summary.BlockedCommandCount
			status.EvidenceDeclaredMatchCount = summary.EvidenceDeclaredMatchCount
			status.EvidenceDeclaredMismatchCount = summary.EvidenceDeclaredMismatchCount
		}
	}
	if raw.SetupActions != nil && raw.SetupActions.Path != "" {
		status.SetupPath = filepath.Join(finalizerDir, raw.SetupActions.Path)
		status.SetupSHA256 = raw.SetupActions.SHA256
		status.SetupCurrentSHA256, status.SetupMatchesDeclared = checklistFingerprint(status.SetupPath, raw.SetupActions.SHA256)
		status.SetupRequiredActionCount = raw.SetupActions.RequiredActionCount
	}
	return status, nil
}

func checklistFingerprint(path string, declared string) (string, *bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	sum := sha256.Sum256(content)
	current := fmt.Sprintf("%x", sum[:])
	if declared == "" {
		return current, nil
	}
	matches := current == declared
	return current, &matches
}

type productionActionSummary struct {
	StateCounts                   map[string]int
	RunnableCommandCount          int
	BlockedCommandCount           int
	EvidenceDeclaredMatchCount    int
	EvidenceDeclaredMismatchCount int
}

func productionActionSummaryFromJSON(path string) (productionActionSummary, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return productionActionSummary{}, err
	}
	var raw struct {
		Actions []map[string]any `json:"actions"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return productionActionSummary{}, err
	}
	annotated := annotateProductionActions(raw.Actions, filepath.Dir(path))
	matches, mismatches := countProductionEvidenceDeclaredMatches(annotated)
	return productionActionSummary{
		StateCounts:                   countProductionActionStates(annotated),
		RunnableCommandCount:          countRunnableProductionActionCommands(annotated),
		BlockedCommandCount:           countBlockedProductionActionCommands(annotated),
		EvidenceDeclaredMatchCount:    matches,
		EvidenceDeclaredMismatchCount: mismatches,
	}, nil
}

func countProductionEvidenceDeclaredMatches(actions []map[string]any) (int, int) {
	matches := 0
	mismatches := 0
	for _, action := range actions {
		for _, file := range evidenceFileEntries(action["evidence_files"]) {
			match, ok := file["matches_declared"].(bool)
			if !ok {
				continue
			}
			if match {
				matches++
			} else {
				mismatches++
			}
		}
	}
	return matches, mismatches
}

func productionFinalizerSummaryHasSkippedSteps(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	var raw struct {
		Status       string   `json:"status"`
		SkippedSteps []string `json:"skipped_steps"`
	}
	if err := json.Unmarshal(content, &raw); err != nil {
		return true
	}
	return len(raw.SkippedSteps) > 0
}

func writeProductionStatusReport(out io.Writer, report productionStatusReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := fmt.Fprint(out, renderProductionStatusText(report))
		return err
	case reportFormatEvents:
		encoder := json.NewEncoder(out)
		return encoder.Encode(struct {
			Kind   string                 `json:"kind"`
			Status productionStatusReport `json:"status"`
		}{Kind: "production_status", Status: report})
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderProductionStatusText(report productionStatusReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Production status: %s\n", report.Status)
	fmt.Fprintf(&builder, "Local ready: %t\n", report.LocalProductionReady)
	fmt.Fprintf(&builder, "Public ready: %t\n", report.PublicProductionReady)
	if report.ExternalSetupRequired {
		fmt.Fprintf(&builder, "External setup required: true\n")
	}
	fmt.Fprintf(&builder, "Blocked checks: %d\n", report.BlockedCount)
	for _, check := range report.BlockedChecks {
		fmt.Fprintf(&builder, "- %s\n", check)
	}
	if report.LaunchChecklist != nil {
		fmt.Fprintf(&builder, "Launch checklist: %s (%d actions)", report.LaunchChecklist.Path, report.LaunchChecklist.RequiredActionCount)
		if report.LaunchChecklist.MatchesDeclared != nil {
			fmt.Fprintf(&builder, " declared_match=%t", *report.LaunchChecklist.MatchesDeclared)
		}
		builder.WriteString("\n")
	}
	if report.FinalizerNextActions != nil {
		fmt.Fprintf(&builder, "Finalizer next actions: %s (%d actions)\n", report.FinalizerNextActions.Path, report.FinalizerNextActions.RequiredActionCount)
		if report.FinalizerNextActions.JSONPath != "" {
			fmt.Fprintf(&builder, "Finalizer actions JSON: %s\n", report.FinalizerNextActions.JSONPath)
		}
		if len(report.FinalizerNextActions.ActionStateCounts) > 0 {
			fmt.Fprintf(&builder, "Finalizer action states: %s\n", renderActionStateCounts(report.FinalizerNextActions.ActionStateCounts))
		}
		fmt.Fprintf(&builder, "Finalizer commands: runnable=%d blocked=%d\n", report.FinalizerNextActions.RunnableCommandCount, report.FinalizerNextActions.BlockedCommandCount)
		fmt.Fprintf(&builder, "Finalizer evidence matches: declared=%d mismatched=%d\n", report.FinalizerNextActions.EvidenceDeclaredMatchCount, report.FinalizerNextActions.EvidenceDeclaredMismatchCount)
		if report.FinalizerNextActions.SetupPath != "" {
			fmt.Fprintf(&builder, "Finalizer setup actions: %s", report.FinalizerNextActions.SetupPath)
			if report.FinalizerNextActions.SetupRequiredActionCount > 0 {
				fmt.Fprintf(&builder, " (%d actions)", report.FinalizerNextActions.SetupRequiredActionCount)
			}
			if report.FinalizerNextActions.SetupSHA256 != "" {
				fmt.Fprintf(&builder, " sha256=%s", report.FinalizerNextActions.SetupSHA256)
			}
			if report.FinalizerNextActions.SetupMatchesDeclared != nil {
				fmt.Fprintf(&builder, " declared_match=%t", *report.FinalizerNextActions.SetupMatchesDeclared)
			}
			builder.WriteString("\n")
		}
	}
	if report.ReleaseBootstrap != nil {
		fmt.Fprintf(&builder, "Release bootstrap: %s blocked=%d", report.ReleaseBootstrap.Status, report.ReleaseBootstrap.BlockedCount)
		if report.ReleaseBootstrap.Version != "" {
			fmt.Fprintf(&builder, " version=%s", report.ReleaseBootstrap.Version)
		}
		builder.WriteString("\n")
		if report.ReleaseBootstrap.HandoffPath != "" {
			fmt.Fprintf(&builder, "Release handoff: %s\n", report.ReleaseBootstrap.HandoffPath)
		}
	}
	if report.ProviderSetupPreflight != nil {
		fmt.Fprintf(
			&builder,
			"Provider setup preflight: %s ready=%d blocked=%d providers=%d",
			report.ProviderSetupPreflight.Status,
			report.ProviderSetupPreflight.ReadyCount,
			report.ProviderSetupPreflight.BlockedCount,
			report.ProviderSetupPreflight.ProviderCount,
		)
		if len(report.ProviderSetupPreflight.BlockedProviders) > 0 {
			fmt.Fprintf(&builder, " blocked_providers=%s", strings.Join(report.ProviderSetupPreflight.BlockedProviders, ","))
		}
		fmt.Fprintf(&builder, " path=%s\n", report.ProviderSetupPreflight.Path)
	}
	fmt.Fprintf(&builder, "Next action: %s\n", report.NextAction)
	return builder.String()
}
