package cli

import (
	"fmt"
	"strings"
)

func parseArgs(args []string) (options, error) {
	args, err := normalizeVerbArgs(args)
	if err != nil {
		return options{}, err
	}
	opts := options{}
	taskArgs := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		handled, nextIndex, err := parseHTTPInitFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		handled, nextIndex, err = parseReportFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		handled, nextIndex, err = parseModelFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		handled, nextIndex, err = parseCoreFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		handled, nextIndex, err = parseCheckFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		handled, nextIndex, err = parseProviderPolicyFlag(args, index, &opts)
		if err != nil {
			return options{}, err
		}
		if handled {
			index = nextIndex
			continue
		}
		if args[index] == "--" {
			taskArgs = append(taskArgs, args[index+1:]...)
			break
		}
		if strings.HasPrefix(args[index], "-") {
			return options{}, fmt.Errorf("unknown flag %s", args[index])
		}
		taskArgs = append(taskArgs, args[index])
	}
	opts.task = strings.Join(taskArgs, " ")
	opts, err = optionsWithRepairPreset(opts)
	if err != nil {
		return options{}, err
	}
	opts.finishHTTPInit()
	return opts, nil
}
