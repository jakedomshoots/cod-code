package cli

import (
	"context"

	"ceoharness/internal/ceo"
	"ceoharness/internal/subagent"
)

func runModelCommandDoctorCheck(ctx context.Context, opts options) (doctorCheck, bool) {
	selection, err := selectModelCommand(ctx, opts)
	if err != nil {
		return failedDoctorCheck("model_command", err), true
	}
	if len(selection.argv) == 0 {
		return doctorCheck{}, false
	}
	client, err := newModelCommandClient(selection.argv, selection.modelCommandTimeoutMS)
	if err != nil {
		return failedDoctorCheck("model_command", err), true
	}
	result, err := subagent.NewRunnerWithModel(client).Run(ctx, subagent.TaskPacket{
		Task:            "Doctor model command check",
		AgentName:       "doctor",
		Role:            "adapter health",
		ContextMode:     "lean",
		MaxContextBytes: 1024,
	})
	if err != nil {
		return failedDoctorCheck("model_command", err), true
	}
	check := doctorCheck{Name: "model_command", Status: result.Status, Source: selection.source, Verdict: result.Status}
	if result.Status != "pass" {
		check.Status = "fail"
		check.Error = result.Summary
	}
	return check, true
}

func runCEOModelCommandDoctorCheck(ctx context.Context, opts options) (doctorCheck, bool) {
	selection, err := selectCEOModelCommand(ctx, opts)
	if err != nil {
		return failedDoctorCheck("ceo_model_command", err), true
	}
	checkName := ceoDoctorCheckName(selection)
	if checkName == "" {
		return doctorCheck{}, false
	}
	client, err := ceoReviewerFromSelection(selection)
	if err != nil {
		return failedDoctorCheck(checkName, err), true
	}
	report, err := ceo.NewRuntimeWithCEOReviewerAndRoute(client, ceoReviewerRouteFromSelection(selection)).RunJob(ctx, ceo.JobRequest{
		Task: "Doctor CEO model command check",
	})
	if err != nil {
		return failedDoctorCheck(checkName, err), true
	}
	check := doctorCheck{Name: checkName, Status: report.Verdict, Source: selection.source, Verdict: report.Verdict}
	if report.CEODelegation == nil || report.CEOReview == nil {
		check.Status = "fail"
		check.Error = "CEO reviewer did not complete delegation and review"
	}
	return check, true
}

func ceoDoctorCheckName(selection commandSelection) string {
	if selection.providerName != "" {
		return "ceo_provider"
	}
	if len(selection.argv) > 0 {
		return "ceo_model_command"
	}
	return ""
}
