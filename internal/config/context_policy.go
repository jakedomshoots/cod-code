package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (cfg Config) HasContextPolicy() bool {
	return cfg.MaxContextBytes > 0 ||
		cfg.WorkspaceBriefMaxFiles > 0 ||
		len(cfg.WorkspaceBriefExcludes) > 0
}

func validateContextPolicy(cfg Config) error {
	if cfg.MaxContextBytes < 0 {
		return fmt.Errorf("max_context_bytes: %w", ErrInvalidConfig)
	}
	if cfg.WorkspaceBriefMaxFiles < 0 {
		return fmt.Errorf("workspace_brief_max_files: %w", ErrInvalidConfig)
	}
	for index, pattern := range cfg.WorkspaceBriefExcludes {
		if err := validateWorkspaceBriefExclude(pattern); err != nil {
			return fmt.Errorf("workspace_brief_excludes[%d]: %w", index, err)
		}
	}
	return nil
}

func validateWorkspaceBriefExclude(pattern string) error {
	cleanPattern := filepath.Clean(strings.TrimSpace(pattern))
	if cleanPattern == "." || filepath.IsAbs(cleanPattern) || cleanPattern == ".." || strings.HasPrefix(cleanPattern, ".."+string(filepath.Separator)) {
		return ErrInvalidConfig
	}
	return nil
}
