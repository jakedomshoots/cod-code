package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"ceoharness/internal/ceo"
	"ceoharness/internal/workspace"
)

type rollbackReport struct {
	ReportPath  string   `json:"report_path"`
	RolledBack  int      `json:"rolled_back"`
	RolledFiles []string `json:"rolled_files"`
}

func runRollbackReport(ctx context.Context, out io.Writer, opts options) error {
	raw, err := os.ReadFile(opts.rollbackReportPath)
	if err != nil {
		return fmt.Errorf("read rollback report: %w", err)
	}
	var report ceo.Report
	if err := json.Unmarshal(raw, &report); err != nil {
		return fmt.Errorf("decode rollback report: %w", err)
	}
	space, err := workspace.New(opts.workspaceDir)
	if err != nil {
		return err
	}
	rolledFiles := make([]string, 0, len(report.PatchResults))
	for _, patch := range report.PatchResults {
		result, err := space.RollbackReplaceText(ctx, patch)
		if err != nil {
			return fmt.Errorf("rollback %s: %w", patch.Path, err)
		}
		rolledFiles = append(rolledFiles, result.Path)
	}
	response := rollbackReport{
		ReportPath:  opts.rollbackReportPath,
		RolledBack:  len(rolledFiles),
		RolledFiles: rolledFiles,
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("write rollback report: %w", err)
	}
	return nil
}
