package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func runProductionFinalize(ctx context.Context, out io.Writer, opts options) error {
	workspaceDir := opts.workspaceDir
	if workspaceDir == "" {
		var err error
		workspaceDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve working directory: %w", err)
		}
	}
	scriptPath := filepath.Join(workspaceDir, "scripts", "production-finalize.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("production-finalize requires %s", scriptPath)
	}

	args := []string{scriptPath}
	if opts.dryRun {
		args = append(args, "--dry-run")
	}
	if opts.productionFinalizeRunComparison {
		args = append(args, "--run-comparison")
	}
	if opts.completionOutputDir != "" {
		args = append(args, "--output-dir", opts.completionOutputDir)
	}
	if opts.productionFinalizeEvidenceRoot != "" {
		args = append(args, "--evidence-root", opts.productionFinalizeEvidenceRoot)
	}
	if opts.productionFinalizeDist != "" {
		args = append(args, "--dist", opts.productionFinalizeDist)
	}
	if opts.productionFinalizeProviderTimeoutSeconds > 0 {
		args = append(args, "--provider-timeout-seconds", strconv.Itoa(opts.productionFinalizeProviderTimeoutSeconds))
	}
	if opts.productionFinalizeComparisonTimeoutSeconds > 0 {
		args = append(args, "--comparison-timeout-seconds", strconv.Itoa(opts.productionFinalizeComparisonTimeoutSeconds))
	}
	if opts.productionFinalizeComparisonTimeoutRetries > 0 {
		args = append(args, "--comparison-timeout-retries", strconv.Itoa(opts.productionFinalizeComparisonTimeoutRetries))
	}
	if opts.productionFinalizeComparisonResultRetries > 0 {
		args = append(args, "--comparison-result-retries", strconv.Itoa(opts.productionFinalizeComparisonResultRetries))
	}

	cmd := exec.CommandContext(ctx, "sh", args...)
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	if _, writeErr := out.Write(output); writeErr != nil {
		return writeErr
	}
	if err != nil {
		return fmt.Errorf("production-finalize failed: %w", err)
	}
	return nil
}
