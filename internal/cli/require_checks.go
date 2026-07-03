package cli

import "fmt"

func requireVerificationChecks(opts options) error {
	if !opts.requireChecks {
		return nil
	}
	if len(planCheckCommands(opts)) > 0 {
		return nil
	}
	return fmt.Errorf("--require-checks requires at least one verification command")
}
