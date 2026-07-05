package cli

func parseToolSurfaceFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--oauth":
		value, err := parseNextValue(args, index, "--oauth requires list, doctor, or init")
		if err != nil {
			return true, index, err
		}
		opts.oauthCommand = value
	case "--oauth-provider":
		value, err := parseNextValue(args, index, "--oauth-provider requires a provider name")
		if err != nil {
			return true, index, err
		}
		opts.oauthProvider = value
	case "--browser":
		value, err := parseNextValue(args, index, "--browser requires doctor, manifest, or read")
		if err != nil {
			return true, index, err
		}
		opts.browserCommand = value
	case "--browser-url":
		value, err := parseNextValue(args, index, "--browser-url requires a URL")
		if err != nil {
			return true, index, err
		}
		opts.browserURL = value
	case "--browser-policy":
		value, err := parseNextValue(args, index, "--browser-policy requires deny, ask, allow-localhost, or allow")
		if err != nil {
			return true, index, err
		}
		opts.browserPolicy = value
	case "--browser-command":
		command, nextIndex, err := parseDelimitedCommand(args, index+1, "--browser-command")
		if err != nil {
			return true, index, err
		}
		opts.browserBackendCommand = command
		return true, nextIndex, nil
	case "--computer":
		value, err := parseNextValue(args, index, "--computer requires doctor, manifest, or snapshot")
		if err != nil {
			return true, index, err
		}
		opts.computerCommand = value
	case "--computer-app":
		value, err := parseNextValue(args, index, "--computer-app requires an app name")
		if err != nil {
			return true, index, err
		}
		opts.computerApp = value
	case "--computer-policy":
		value, err := parseNextValue(args, index, "--computer-policy requires deny, ask, or allow")
		if err != nil {
			return true, index, err
		}
		opts.computerPolicy = value
	case "--computer-command":
		command, nextIndex, err := parseDelimitedCommand(args, index+1, "--computer-command")
		if err != nil {
			return true, index, err
		}
		opts.computerBackendCommand = command
		return true, nextIndex, nil
	case "--tools-manifest":
		opts.showToolManifest = true
		return true, index, nil
	case "--provider-wizard":
		value, err := parseNextValue(args, index, "--provider-wizard requires a preset name")
		if err != nil {
			return true, index, err
		}
		opts.providerWizardPreset = value
	case "--repair-preset":
		value, err := parseNextValue(args, index, "--repair-preset requires standard")
		if err != nil {
			return true, index, err
		}
		opts.repairPreset = value
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}
