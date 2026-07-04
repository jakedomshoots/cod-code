package cli

import (
	"fmt"

	"ceoharness/internal/ceo"
)

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
	case "--env-ready-only":
		opts.productionActionsEnvReadyOnly = true
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
	case "--plan-only":
		opts.planOnly = true
		return true, index, nil
	case "--demo":
		opts.showDemo = true
		return true, index, nil
	case "--start":
		value, err := parseNextValue(args, index, "--start requires a workspace path")
		if err != nil {
			return true, index, err
		}
		opts.startDir = value
	case "--inbox":
		opts.showInbox = true
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
	case "--init-demo-repo":
		value, err := parseNextValue(args, index, "--init-demo-repo requires a path")
		if err != nil {
			return true, index, err
		}
		opts.initDemoRepoDir = value
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
	case "--init-config":
		opts.initConfig = true
		return true, index, nil
	case "--init-example-adapters":
		opts.initExampleAdapters = true
		return true, index, nil
	case "--adapter":
		value, err := parseNextValue(args, index, "--adapter requires a preset name")
		if err != nil {
			return true, index, err
		}
		opts.adapterName = value
	case "--quickstart":
		value, err := parseNextValue(args, index, "--quickstart requires a workspace path")
		if err != nil {
			return true, index, err
		}
		opts.quickstartDir = value
	case "--help", "-h":
		opts.showHelp = true
		return true, index, nil
	case "--help-advanced":
		opts.showAdvancedHelp = true
		return true, index, nil
	case "--version":
		opts.showVersion = true
		return true, index, nil
	case "--replace":
		if index+3 >= len(args) {
			return true, index, fmt.Errorf("--replace requires path, old text, and new text")
		}
		opts.patches = append(opts.patches, ceo.PatchRequest{
			Path: args[index+1],
			Old:  args[index+2],
			New:  args[index+3],
		})
		return true, index + 3, nil
	default:
		return false, index, nil
	}
	return true, index + 1, nil
}
