package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	comparisonModePlanOnly      = "plan_only"
	comparisonStatusPlanned     = "planned_no_result"
	comparisonStatusMissingTool = "skipped_missing_binary"
)

var requiredCompetitorIDs = []string{
	"codex_cli",
	"claude_code",
	"aider",
	"opencode",
	"goose",
}

var requiredComparisonDimensions = []string{
	"task_success",
	"time_to_complete",
	"files_changed",
	"safety_prompts",
	"cost_provider_used",
	"evidence_quality",
}

func LoadCompetitors(path string) (CompetitorConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return CompetitorConfig{}, fmt.Errorf("read competitors %s: %w", path, err)
	}
	var config CompetitorConfig
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return CompetitorConfig{}, fmt.Errorf("%w: decode competitors %s: %w", ErrInvalidCompetitor, path, err)
	}
	normalizeCompetitors(&config)
	if err := validateCompetitorConfig(config); err != nil {
		return CompetitorConfig{}, err
	}
	return config, nil
}

func BuildComparisonPlan(ctx context.Context, config CompetitorConfig) (ComparisonPlan, error) {
	if err := ctx.Err(); err != nil {
		return ComparisonPlan{}, err
	}
	if err := validateCompetitorConfig(config); err != nil {
		return ComparisonPlan{}, err
	}
	results := make([]ComparisonResult, 0, len(config.Competitors))
	for _, competitor := range config.Competitors {
		if err := ctx.Err(); err != nil {
			return ComparisonPlan{}, err
		}
		results = append(results, plannedComparisonResult(competitor))
	}
	return ComparisonPlan{
		SchemaVersion: config.SchemaVersion,
		Mode:          comparisonModePlanOnly,
		Results:       results,
	}, nil
}

func plannedComparisonResult(competitor Competitor) ComparisonResult {
	status := comparisonStatusPlanned
	note := "empty placeholder only; no pass/fail result exists until command logs and evidence are saved"
	if _, err := exec.LookPath(competitor.Binary); err != nil {
		status = comparisonStatusMissingTool
		note = "binary not found on PATH; skipped instead of failed"
	}
	return ComparisonResult{
		ID:                   competitor.ID,
		Name:                 competitor.Name,
		Status:               status,
		Binary:               competitor.Binary,
		Command:              append([]string{competitor.Binary}, competitor.DryRunArgs...),
		TimeoutSeconds:       competitor.TimeoutSeconds,
		ComparisonDimensions: append([]string(nil), competitor.ComparisonDimensions...),
		EvidencePaths:        []string{},
		Note:                 note,
	}
}

func normalizeCompetitors(config *CompetitorConfig) {
	for index := range config.Competitors {
		competitor := &config.Competitors[index]
		competitor.ID = strings.TrimSpace(competitor.ID)
		competitor.Name = strings.TrimSpace(competitor.Name)
		competitor.Binary = strings.TrimSpace(competitor.Binary)
		competitor.Homepage = strings.TrimSpace(competitor.Homepage)
		competitor.SetupHint = strings.TrimSpace(competitor.SetupHint)
		competitor.VersionArgs = nonEmptyArgs(competitor.VersionArgs)
		competitor.DryRunArgs = nonEmptyArgs(competitor.DryRunArgs)
		competitor.ComparisonDimensions = nonEmptyArgs(competitor.ComparisonDimensions)
	}
}

func validateCompetitorConfig(config CompetitorConfig) error {
	if config.SchemaVersion != 1 {
		return fmt.Errorf("%w: schema_version must be 1", ErrInvalidCompetitor)
	}
	if len(config.Competitors) != len(requiredCompetitorIDs) {
		return fmt.Errorf("%w: expected %d competitors, got %d", ErrInvalidCompetitor, len(requiredCompetitorIDs), len(config.Competitors))
	}
	seen := map[string]struct{}{}
	for _, competitor := range config.Competitors {
		if err := validateCompetitor(competitor); err != nil {
			return err
		}
		if _, ok := seen[competitor.ID]; ok {
			return fmt.Errorf("%w: duplicate competitor id %q", ErrInvalidCompetitor, competitor.ID)
		}
		seen[competitor.ID] = struct{}{}
	}
	for _, id := range requiredCompetitorIDs {
		if _, ok := seen[id]; !ok {
			return fmt.Errorf("%w: missing competitor %q", ErrInvalidCompetitor, id)
		}
	}
	return nil
}

func validateCompetitor(competitor Competitor) error {
	if competitor.ID == "" {
		return fmt.Errorf("%w: competitor id is required", ErrInvalidCompetitor)
	}
	if strings.TrimSpace(competitor.Name) == "" || strings.TrimSpace(competitor.Binary) == "" {
		return fmt.Errorf("%w: competitor %s needs name and binary", ErrInvalidCompetitor, competitor.ID)
	}
	if strings.TrimSpace(competitor.Homepage) == "" || strings.TrimSpace(competitor.SetupHint) == "" {
		return fmt.Errorf("%w: competitor %s needs homepage and setup_hint", ErrInvalidCompetitor, competitor.ID)
	}
	if len(nonEmptyArgs(competitor.VersionArgs)) == 0 || len(nonEmptyArgs(competitor.DryRunArgs)) == 0 {
		return fmt.Errorf("%w: competitor %s needs version_args and dry_run_args", ErrInvalidCompetitor, competitor.ID)
	}
	if competitor.TimeoutSeconds <= 0 {
		return fmt.Errorf("%w: competitor %s needs positive timeout_seconds", ErrInvalidCompetitor, competitor.ID)
	}
	for _, dimension := range requiredComparisonDimensions {
		if !stringInSlice(dimension, competitor.ComparisonDimensions) {
			return fmt.Errorf("%w: competitor %s missing dimension %q", ErrInvalidCompetitor, competitor.ID, dimension)
		}
	}
	return nil
}

func nonEmptyArgs(values []string) []string {
	args := make([]string, 0, len(values))
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean != "" {
			args = append(args, clean)
		}
	}
	return args
}
