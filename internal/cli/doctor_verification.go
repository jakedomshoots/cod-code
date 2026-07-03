package cli

import "context"

func runVerificationPolicyDoctorCheck(ctx context.Context, opts options) doctorCheck {
	opts, err := optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return failedDoctorCheck("verification_policy", err)
	}
	if err := requireVerificationChecks(opts); err != nil {
		return failedDoctorCheck("verification_policy", err)
	}
	status := "pass"
	if !opts.requireChecks {
		status = "skipped"
	}
	return doctorCheck{Name: "verification_policy", Status: status}
}
