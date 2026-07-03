package config

import (
	"fmt"
	"strings"
)

func (cfg Config) AgentCommands() map[string][]string {
	return cfg.AgentCommandsFor(cfg.AgentProviders)
}

func (cfg Config) AgentCommandsFor(agentProviders map[string]string) map[string][]string {
	commands := map[string][]string{}
	for agentName, providerName := range agentProviders {
		command := cfg.Providers[providerName].ModelCommand
		if len(command) > 0 {
			commands[agentName] = append([]string(nil), command...)
		}
	}
	for agentName, command := range cfg.AgentModelCommands {
		commands[agentName] = append([]string(nil), command...)
	}
	return commands
}

func (cfg Config) CheckCommandList() [][]string {
	if len(cfg.CheckCommands) > 0 {
		return cloneCommands(cfg.CheckCommands)
	}
	if len(cfg.CheckCommand) == 0 {
		return nil
	}
	return [][]string{append([]string(nil), cfg.CheckCommand...)}
}

func (cfg Config) CheckCommandsForSet(name string) ([][]string, bool) {
	commands, ok := cfg.CheckSets[strings.TrimSpace(name)]
	if !ok {
		return nil, false
	}
	return cloneCommands(commands), true
}

func cloneCommands(commands [][]string) [][]string {
	copied := make([][]string, 0, len(commands))
	for _, command := range commands {
		copied = append(copied, append([]string(nil), command...))
	}
	return copied
}

func validateCommand(name string, command []string, allowEmpty bool) error {
	if len(command) == 0 {
		if allowEmpty {
			return nil
		}
		return fmt.Errorf("%s: %w", name, ErrInvalidConfig)
	}
	for index, arg := range command {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf("%s[%d]: %w", name, index, ErrInvalidConfig)
		}
	}
	return nil
}
