package cli

import "fmt"

func parseProductionFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--production-status":
		opts.showProductionStatus = true
		return true, index, nil
	case "--production-actions":
		opts.showProductionActions = true
		return true, index, nil
	case "--action-id":
		value, err := parseNextValue(args, index, "--action-id requires an action id")
		if err != nil {
			return true, index, err
		}
		opts.productionActionID = value
	case "--action-kind":
		value, err := parseNextValue(args, index, "--action-kind requires an action kind")
		if err != nil {
			return true, index, err
		}
		opts.productionActionKind = value
	case "--action-provider":
		value, err := parseNextValue(args, index, "--action-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.productionActionProvider = value
	case "--action-state":
		value, err := parseNextValue(args, index, "--action-state requires ready, missing_env, empty_env, setup_blocked, or waiting")
		if err != nil {
			return true, index, err
		}
		if !validProductionActionState(value) {
			return true, index, fmt.Errorf("--action-state must be ready, missing_env, empty_env, setup_blocked, or waiting")
		}
		opts.productionActionState = value
	case "--env-ready-only":
		opts.productionActionsEnvReadyOnly = true
		return true, index, nil
	case "--ready-only":
		opts.productionActionsReadyOnly = true
		return true, index, nil
	case "--next":
		opts.productionActionsNextOnly = true
		return true, index, nil
	case "--commands-only":
		opts.productionActionsCommandsOnly = true
		return true, index, nil
	case "--production-finalize":
		opts.showProductionFinalize = true
		return true, index, nil
	case "--run-comparison":
		opts.productionFinalizeRunComparison = true
		return true, index, nil
	case "--dist":
		value, err := parseNextValue(args, index, "--dist requires a directory")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeDist = value
	case "--evidence-root":
		value, err := parseNextValue(args, index, "--evidence-root requires a directory")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeEvidenceRoot = value
	case "--provider-timeout-seconds":
		value, err := parseNonNegativeIntFlag(args, index, "--provider-timeout-seconds")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeProviderTimeoutSeconds = value
	case "--comparison-timeout-seconds":
		value, err := parseNonNegativeIntFlag(args, index, "--comparison-timeout-seconds")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeComparisonTimeoutSeconds = value
	case "--comparison-timeout-retries":
		value, err := parseNonNegativeIntFlag(args, index, "--comparison-timeout-retries")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeComparisonTimeoutRetries = value
	case "--comparison-result-retries":
		value, err := parseNonNegativeIntFlag(args, index, "--comparison-result-retries")
		if err != nil {
			return true, index, err
		}
		opts.productionFinalizeComparisonResultRetries = value
	case "--plan-only":
		opts.planOnly = true
		return true, index, nil
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}
