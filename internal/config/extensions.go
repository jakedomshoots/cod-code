package config

import (
	"fmt"
	"strings"
)

type SkillConfig struct {
	Path           string   `json:"path,omitempty"`
	Description    string   `json:"description,omitempty"`
	AllowedActions []string `json:"allowed_actions,omitempty"`
}

type MCPServerConfig struct {
	Transport   string   `json:"transport,omitempty"`
	Command     []string `json:"command,omitempty"`
	URL         string   `json:"url,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func validateExtensions(cfg Config) error {
	for name, skill := range cfg.Skills {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("skills key: %w", ErrInvalidConfig)
		}
		if strings.TrimSpace(skill.Path) == "" {
			return fmt.Errorf("skills[%s].path: %w", name, ErrInvalidConfig)
		}
		for index, action := range skill.AllowedActions {
			if strings.TrimSpace(action) == "" {
				return fmt.Errorf("skills[%s].allowed_actions[%d]: %w", name, index, ErrInvalidConfig)
			}
		}
	}
	for name, server := range cfg.MCPServers {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("mcp_servers key: %w", ErrInvalidConfig)
		}
		transport := strings.TrimSpace(server.Transport)
		switch transport {
		case "", "stdio", "http":
		default:
			return fmt.Errorf("mcp_servers[%s].transport: %w", name, ErrInvalidConfig)
		}
		if len(server.Command) == 0 && strings.TrimSpace(server.URL) == "" {
			return fmt.Errorf("mcp_servers[%s]: command or url is required: %w", name, ErrInvalidConfig)
		}
		if len(server.Command) > 0 {
			if err := validateCommand(fmt.Sprintf("mcp_servers[%s].command", name), server.Command, false); err != nil {
				return err
			}
		}
		for index, permission := range server.Permissions {
			if strings.TrimSpace(permission) == "" {
				return fmt.Errorf("mcp_servers[%s].permissions[%d]: %w", name, index, ErrInvalidConfig)
			}
		}
	}
	return nil
}
