package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type startReport struct {
	Status     string           `json:"status"`
	Workspace  string           `json:"workspace"`
	ConfigInit configInitReport `json:"config_init"`
	Doctor     doctorReport     `json:"doctor"`
	NextSteps  []string         `json:"next_steps,omitempty"`
}

func runStart(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildStartReport(ctx, opts)
	if err != nil {
		return err
	}
	if err := writeStartReport(out, report, opts.reportFormat); err != nil {
		return err
	}
	if report.Status != "pass" {
		return ErrVerdictFailed
	}
	return nil
}

func buildStartReport(ctx context.Context, opts options) (startReport, error) {
	opts.workspaceDir = opts.startDir
	opts.quickstartDir = opts.startDir
	opts.initConfig = true
	opts.initExampleAdapters = true
	opts = optionsWithQuickstartDefaults(opts)

	configReport, err := startConfigReport(ctx, opts)
	if err != nil {
		return startReport{}, err
	}
	configCheck, err := buildConfigCheckReport(ctx, opts)
	if err != nil {
		return startReport{}, err
	}
	doctor, _ := buildDoctorReport(ctx, opts)
	return startReport{
		Status:     doctor.Status,
		Workspace:  opts.workspaceDir,
		ConfigInit: configReport,
		Doctor:     doctor,
		NextSteps:  quickstartNextSteps(opts.workspaceDir, configCheck.ProviderSetupSteps),
	}, nil
}

func startConfigReport(ctx context.Context, opts options) (configInitReport, error) {
	path := workspaceConfigPath(opts.workspaceDir)
	if _, err := os.Stat(path); err == nil {
		return configInitReport{ConfigPath: path, Created: false, ExampleAdapters: opts.initExampleAdapters}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return configInitReport{}, fmt.Errorf("stat workspace config: %w", err)
	}
	return buildConfigInitReport(ctx, opts)
}

func writeStartReport(out io.Writer, report startReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return fmt.Errorf("write start report: %w", err)
		}
		return nil
	case reportFormatText:
		if _, err := io.WriteString(out, renderStartText(report)); err != nil {
			return fmt.Errorf("write text start report: %w", err)
		}
		return nil
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderStartText(report startReport) string {
	base := quickstartReport{
		Status:     report.Status,
		Workspace:  report.Workspace,
		ConfigInit: report.ConfigInit,
		Doctor:     report.Doctor,
		NextSteps:  report.NextSteps,
	}
	text := renderQuickstartText(base)
	return "Start" + text[len("Quickstart"):]
}
