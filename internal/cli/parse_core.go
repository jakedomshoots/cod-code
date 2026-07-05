package cli

func parseCoreFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--workspace":
		value, err := parseNextValue(args, index, "--workspace requires a path")
		if err != nil {
			return true, index, err
		}
		opts.workspaceDir = value
	case "--artifact-root":
		value, err := parseNextValue(args, index, "--artifact-root requires a path")
		if err != nil {
			return true, index, err
		}
		opts.artifactRoot = value
	case "--config-check":
		opts.showConfigCheck = true
		return true, index, nil
	case "--config-doctor":
		opts.showConfigCheck = true
		opts.showConfigDoctor = true
		return true, index, nil
	case "--config-explain":
		opts.showConfigCheck = true
		opts.showConfigExplain = true
		return true, index, nil
	case "--config-completions":
		opts.showConfigCompletions = true
		return true, index, nil
	case "--shell":
		value, err := parseNextValue(args, index, "--shell requires zsh, bash, or fish")
		if err != nil {
			return true, index, err
		}
		opts.completionShell = value
	case "--output-dir":
		value, err := parseNextValue(args, index, "--output-dir requires a directory")
		if err != nil {
			return true, index, err
		}
		opts.completionOutputDir = value
	case "--doctor":
		opts.showDoctor = true
		return true, index, nil
	default:
		return parseExtendedCoreFlag(args, index, opts)
	}
	return true, index + 1, nil
}

func parseExtendedCoreFlag(args []string, index int, opts *options) (bool, int, error) {
	if handled, next, err := parseProductionFlag(args, index, opts); handled || err != nil {
		return handled, next, err
	}
	if handled, next, err := parseToolSurfaceFlag(args, index, opts); handled || err != nil {
		return handled, next, err
	}
	if handled, next, err := parseRuntimeFlag(args, index, opts); handled || err != nil {
		return handled, next, err
	}
	return parseMiscCoreFlag(args, index, opts)
}
