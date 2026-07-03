package cli

import "fmt"

const repairPresetStandard = "standard"

func optionsWithRepairPreset(opts options) (options, error) {
	if opts.repairPreset == "" {
		return opts, nil
	}
	if opts.repairPreset != repairPresetStandard {
		return options{}, fmt.Errorf("--repair-preset %q is not supported; use standard", opts.repairPreset)
	}
	if !opts.checkFixAttemptsSet {
		opts.checkFixAttempts = 1
	}
	if !opts.ceoRevisionAttemptsSet {
		opts.ceoRevisionAttempts = 1
	}
	if !opts.maxCEOIterationsSet {
		opts.maxCEOIterations = 3
	}
	return opts, nil
}
