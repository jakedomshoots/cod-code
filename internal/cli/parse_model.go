package cli

func parseModelFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--model-command":
		modelCommand, nextIndex, err := parseDelimitedCommand(args, index+1, "--model-command")
		if err != nil {
			return false, index, err
		}
		opts.modelCommand = modelCommand
		return true, nextIndex, nil
	case "--ceo-model-command":
		modelCommand, nextIndex, err := parseDelimitedCommand(args, index+1, "--ceo-model-command")
		if err != nil {
			return false, index, err
		}
		opts.ceoModelCommand = modelCommand
		return true, nextIndex, nil
	case "--research-command":
		researchCommand, nextIndex, err := parseDelimitedCommand(args, index+1, "--research-command")
		if err != nil {
			return false, index, err
		}
		opts.researchCommand = researchCommand
		return true, nextIndex, nil
	case "--model-command-timeout-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--model-command-timeout-ms")
		if err != nil {
			return true, index, err
		}
		opts.modelCommandTimeoutMS = value
		return true, index + 1, nil
	default:
		return false, index, nil
	}
}
