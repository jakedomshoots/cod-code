package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type exampleAdapterSet struct {
	ModelCommand    []string
	CEOModelCommand []string
	ResearchCommand []string
}

func exampleAdapterCommands() (exampleAdapterSet, error) {
	model, err := exampleCommandForScript("command-model.sh")
	if err != nil {
		return exampleAdapterSet{}, err
	}
	ceo, err := exampleCommandForScript("ceo-model.sh")
	if err != nil {
		return exampleAdapterSet{}, err
	}
	research, err := exampleCommandForScript("research-command.sh")
	if err != nil {
		return exampleAdapterSet{}, err
	}
	return exampleAdapterSet{ModelCommand: model, CEOModelCommand: ceo, ResearchCommand: research}, nil
}

func exampleCommandForScript(name string) ([]string, error) {
	path, err := findRepoFile(filepath.Join("examples", name), "example adapter script "+name)
	if err != nil {
		return nil, err
	}
	return []string{"sh", path}, nil
}

func commandForExternalAdapter(name string) ([]string, error) {
	clean := strings.ToLower(strings.TrimSpace(name))
	switch clean {
	case "codex", "claude", "opencode", "aider", "goose":
		return exampleCommandForScript(filepath.Join("adapters", clean+".sh"))
	case "kimi":
		path, err := findRepoFile(filepath.Join("scripts", "kimi-model-command.sh"), "Kimi model command script")
		if err != nil {
			return nil, err
		}
		return []string{"sh", path}, nil
	default:
		return nil, fmt.Errorf("unknown --adapter %q", name)
	}
}

func findRepoFile(name string, description string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("find %s: %w", description, err)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		path := filepath.Join(dir, name)
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() {
			return filepath.Abs(path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found from %s", description, wd)
		}
	}
}
