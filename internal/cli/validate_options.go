package cli

import (
	"errors"
	"strings"
)

func validateOptions(opts options) error {
	if opts.initExampleAdapters && !opts.initConfig {
		return errors.New("--init-example-adapters requires --init-config")
	}
	if opts.quickstartDir != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--quickstart cannot be combined with task text")
	}
	if opts.startDir != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--start cannot be combined with task text")
	}
	if opts.initDemoRepoDir != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--init-demo-repo cannot be combined with task text")
	}
	if opts.providerWizardPreset != "" && strings.TrimSpace(opts.workspaceDir) == "" {
		return errors.New("--provider-wizard requires --workspace")
	}
	if opts.providerWizardPreset != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--provider-wizard cannot be combined with task text")
	}
	if opts.oauthCommand != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("oauth cannot be combined with task text")
	}
	if opts.oauthCommand == "init" && strings.TrimSpace(opts.workspaceDir) == "" {
		return errors.New("oauth init requires --workspace")
	}
	if opts.oauthCommand == "init" && strings.TrimSpace(opts.oauthProvider) == "" {
		return errors.New("oauth init requires a provider name")
	}
	if opts.browserCommand != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("browser cannot be combined with task text")
	}
	if opts.browserCommand == "read" && strings.TrimSpace(opts.browserURL) == "" {
		return errors.New("browser read requires a URL")
	}
	if opts.computerCommand != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("computer cannot be combined with task text")
	}
	if opts.computerCommand == "snapshot" && strings.TrimSpace(opts.computerApp) == "" {
		return errors.New("computer snapshot requires an app name")
	}
	if opts.showToolManifest && strings.TrimSpace(opts.task) != "" {
		return errors.New("tools manifest cannot be combined with task text")
	}
	if opts.adapterName != "" && !opts.initConfig {
		return errors.New("--adapter requires --init-config")
	}
	if opts.reviewDetails && !opts.showReviewQueue {
		return errors.New("--review-details requires --review-queue")
	}
	if strings.TrimSpace(opts.humanVerdict) != "" && strings.TrimSpace(opts.judgeJobID) == "" {
		return errors.New("--human-verdict requires --judge-job")
	}
	if strings.TrimSpace(opts.judgmentNote) != "" && strings.TrimSpace(opts.judgeJobID) == "" {
		return errors.New("--judgment-note requires --judge-job")
	}
	if strings.TrimSpace(opts.judgeJobID) != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--judge-job cannot be combined with task text")
	}
	if strings.TrimSpace(opts.rollbackReportPath) != "" && strings.TrimSpace(opts.workspaceDir) == "" {
		return errors.New("--rollback-report requires --workspace")
	}
	if strings.TrimSpace(opts.rollbackReportPath) != "" && strings.TrimSpace(opts.task) != "" {
		return errors.New("--rollback-report cannot be combined with task text")
	}
	if opts.applyModelPatches && opts.previewModelPatches {
		return errors.New("choose either model patch preview or model patch application")
	}
	return nil
}
