package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadWorkspace(ctx context.Context, workspaceDir string) (Config, error) {
	if err := ctx.Err(); err != nil {
		return Config{}, err
	}
	root := strings.TrimSpace(workspaceDir)
	if root == "" {
		return Config{}, nil
	}
	content, err := os.ReadFile(filepath.Join(root, WorkspaceConfigName))
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read workspace config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse workspace config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate workspace config: %w", err)
	}
	return cfg, nil
}

func CreateWorkspace(ctx context.Context, workspaceDir string, cfg Config) (path string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	root := strings.TrimSpace(workspaceDir)
	if root == "" {
		return "", fmt.Errorf("workspace dir is required")
	}
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("validate workspace config: %w", err)
	}
	path = filepath.Join(root, WorkspaceConfigName)
	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal workspace config: %w", err)
	}
	content = append(content, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("%s: %w", path, ErrConfigExists)
	}
	if err != nil {
		return "", fmt.Errorf("create workspace config: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close workspace config: %w", closeErr)
		}
	}()
	if _, err := file.Write(content); err != nil {
		return "", fmt.Errorf("write workspace config: %w", err)
	}
	return path, nil
}
