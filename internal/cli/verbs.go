package cli

import (
	"fmt"
	"strings"
)

func normalizeVerbArgs(args []string) ([]string, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return args, nil
	}
	verb := args[0]
	rest := args[1:]
	if verbHelpRequested(rest) {
		if verb == "config" || isKnownVerb(verb) {
			return []string{"--help"}, nil
		}
		return nil, unknownVerbError(verb)
	}
	switch verb {
	case "start":
		return append([]string{"--start"}, rest...), nil
	case "run":
		return append([]string{}, rest...), nil
	case "inbox":
		return append([]string{"--inbox"}, rest...), nil
	case "status":
		return append([]string{"--history", "--summary-only"}, rest...), nil
	case "production-status":
		return append([]string{"--production-status"}, rest...), nil
	case "production-actions":
		return append([]string{"--production-actions"}, rest...), nil
	case "production-finalize":
		return append([]string{"--production-finalize"}, rest...), nil
	case "review":
		return append([]string{"--review-queue", "--review-details"}, rest...), nil
	case "context":
		return normalizeContextVerb(rest)
	case "oauth":
		return normalizeOAuthVerb(rest)
	case "browser":
		return normalizeBrowserVerb(rest)
	case "computer":
		return normalizeComputerVerb(rest)
	case "tools":
		return normalizeToolsVerb(rest)
	case "config":
		return normalizeConfigVerb(rest)
	case "doctor":
		return append([]string{"--doctor"}, rest...), nil
	case "resume":
		return normalizeIDVerb(rest, "--resume", "resume requires a job id; use latest or a saved job id")
	case "retry":
		return normalizeIDVerb(rest, "--rerun", "retry requires a job id; use latest or a saved job id")
	case "rollback":
		return normalizeIDVerb(rest, "--rollback-report", "rollback requires a saved report path")
	case "explain-failure":
		return normalizeIDVerb(rest, "--explain-failure", "explain-failure requires a job id; use latest or a saved job id")
	case "tui":
		return append([]string{"--tui"}, rest...), nil
	case "eval":
		return args, nil
	default:
		return args, nil
	}
}

func normalizeIDVerb(args []string, flag string, missingMessage string) ([]string, error) {
	idIndex := firstValueIndex(args)
	if idIndex < 0 {
		return nil, fmt.Errorf("%s", missingMessage)
	}
	normalized := make([]string, 0, len(args)+1)
	normalized = append(normalized, args[:idIndex]...)
	normalized = append(normalized, args[idIndex+1:]...)
	normalized = append(normalized, flag, args[idIndex])
	return normalized, nil
}

func firstValueIndex(args []string) int {
	for index := 0; index < len(args); index++ {
		if strings.HasPrefix(args[index], "-") {
			if verbFlagConsumesValue(args[index]) {
				index++
			}
			continue
		}
		return index
	}
	return -1
}

func verbFlagConsumesValue(flag string) bool {
	switch flag {
	case "--workspace", "--format", "--answer", "--browser-policy", "--computer-policy", "--browser-url", "--computer-app":
		return true
	default:
		return false
	}
}

func normalizeConfigVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("config requires check, doctor, explain, completions, or init; run ceo-packet --help")
	}
	subcommand := args[0]
	rest := args[1:]
	switch subcommand {
	case "check":
		return append([]string{"--config-check"}, rest...), nil
	case "doctor":
		return append([]string{"--config-doctor"}, rest...), nil
	case "explain":
		return append([]string{"--config-explain"}, rest...), nil
	case "completions":
		return append([]string{"--config-completions"}, rest...), nil
	case "init":
		return append([]string{"--init-config"}, rest...), nil
	default:
		return nil, fmt.Errorf("unknown config command %q; run ceo-packet --help", subcommand)
	}
}

func normalizeContextVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("context requires a job id; use latest or a saved job id")
	}
	jobIndex := contextVerbJobIndex(args)
	if jobIndex < 0 {
		return nil, fmt.Errorf("context requires a job id; use latest or a saved job id")
	}
	normalized := make([]string, 0, len(args)+1)
	normalized = append(normalized, args[:jobIndex]...)
	normalized = append(normalized, args[jobIndex+1:]...)
	normalized = append(normalized, "--context-trace", args[jobIndex])
	return normalized, nil
}

func contextVerbJobIndex(args []string) int {
	return firstValueIndex(args)
}

func isKnownVerb(verb string) bool {
	switch verb {
	case "start", "run", "gauntlet", "doctor", "inbox", "status", "production-status", "production-actions", "production-finalize", "resume", "retry", "rollback", "explain-failure", "review", "context", "oauth", "browser", "computer", "tools", "tui", "eval":
		return true
	default:
		return false
	}
}

func normalizeBrowserVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("browser requires doctor, manifest, or read; run ceo-packet --help")
	}
	subcommand := args[0]
	rest := args[1:]
	switch subcommand {
	case "doctor", "manifest":
		return append([]string{"--browser", subcommand}, rest...), nil
	case "read":
		urlIndex := firstValueIndex(rest)
		if urlIndex < 0 {
			return nil, fmt.Errorf("browser read requires a URL")
		}
		normalized := []string{"--browser", "read"}
		normalized = append(normalized, rest[:urlIndex]...)
		normalized = append(normalized, "--browser-url", rest[urlIndex])
		normalized = append(normalized, rest[urlIndex+1:]...)
		return normalized, nil
	default:
		return nil, fmt.Errorf("unknown browser command %q; run ceo-packet --help", subcommand)
	}
}

func normalizeComputerVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("computer requires doctor, manifest, or snapshot; run ceo-packet --help")
	}
	subcommand := args[0]
	rest := args[1:]
	switch subcommand {
	case "doctor", "manifest":
		return append([]string{"--computer", subcommand}, rest...), nil
	case "snapshot":
		appIndex := firstValueIndex(rest)
		if appIndex < 0 {
			return nil, fmt.Errorf("computer snapshot requires an app name")
		}
		normalized := []string{"--computer", "snapshot"}
		normalized = append(normalized, rest[:appIndex]...)
		normalized = append(normalized, "--computer-app", rest[appIndex])
		normalized = append(normalized, rest[appIndex+1:]...)
		return normalized, nil
	default:
		return nil, fmt.Errorf("unknown computer command %q; run ceo-packet --help", subcommand)
	}
}

func normalizeToolsVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("tools requires manifest; run ceo-packet --help")
	}
	switch args[0] {
	case "manifest":
		return append([]string{"--tools-manifest"}, args[1:]...), nil
	default:
		return nil, fmt.Errorf("unknown tools command %q; run ceo-packet --help", args[0])
	}
}

func normalizeOAuthVerb(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("oauth requires list, doctor, or init; run ceo-packet --help")
	}
	subcommand := args[0]
	rest := args[1:]
	switch subcommand {
	case "list":
		return append([]string{"--oauth", "list"}, rest...), nil
	case "doctor":
		normalized := []string{"--oauth", "doctor"}
		providerIndex := firstValueIndex(rest)
		if providerIndex >= 0 {
			normalized = append(normalized, rest[:providerIndex]...)
			normalized = append(normalized, "--oauth-provider", rest[providerIndex])
			normalized = append(normalized, rest[providerIndex+1:]...)
			return normalized, nil
		}
		return append(normalized, rest...), nil
	case "init":
		providerIndex := firstValueIndex(rest)
		if providerIndex < 0 {
			return nil, fmt.Errorf("oauth init requires a provider name")
		}
		normalized := []string{"--oauth", "init"}
		normalized = append(normalized, rest[:providerIndex]...)
		normalized = append(normalized, "--oauth-provider", rest[providerIndex])
		normalized = append(normalized, rest[providerIndex+1:]...)
		return normalized, nil
	default:
		return nil, fmt.Errorf("unknown oauth command %q; run ceo-packet --help", subcommand)
	}
}

func verbHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func unknownVerbError(verb string) error {
	return fmt.Errorf("unknown command %q; run ceo-packet --help", verb)
}
