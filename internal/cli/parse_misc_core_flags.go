package cli

import (
	"fmt"

	"ceoharness/internal/ceo"
)

func parseMiscCoreFlag(args []string, index int, opts *options) (bool, int, error) {
	switch args[index] {
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
	case "--init-demo-repo":
		value, err := parseNextValue(args, index, "--init-demo-repo requires a path")
		if err != nil {
			return true, index, err
		}
		opts.initDemoRepoDir = value
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
