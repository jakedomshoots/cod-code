package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/history"
)

type doctorReport struct {
	Status  string        `json:"status"`
	Summary string        `json:"summary"`
	Version string        `json:"version"`
	Checks  []doctorCheck `json:"checks"`
}

type doctorCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Requirement string `json:"requirement,omitempty"`
	Source      string `json:"source,omitempty"`
	Verdict     string `json:"verdict,omitempty"`
	Workspace   string `json:"workspace,omitempty"`
	PatchCount  int    `json:"patch_count,omitempty"`
	CheckCount  int    `json:"check_count,omitempty"`
	EventCount  int    `json:"event_count,omitempty"`
	JobID       string `json:"job_id,omitempty"`
	IssueKind   string `json:"issue_kind,omitempty"`
	Path        string `json:"path,omitempty"`
	Guidance    string `json:"guidance,omitempty"`
	Error       string `json:"error,omitempty"`
}

func runDoctor(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildDoctorReport(ctx, opts)
	if writeErr := writeDoctorReport(out, report, opts.reportFormat); writeErr != nil {
		return writeErr
	}
	if report.Status != "pass" {
		if err != nil {
			return err
		}
		return ErrVerdictFailed
	}
	return nil
}

func buildDoctorReport(ctx context.Context, opts options) (doctorReport, error) {
	demo, err := runDemoReport(ctx, opts)
	report := doctorReportFromDemo(demo, err)
	for _, check := range runLocalDoctorChecks(ctx, opts) {
		appendDoctorCheck(&report, check)
	}
	if check, ok := runModelCommandDoctorCheck(ctx, opts); ok {
		appendDoctorCheck(&report, check)
	}
	if check, ok := runCEOModelCommandDoctorCheck(ctx, opts); ok {
		appendDoctorCheck(&report, check)
	}
	verificationCheck := runVerificationPolicyDoctorCheck(ctx, opts)
	if verificationCheck.Status != "skipped" {
		appendDoctorCheck(&report, verificationCheck)
	}
	if check, ok := runJobRecoveryDoctorCheck(ctx, opts); ok {
		appendDoctorCheck(&report, check)
	}
	for _, check := range runProviderDoctorChecks(ctx, opts) {
		appendDoctorCheck(&report, check)
	}
	if check, ok := runResearchCommandDoctorCheck(ctx, opts); ok {
		appendDoctorCheck(&report, check)
	}
	return report, err
}

func failedDoctorCheck(name string, err error) doctorCheck {
	return doctorCheck{Name: name, Status: "fail", Error: err.Error()}
}

func runJobRecoveryDoctorCheck(ctx context.Context, opts options) (doctorCheck, bool) {
	if strings.TrimSpace(opts.workspaceDir) == "" {
		return doctorCheck{}, false
	}
	store, err := history.New(opts.workspaceDir)
	if err != nil {
		return failedDoctorCheck("job_recovery", err), true
	}
	issues, err := store.InspectReportRecovery(ctx)
	if err != nil {
		return failedDoctorCheck("job_recovery", err), true
	}
	if len(issues) == 0 {
		return doctorCheck{Name: "job_recovery", Status: "pass"}, true
	}
	issue := issues[0]
	return doctorCheck{
		Name:      "job_recovery",
		Status:    "fail",
		JobID:     issue.JobID,
		IssueKind: issue.Kind,
		Path:      issue.Path,
		Guidance:  issue.Guidance,
		Error:     issue.Error,
	}, true
}

func doctorReportFromDemo(demo demoRun, err error) doctorReport {
	check := doctorCheck{Name: "golden_demo"}
	if err != nil {
		check.Status = "fail"
		check.Error = err.Error()
		return doctorReport{
			Status:  "fail",
			Summary: "harness doctor failed",
			Version: versionDetails(),
			Checks:  []doctorCheck{check},
		}
	}
	check.Status = demo.Report.Verdict
	check.Verdict = demo.Report.Verdict
	check.Workspace = demo.WorkspaceDir
	check.PatchCount = len(demo.Report.PatchResults)
	check.CheckCount = len(demo.Report.CheckResults)
	check.EventCount = len(demo.Report.RunEvents)
	status := "pass"
	if check.Status != "pass" || check.PatchCount == 0 || check.CheckCount == 0 {
		status = "fail"
		check.Status = "fail"
	}
	return doctorReport{
		Status:  status,
		Summary: "harness doctor " + status,
		Version: versionDetails(),
		Checks:  []doctorCheck{check},
	}
}

func writeDoctorReport(out io.Writer, report doctorReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := io.WriteString(out, renderDoctorText(report))
		return err
	case reportFormatEvents:
		return writeDoctorEvents(out, report)
	default:
		return errors.New(reportFormatGuidance)
	}
}

func renderDoctorText(report doctorReport) string {
	var builder strings.Builder
	builder.WriteString("Doctor: " + report.Status + "\n")
	builder.WriteString("Version: " + report.Version + "\n")
	for _, check := range report.Checks {
		builder.WriteString(fmt.Sprintf(
			"- %s [%s]: requirement=%s source=%s verdict=%s patches=%d checks=%d workspace=%s\n",
			check.Name,
			check.Status,
			check.Requirement,
			check.Source,
			check.Verdict,
			check.PatchCount,
			check.CheckCount,
			check.Workspace,
		))
		if check.Error != "" {
			builder.WriteString("  error: " + check.Error + "\n")
		}
		if check.Guidance != "" {
			builder.WriteString("  guidance: " + check.Guidance + "\n")
		}
	}
	return builder.String()
}

func writeDoctorEvents(out io.Writer, report doctorReport) error {
	encoder := json.NewEncoder(out)
	for _, check := range report.Checks {
		event := struct {
			Kind        string `json:"kind"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			Requirement string `json:"requirement,omitempty"`
			Source      string `json:"source,omitempty"`
			Verdict     string `json:"verdict,omitempty"`
			Workspace   string `json:"workspace,omitempty"`
			PatchCount  int    `json:"patch_count,omitempty"`
			CheckCount  int    `json:"check_count,omitempty"`
			EventCount  int    `json:"event_count,omitempty"`
			JobID       string `json:"job_id,omitempty"`
			IssueKind   string `json:"issue_kind,omitempty"`
			Path        string `json:"path,omitempty"`
			Guidance    string `json:"guidance,omitempty"`
			Error       string `json:"error,omitempty"`
		}{
			Kind:        "doctor_check",
			Name:        check.Name,
			Status:      check.Status,
			Requirement: check.Requirement,
			Source:      check.Source,
			Verdict:     check.Verdict,
			Workspace:   check.Workspace,
			PatchCount:  check.PatchCount,
			CheckCount:  check.CheckCount,
			EventCount:  check.EventCount,
			JobID:       check.JobID,
			IssueKind:   check.IssueKind,
			Path:        check.Path,
			Guidance:    check.Guidance,
			Error:       check.Error,
		}
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}
