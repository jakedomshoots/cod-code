package cli

import (
	"context"
	"io"
)

type quickstartReport struct {
	Status     string           `json:"status"`
	Workspace  string           `json:"workspace"`
	ConfigInit configInitReport `json:"config_init"`
	Doctor     doctorReport     `json:"doctor"`
	NextSteps  []string         `json:"next_steps,omitempty"`
}

func runQuickstart(ctx context.Context, out io.Writer, opts options) error {
	opts.workspaceDir = opts.quickstartDir
	opts.initConfig = true
	opts.initExampleAdapters = true
	opts = optionsWithQuickstartDefaults(opts)

	configReport, err := buildConfigInitReport(ctx, opts)
	if err != nil {
		return err
	}
	configCheck, err := buildConfigCheckReport(ctx, opts)
	if err != nil {
		return err
	}
	doctor, doctorErr := buildDoctorReport(ctx, opts)
	report := quickstartReport{
		Status:     doctor.Status,
		Workspace:  opts.workspaceDir,
		ConfigInit: configReport,
		Doctor:     doctor,
		NextSteps:  quickstartNextSteps(opts.workspaceDir, configCheck.ProviderSetupSteps),
	}
	if err := writeQuickstartReport(out, report, opts.reportFormat); err != nil {
		return err
	}
	if report.Status != "pass" {
		if doctorErr != nil {
			return doctorErr
		}
		return ErrVerdictFailed
	}
	return nil
}
