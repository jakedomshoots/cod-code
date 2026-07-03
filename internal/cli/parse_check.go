package cli

func parseCheckFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--check":
		checkCommand, nextIndex, err := parseDelimitedCommand(args, index+1, "--check")
		if err != nil {
			return true, index, err
		}
		opts.checkCommand = checkCommand
		return true, nextIndex, nil
	case "--check-set":
		value, err := parseNextValue(args, index, "--check-set requires a name")
		if err != nil {
			return true, index, err
		}
		opts.checkSet = value
	case "--check-attempts":
		value, err := parsePositiveIntFlag(args, index, "--check-attempts")
		if err != nil {
			return true, index, err
		}
		opts.checkAttempts = value
	case "--check-backoff-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--check-backoff-ms")
		if err != nil {
			return true, index, err
		}
		opts.checkBackoffMS = value
	case "--tool-command-timeout-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--tool-command-timeout-ms")
		if err != nil {
			return true, index, err
		}
		opts.toolCommandTimeoutMS = value
	case "--check-fix-attempts":
		value, err := parsePositiveIntFlag(args, index, "--check-fix-attempts")
		if err != nil {
			return true, index, err
		}
		opts.checkFixAttempts = value
		opts.checkFixAttemptsSet = true
	case "--require-checks":
		opts.requireChecks = true
		return true, index, nil
	case "--ceo-revision-attempts":
		value, err := parsePositiveIntFlag(args, index, "--ceo-revision-attempts")
		if err != nil {
			return true, index, err
		}
		opts.ceoRevisionAttempts = value
		opts.ceoRevisionAttemptsSet = true
	case "--subagent-concurrency":
		value, err := parsePositiveIntFlag(args, index, "--subagent-concurrency")
		if err != nil {
			return true, index, err
		}
		opts.subagentConcurrency = value
	case "--max-tool-requests":
		value, err := parsePositiveIntFlag(args, index, "--max-tool-requests")
		if err != nil {
			return true, index, err
		}
		opts.maxToolRequests = value
	case "--subagent-attempts":
		value, err := parsePositiveIntFlag(args, index, "--subagent-attempts")
		if err != nil {
			return true, index, err
		}
		opts.subagentAttempts = value
	case "--subagent-backoff-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--subagent-backoff-ms")
		if err != nil {
			return true, index, err
		}
		opts.subagentBackoffMS = value
	case "--no-progress-stop":
		value, err := parsePositiveIntFlag(args, index, "--no-progress-stop")
		if err != nil {
			return true, index, err
		}
		opts.noProgressStop = value
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}
