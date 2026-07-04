package eval

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	competitorSmokeMode          = "local_version_and_dry_run_smoke"
	competitorSmokeStatusPass    = "smoke_pass"
	competitorSmokeStatusFail    = "smoke_failed"
	competitorSmokeStatusBlocked = "setup_blocked"
	competitorSmokeStatusSkipped = "skipped_missing_binary"
)

func RunCompetitorSmoke(ctx context.Context, req CompetitorSmokeRequest) (CompetitorSmokeSummary, error) {
	config, err := LoadCompetitors(req.CompetitorsPath)
	if err != nil {
		return CompetitorSmokeSummary{}, err
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return CompetitorSmokeSummary{}, fmt.Errorf("create competitor smoke output dir: %w", err)
	}
	summary := CompetitorSmokeSummary{
		SchemaVersion: config.SchemaVersion,
		Mode:          competitorSmokeMode,
		Competitors:   len(config.Competitors),
		Results:       make([]CompetitorSmokeResult, 0, len(config.Competitors)),
	}
	for _, competitor := range config.Competitors {
		result := runCompetitorSmoke(ctx, req, competitor)
		summary.Results = append(summary.Results, result)
		switch result.Status {
		case competitorSmokeStatusPass:
			summary.SmokePassed++
		case competitorSmokeStatusBlocked:
			summary.SetupBlocked++
		case competitorSmokeStatusSkipped:
			summary.Skipped++
		default:
			summary.SmokeFailed++
		}
	}
	if summary.SetupBlocked > 0 || summary.Skipped > 0 || summary.SmokeFailed > 0 {
		summary.SetupActions = "setup-actions.md"
	}
	if err := writeJSONFile(filepath.Join(req.OutputDir, "summary.json"), summary); err != nil {
		return CompetitorSmokeSummary{}, err
	}
	if summary.SetupActions != "" {
		if err := writeCompetitorSetupActions(filepath.Join(req.OutputDir, summary.SetupActions), summary); err != nil {
			return CompetitorSmokeSummary{}, err
		}
	}
	return summary, nil
}

func writeCompetitorSetupActions(path string, summary CompetitorSmokeSummary) error {
	var builder strings.Builder
	builder.WriteString("# Competitor Setup Actions\n\n")
	builder.WriteString("Run these before the full all-agent comparison.\n\n")
	for _, result := range summary.Results {
		switch result.Status {
		case competitorSmokeStatusPass:
			continue
		case competitorSmokeStatusSkipped:
			fmt.Fprintf(&builder, "- %s: install/authenticate `%s` before comparison. %s\n", result.ID, result.Binary, setupReason(result))
		case competitorSmokeStatusBlocked:
			fmt.Fprintf(&builder, "- %s: fix provider auth/quota for `%s`. %s\n", result.ID, result.Binary, setupReason(result))
		default:
			fmt.Fprintf(&builder, "- %s: fix smoke failure for `%s`. %s\n", result.ID, result.Binary, setupReason(result))
		}
	}
	builder.WriteString("\nAfter setup is clean, rerun:\n\n")
	builder.WriteString("```sh\n")
	builder.WriteString("ceo-packet production-finalize --workspace . --dry-run\n")
	builder.WriteString("ceo-packet production-finalize --workspace . --run-comparison\n")
	builder.WriteString("```\n")
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func setupReason(result CompetitorSmokeResult) string {
	if strings.TrimSpace(result.SetupHint) != "" {
		return result.SetupHint
	}
	if strings.TrimSpace(result.Note) != "" {
		return result.Note
	}
	if strings.TrimSpace(result.DryRun.Error) != "" {
		return result.DryRun.Error
	}
	if strings.TrimSpace(result.Version.Error) != "" {
		return result.Version.Error
	}
	return "Inspect summary.json and saved stdout/stderr."
}

func runCompetitorSmoke(ctx context.Context, req CompetitorSmokeRequest, competitor Competitor) CompetitorSmokeResult {
	result := CompetitorSmokeResult{
		ID:     competitor.ID,
		Name:   competitor.Name,
		Binary: competitor.Binary,
		Status: competitorSmokeStatusSkipped,
		Version: SmokeCommandResult{
			Command:  append([]string{competitor.Binary}, competitor.VersionArgs...),
			ExitCode: -1,
			Error:    "binary not found on PATH",
		},
		DryRun: SmokeCommandResult{
			Command:  append([]string{competitor.Binary}, competitor.DryRunArgs...),
			ExitCode: -1,
			Error:    "binary not found on PATH",
		},
		SetupHint: competitor.SetupHint,
		Note:      "binary not found on PATH; skipped instead of failed",
	}
	resolved, err := exec.LookPath(competitor.Binary)
	if err != nil {
		return result
	}
	result.ResolvedPath = resolved
	result.SetupHint = ""
	result.Note = "ran local version and dry-run commands only; no head-to-head benchmark task was attempted"

	resultDir := filepath.Join(req.OutputDir, competitor.ID)
	version := runSmokeCommand(ctx, resolved, competitor.VersionArgs, req.TimeoutSeconds)
	dryRun := runSmokeCommand(ctx, resolved, competitor.DryRunArgs, req.TimeoutSeconds)
	result.Version = version
	result.DryRun = dryRun
	evidencePaths, err := writeSmokeEvidence(resultDir, version, dryRun)
	if err != nil {
		result.Status = competitorSmokeStatusFail
		result.Note = err.Error()
		return result
	}
	result.EvidencePaths = evidencePaths
	if version.ExitCode == 0 && dryRun.ExitCode == 0 && version.Error == "" && dryRun.Error == "" {
		result.Status = competitorSmokeStatusPass
		return result
	}
	if smokeSetupBlocked(version) || smokeSetupBlocked(dryRun) {
		result.Status = competitorSmokeStatusBlocked
		result.Note = "provider setup is blocked; smoke stdout/stderr captured auth, quota, or credential evidence"
		return result
	}
	result.Status = competitorSmokeStatusFail
	return result
}

func smokeSetupBlocked(result SmokeCommandResult) bool {
	run := localAgentRunResult{
		stdout:  result.Stdout,
		stderr:  result.Stderr,
		errText: result.Error,
	}
	return localAgentSetupBlocked(run)
}

func runSmokeCommand(ctx context.Context, binary string, args []string, timeoutSeconds int) SmokeCommandResult {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := SmokeCommandResult{
		Command:  append([]string{binary}, args...),
		ExitCode: commandExitCode(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
	if err != nil {
		result.Error = err.Error()
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			result.Error = "command timed out"
		}
	}
	return result
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func writeSmokeEvidence(resultDir string, version SmokeCommandResult, dryRun SmokeCommandResult) ([]string, error) {
	paths := []string{
		filepath.Join(resultDir, "version.stdout"),
		filepath.Join(resultDir, "version.stderr"),
		filepath.Join(resultDir, "dry-run.stdout"),
		filepath.Join(resultDir, "dry-run.stderr"),
	}
	contents := []string{version.Stdout, version.Stderr, dryRun.Stdout, dryRun.Stderr}
	for index, path := range paths {
		content := contents[index]
		if strings.TrimSpace(content) == "" {
			content = "(empty)\n"
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("create parent for %s: %w", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
	}
	return paths, nil
}
