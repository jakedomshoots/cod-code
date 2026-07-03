package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ScoreReport(ctx context.Context, req ScoreRequest) (ScoreResult, error) {
	if err := ctx.Err(); err != nil {
		return ScoreResult{}, err
	}
	report, err := loadReport(req.ReportPath)
	if err != nil {
		return ScoreResult{}, err
	}
	if report.TaskID != req.Task.ID {
		return ScoreResult{}, fmt.Errorf("%w: report task_id %q does not match %q", ErrInvalidReport, report.TaskID, req.Task.ID)
	}
	checks := make([]CheckResult, 0)
	checks = appendChangedFileChecks(checks, req.Task, report)
	checks = appendForbiddenFileChecks(checks, req.Task, report)
	checks = appendCommandChecks(checks, req.Task, report)
	checks = appendArtifactChecks(checks, req, report)
	checks = appendDiffChecks(checks, req.Task, report)
	checks = appendReportFieldChecks(checks, req.Task, report)
	checks = appendDirtyWorktreeChecks(ctx, checks, req, report)

	passed := countPassedChecks(checks)
	return ScoreResult{
		TaskID:        req.Task.ID,
		Verdict:       scoreVerdict(passed, len(checks)),
		Passed:        passed,
		Total:         len(checks),
		Checks:        checks,
		EvidencePaths: append([]string(nil), report.EvidencePaths...),
	}, nil
}

func appendChangedFileChecks(checks []CheckResult, task Task, report savedReport) []CheckResult {
	for _, required := range task.RequiredChangedFiles {
		checks = append(checks, membershipCheck("changed_file:"+required, required, report.ChangedFiles))
	}
	return checks
}

func appendForbiddenFileChecks(checks []CheckResult, task Task, report savedReport) []CheckResult {
	for _, forbidden := range task.ForbiddenChangedFiles {
		if stringInSlice(forbidden, report.ChangedFiles) {
			checks = append(checks, CheckResult{Name: "forbidden_file:" + forbidden, Status: "fail", Message: "forbidden file was changed"})
			continue
		}
		checks = append(checks, CheckResult{Name: "forbidden_file:" + forbidden, Status: "pass"})
	}
	return checks
}

func appendCommandChecks(checks []CheckResult, task Task, report savedReport) []CheckResult {
	for _, command := range task.RequiredCommands {
		status := "fail"
		evidence := ""
		for _, result := range report.CheckResults {
			if strings.Join(result.Argv, " ") == command && result.Status == "pass" && result.ExitCode == 0 {
				status = "pass"
				evidence = strings.TrimSpace(result.Stdout)
				break
			}
		}
		checks = append(checks, CheckResult{Name: "command:" + command, Status: status, Evidence: evidence})
	}
	return checks
}

func appendArtifactChecks(checks []CheckResult, req ScoreRequest, report savedReport) []CheckResult {
	for _, artifact := range req.Task.RequiredArtifacts {
		status := "fail"
		if stringInSlice(artifact, report.EvidencePaths) && artifactExists(req, artifact) {
			status = "pass"
		}
		checks = append(checks, CheckResult{Name: "artifact:" + artifact, Status: status, Evidence: artifact})
	}
	return checks
}

func appendDiffChecks(checks []CheckResult, task Task, report savedReport) []CheckResult {
	diffText := strings.ToLower(strings.Join(reportDiffs(report), "\n"))
	for _, term := range task.RequiredDiffTerms {
		status := "fail"
		if strings.Contains(diffText, strings.ToLower(term)) {
			status = "pass"
		}
		checks = append(checks, CheckResult{Name: "diff_term:" + term, Status: status})
	}
	return checks
}

func appendReportFieldChecks(checks []CheckResult, task Task, report savedReport) []CheckResult {
	for _, field := range task.RequiredReportFields {
		status := "fail"
		if report.hasReportField(field) {
			status = "pass"
		}
		checks = append(checks, CheckResult{Name: "report_field:" + field, Status: status})
	}
	return checks
}

func artifactExists(req ScoreRequest, artifact string) bool {
	clean, ok := cleanRelativeArtifactPath(artifact)
	if !ok {
		return false
	}
	roots := []string{filepath.Dir(req.ReportPath)}
	if strings.TrimSpace(req.WorkspaceDir) != "" {
		roots = append(roots, req.WorkspaceDir)
	}
	for _, root := range roots {
		info, err := os.Stat(filepath.Join(root, clean))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func cleanRelativeArtifactPath(path string) (string, bool) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", false
	}
	return clean, true
}

func membershipCheck(name string, want string, values []string) CheckResult {
	if stringInSlice(want, values) {
		return CheckResult{Name: name, Status: "pass", Evidence: want}
	}
	return CheckResult{Name: name, Status: "fail", Message: "missing required value"}
}

func stringInSlice(want string, values []string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func reportDiffs(report savedReport) []string {
	diffs := make([]string, 0, len(report.PatchResults)+len(report.PatchPreviews))
	for _, patch := range report.PatchResults {
		diffs = append(diffs, patch.Diff)
	}
	for _, patch := range report.PatchPreviews {
		diffs = append(diffs, patch.Diff)
	}
	return diffs
}

func countPassedChecks(checks []CheckResult) int {
	passed := 0
	for _, check := range checks {
		if check.Status == "pass" {
			passed++
		}
	}
	return passed
}

func scoreVerdict(passed int, total int) string {
	if total > 0 && passed == total {
		return "pass"
	}
	if passed > 0 {
		return "partial"
	}
	return "fail"
}
