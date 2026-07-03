package config

import (
	"fmt"
	"strings"
)

type AutoCheckSet struct {
	CheckSet string   `json:"check_set"`
	Keywords []string `json:"keywords"`
}

func (cfg Config) AutoCheckCommandsForTask(task string) ([][]string, bool) {
	setName, ok := cfg.AutoCheckSetForTask(task)
	if !ok {
		return nil, false
	}
	return cfg.CheckCommandsForSet(setName)
}

func (cfg Config) AutoCheckSetForTask(task string) (string, bool) {
	cleanTask := strings.ToLower(strings.TrimSpace(task))
	if cleanTask == "" {
		return "", false
	}
	for _, rule := range cfg.AutoCheckSets {
		for _, keyword := range rule.Keywords {
			cleanKeyword := strings.ToLower(strings.TrimSpace(keyword))
			if cleanKeyword != "" && strings.Contains(cleanTask, cleanKeyword) {
				return strings.TrimSpace(rule.CheckSet), true
			}
		}
	}
	return "", false
}

func (cfg Config) validateAutoCheckSets() error {
	for index, rule := range cfg.AutoCheckSets {
		setName := strings.TrimSpace(rule.CheckSet)
		if setName == "" || len(rule.Keywords) == 0 {
			return fmt.Errorf("auto_check_sets[%d]: %w", index, ErrInvalidConfig)
		}
		if _, ok := cfg.CheckSets[setName]; !ok {
			return fmt.Errorf("auto_check_sets[%d].check_set: %w", index, ErrInvalidConfig)
		}
		for keywordIndex, keyword := range rule.Keywords {
			if strings.TrimSpace(keyword) == "" {
				return fmt.Errorf("auto_check_sets[%d].keywords[%d]: %w", index, keywordIndex, ErrInvalidConfig)
			}
		}
	}
	return nil
}
