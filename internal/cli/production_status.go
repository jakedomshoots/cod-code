package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type productionStatusReport struct {
	Status                string                     `json:"status"`
	LocalProductionReady  bool                       `json:"local_production_ready"`
	PublicProductionReady bool                       `json:"public_production_ready"`
	BlockedCount          int                        `json:"blocked_count"`
	BlockedChecks         []string                   `json:"blocked_checks"`
	SummaryPath           string                     `json:"summary_path,omitempty"`
	LaunchChecklist       *productionChecklistStatus `json:"launch_checklist,omitempty"`
	NextAction            string                     `json:"next_action"`
}

type productionChecklistStatus struct {
	Path                string `json:"path"`
	SHA256              string `json:"sha256,omitempty"`
	RequiredActionCount int    `json:"required_action_count"`
	Status              string `json:"status"`
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
		report.LaunchChecklist = &productionChecklistStatus{
			Path:                raw.LaunchChecklist.Path,
			SHA256:              raw.LaunchChecklist.SHA256,
			RequiredActionCount: raw.LaunchChecklist.RequiredActionCount,
			Status:              raw.LaunchChecklist.Status,
		}
	}
	return report, nil
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
		if latest == "" || info.ModTime().After(latestMod) {
			latest = candidate
			latestMod = info.ModTime()
		}
	}
	return latest, nil
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
	fmt.Fprintf(&builder, "Blocked checks: %d\n", report.BlockedCount)
	for _, check := range report.BlockedChecks {
		fmt.Fprintf(&builder, "- %s\n", check)
	}
	if report.LaunchChecklist != nil {
		fmt.Fprintf(&builder, "Launch checklist: %s (%d actions)\n", report.LaunchChecklist.Path, report.LaunchChecklist.RequiredActionCount)
	}
	fmt.Fprintf(&builder, "Next action: %s\n", report.NextAction)
	return builder.String()
}
