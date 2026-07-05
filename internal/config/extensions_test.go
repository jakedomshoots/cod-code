package config

import (
	"errors"
	"strings"
	"testing"
)

// validBaseConfig is the minimal Config that passes Validate() ignoring the
// Skills/MCPServers block. Each test mutates Skills or MCPServers and asserts
// the resulting Validate() verdict; this keeps the focus on the extension
// contract rather than re-proving every other validator.
func validBaseConfig() Config {
	return Config{
		ModelCommand:    []string{"echo", "review"},
		CEOModelCommand: []string{"echo", "ceo"},
		ResearchCommand: []string{"echo", "research"},
	}
}

func Test_Validate_Extensions_accepts_valid_skill(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"echo-skill": {Path: "skills/echo", Description: "Echo skill", AllowedActions: []string{"review"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for a valid skill entry", err)
	}
}

func Test_Validate_Extensions_accepts_multiple_skill_entries(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"alpha": {Path: "skills/alpha"},
		"bravo": {Path: "skills/bravo", AllowedActions: []string{"review", "summarize"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for multiple valid skill entries", err)
	}
}

func Test_Validate_Extensions_rejects_skill_with_empty_path(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"echo-skill": {Path: "", Description: "missing path"},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("Validate() = nil, want ErrInvalidConfig for empty skill path")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig", err)
	}
	if !strings.Contains(err.Error(), "skills[echo-skill].path") {
		t.Fatalf("Validate() error = %q, want it to identify skills[echo-skill].path", err)
	}
}

func Test_Validate_Extensions_rejects_skill_with_whitespace_only_path(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"echo-skill": {Path: "   "},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for whitespace-only path", err)
	}
}

func Test_Validate_Extensions_rejects_skill_with_empty_allowed_action(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"echo-skill": {Path: "skills/echo", AllowedActions: []string{"review", "  "}},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for empty allowed_actions element", err)
	}
	if !strings.Contains(err.Error(), "allowed_actions[1]") {
		t.Fatalf("Validate() error = %q, want it to identify skills[echo-skill].allowed_actions[1]", err)
	}
}

func Test_Validate_Extensions_rejects_skill_with_empty_entry_name(t *testing.T) {
	cfg := validBaseConfig()
	cfg.Skills = map[string]SkillConfig{
		"": {Path: "skills/echo"},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for empty skill entry name", err)
	}
	if !strings.Contains(err.Error(), "skills key") {
		t.Fatalf("Validate() error = %q, want it to identify skills key", err)
	}
}

func Test_Validate_Extensions_accepts_mcp_server_with_stdio_and_command(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Transport: "stdio", Command: []string{"node", "server.js"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for stdio MCP server", err)
	}
}

func Test_Validate_Extensions_accepts_mcp_server_with_http_and_url(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Transport: "http", URL: "https://mcp.example.com/v1"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for http MCP server", err)
	}
}

func Test_Validate_Extensions_accepts_mcp_server_with_default_transport(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Command: []string{"node", "server.js"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() = %v, want nil for default-transport MCP server", err)
	}
}

func Test_Validate_Extensions_rejects_mcp_server_with_invalid_transport(t *testing.T) {
	cases := []struct {
		name      string
		transport string
	}{
		{name: "grpc", transport: "grpc"},
		{name: "websocket", transport: "websocket"},
		{name: "uppercase", transport: "STDIO"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validBaseConfig()
			cfg.MCPServers = map[string]MCPServerConfig{
				"alpha": {Transport: tc.transport, Command: []string{"node", "server.js"}},
			}
			err := cfg.Validate()
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("Validate() error = %v, want ErrInvalidConfig for transport %q", err, tc.transport)
			}
			if !strings.Contains(err.Error(), "mcp_servers[alpha].transport") {
				t.Fatalf("Validate() error = %q, want it to identify mcp_servers[alpha].transport", err)
			}
		})
	}
}

func Test_Validate_Extensions_rejects_mcp_server_with_neither_command_nor_url(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Transport: "stdio"},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig when no command or url is provided", err)
	}
	if !strings.Contains(err.Error(), "mcp_servers[alpha]") {
		t.Fatalf("Validate() error = %q, want it to identify mcp_servers[alpha]", err)
	}
}

func Test_Validate_Extensions_rejects_mcp_server_with_empty_command_arg(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Transport: "stdio", Command: []string{"node", ""}},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for empty command arg", err)
	}
	if !strings.Contains(err.Error(), "mcp_servers[alpha].command") {
		t.Fatalf("Validate() error = %q, want it to identify mcp_servers[alpha].command", err)
	}
}

func Test_Validate_Extensions_rejects_mcp_server_with_empty_permission(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"alpha": {Transport: "stdio", Command: []string{"node", "server.js"}, Permissions: []string{"read", "  "}},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for whitespace-only permissions element", err)
	}
	if !strings.Contains(err.Error(), "mcp_servers[alpha].permissions[1]") {
		t.Fatalf("Validate() error = %q, want it to identify mcp_servers[alpha].permissions[1]", err)
	}
}

func Test_Validate_Extensions_rejects_mcp_server_with_empty_entry_name(t *testing.T) {
	cfg := validBaseConfig()
	cfg.MCPServers = map[string]MCPServerConfig{
		"": {Transport: "stdio", Command: []string{"node", "server.js"}},
	}
	err := cfg.Validate()
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("Validate() error = %v, want ErrInvalidConfig for empty MCP entry name", err)
	}
	if !strings.Contains(err.Error(), "mcp_servers key") {
		t.Fatalf("Validate() error = %q, want it to identify mcp_servers key", err)
	}
}
