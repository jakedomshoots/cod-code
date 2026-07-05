package cli

func parseRuntimeFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
	case "--tui":
		opts.showTUI = true
		return true, index, nil
	case "--snapshot":
		opts.showTUI = true
		opts.tuiSnapshot = true
		return true, index, nil
	case "--dry-run":
		opts.dryRun = true
		return true, index, nil
	case "--write-policy":
		value, err := parseNextValue(args, index, "--write-policy requires observe, preview, dry-run, approved-write, or trusted-local")
		if err != nil {
			return true, index, err
		}
		opts.writePolicy = value
	case "--rollback-report":
		value, err := parseNextValue(args, index, "--rollback-report requires a report path")
		if err != nil {
			return true, index, err
		}
		opts.rollbackReportPath = value
	case "--approve-preview":
		value, err := parseNextValue(args, index, "--approve-preview requires a digest")
		if err != nil {
			return true, index, err
		}
		opts.approvedPreviewDigest = value
	case "--job-timeout-ms":
		value, err := parseNonNegativeIntFlag(args, index, "--job-timeout-ms")
		if err != nil {
			return true, index, err
		}
		opts.jobTimeoutMS = value
	case "--max-ceo-iterations":
		value, err := parsePositiveIntFlag(args, index, "--max-ceo-iterations")
		if err != nil {
			return true, index, err
		}
		opts.maxCEOIterations = value
		opts.maxCEOIterationsSet = true
	case "--max-subagents":
		value, err := parsePositiveIntFlag(args, index, "--max-subagents")
		if err != nil {
			return true, index, err
		}
		opts.maxSubagents = value
	case "--interactive":
		opts.interactive = true
		return true, index, nil
	case "--apply-model-patches":
		opts.applyModelPatches = true
		return true, index, nil
	case "--preview-model-patches":
		opts.previewModelPatches = true
		return true, index, nil
	case "--max-model-patches":
		value, err := parsePositiveIntFlag(args, index, "--max-model-patches")
		if err != nil {
			return true, index, err
		}
		opts.maxModelPatches = value
	case "--max-context-bytes":
		value, err := parsePositiveIntFlag(args, index, "--max-context-bytes")
		if err != nil {
			return true, index, err
		}
		opts.maxContextBytes = value
	case "--max-subagent-output-bytes":
		value, err := parsePositiveIntFlag(args, index, "--max-subagent-output-bytes")
		if err != nil {
			return true, index, err
		}
		opts.maxSubagentOutputBytes = value
	case "--workspace-brief-exclude":
		value, err := parseNextValue(args, index, "--workspace-brief-exclude requires a path or glob")
		if err != nil {
			return true, index, err
		}
		opts.workspaceBriefExcludes = append(opts.workspaceBriefExcludes, value)
	case "--workspace-brief-max-files":
		value, err := parsePositiveIntFlag(args, index, "--workspace-brief-max-files")
		if err != nil {
			return true, index, err
		}
		opts.workspaceBriefMaxFiles = value
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}
