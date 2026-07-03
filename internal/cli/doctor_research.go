package cli

import (
	"context"
	"fmt"
	"strings"

	"ceoharness/internal/researchrunner"
)

func runResearchCommandDoctorCheck(ctx context.Context, opts options) (doctorCheck, bool) {
	selection, err := selectResearchCommand(ctx, opts)
	if err != nil {
		return failedDoctorCheck("research_command", err), true
	}
	if len(selection.argv) == 0 {
		return doctorCheck{}, false
	}
	result, err := researchrunner.NewRunner().Run(ctx, researchrunner.Command{
		Argv:      selection.argv,
		Query:     "Doctor research command check",
		TimeoutMS: selection.timeoutMS,
	})
	if err != nil {
		check := failedDoctorCheck("research_command", err)
		check.Source = selection.source
		return check, true
	}
	check := doctorCheck{Name: "research_command", Status: result.Status, Source: selection.source, Verdict: result.Status}
	if result.Status != "pass" {
		check.Status = "fail"
		check.Error = researchDoctorFailure(result)
	}
	return check, true
}

func researchDoctorFailure(result researchrunner.Result) string {
	if text := strings.TrimSpace(result.Error); text != "" {
		return text
	}
	if text := strings.TrimSpace(result.Output); text != "" {
		return text
	}
	return fmt.Sprintf("research command failed with exit code %d", result.ExitCode)
}
