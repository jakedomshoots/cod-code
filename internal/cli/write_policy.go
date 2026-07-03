package cli

import (
	"fmt"
	"strings"
)

const (
	cliWritePolicyObserve       = "observe"
	cliWritePolicyDryRun        = "dry-run"
	cliWritePolicyPreview       = "preview"
	cliWritePolicyApprovedWrite = "approved-write"
	cliWritePolicyTrustedLocal  = "trusted-local"
)

func optionsWithWritePolicy(opts options) (options, error) {
	switch strings.TrimSpace(opts.writePolicy) {
	case "":
		if shouldDefaultPreview(opts) {
			opts.dryRun = true
		}
		return opts, nil
	case cliWritePolicyObserve, cliWritePolicyDryRun, cliWritePolicyPreview:
		opts.dryRun = true
		return opts, nil
	case cliWritePolicyApprovedWrite:
		if hasWriteIntent(opts) && !opts.dryRun && strings.TrimSpace(opts.approvedPreviewDigest) == "" {
			return options{}, fmt.Errorf("--write-policy approved-write requires --approve-preview for write actions")
		}
		return opts, nil
	case cliWritePolicyTrustedLocal:
		return opts, nil
	default:
		return options{}, fmt.Errorf("--write-policy must be observe, preview, dry-run, approved-write, or trusted-local")
	}
}

func hasWriteIntent(opts options) bool {
	return len(opts.patches) > 0 || opts.applyModelPatches
}

func shouldDefaultPreview(opts options) bool {
	return hasWriteIntent(opts) && strings.TrimSpace(opts.approvedPreviewDigest) == ""
}
