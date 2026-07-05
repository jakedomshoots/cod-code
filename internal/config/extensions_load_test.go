package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_LoadWorkspace_round_trips_valid_skills_and_mcp_servers(t *testing.T) {
	// Given: a workspace config containing a valid skills block and a valid
	// mcp_servers block.
	root := t.TempDir()
	content := `{
  "model_command": ["echo", "review"],
  "ceo_model_command": ["echo", "ceo"],
  "research_command": ["echo", "research"],
  "skills": {
    "alpha": {"path": "skills/alpha", "description": "Alpha skill", "allowed_actions": ["review"]},
    "bravo": {"path": "skills/bravo"}
  },
  "mcp_servers": {
    "stdio-srv": {"transport": "stdio", "command": ["node", "server.js"], "permissions": ["read"]},
    "http-srv": {"transport": "http", "url": "https://mcp.example.com/v1"}
  }
}`
	if err := os.WriteFile(filepath.Join(root, WorkspaceConfigName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	cfg, err := LoadWorkspace(context.Background(), root)

	// Then
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	alpha, ok := cfg.Skills["alpha"]
	if !ok {
		t.Fatalf("Skills[alpha] missing; got keys %v", extensionKeys(cfg.Skills))
	}
	if alpha.Path != "skills/alpha" || alpha.Description != "Alpha skill" || len(alpha.AllowedActions) != 1 || alpha.AllowedActions[0] != "review" {
		t.Fatalf("Skills[alpha] = %+v, want path=skills/alpha description=Alpha skill allowed_actions=[review]", alpha)
	}
	bravo, ok := cfg.Skills["bravo"]
	if !ok {
		t.Fatalf("Skills[bravo] missing; got keys %v", extensionKeys(cfg.Skills))
	}
	if bravo.Path != "skills/bravo" {
		t.Fatalf("Skills[bravo].Path = %q, want skills/bravo", bravo.Path)
	}
	stdioSrv, ok := cfg.MCPServers["stdio-srv"]
	if !ok {
		t.Fatalf("MCPServers[stdio-srv] missing; got keys %v", extensionKeys(cfg.MCPServers))
	}
	if stdioSrv.Transport != "stdio" || len(stdioSrv.Command) != 2 || stdioSrv.Command[0] != "node" {
		t.Fatalf("MCPServers[stdio-srv] = %+v, want stdio command [node server.js]", stdioSrv)
	}
	if len(stdioSrv.Permissions) != 1 || stdioSrv.Permissions[0] != "read" {
		t.Fatalf("MCPServers[stdio-srv].Permissions = %v, want [read]", stdioSrv.Permissions)
	}
	httpSrv, ok := cfg.MCPServers["http-srv"]
	if !ok {
		t.Fatalf("MCPServers[http-srv] missing; got keys %v", extensionKeys(cfg.MCPServers))
	}
	if httpSrv.Transport != "http" || httpSrv.URL != "https://mcp.example.com/v1" {
		t.Fatalf("MCPServers[http-srv] = %+v, want http url https://mcp.example.com/v1", httpSrv)
	}
}

func Test_LoadWorkspace_rejects_invalid_skill_path_from_disk(t *testing.T) {
	// Given: an on-disk config that declares a skill whose path is missing
	// (the "invalid empty skill path" case from the contract).
	root := t.TempDir()
	content := `{
  "model_command": ["echo", "review"],
  "skills": {"alpha": {"description": "no path"}}
}`
	if err := os.WriteFile(filepath.Join(root, WorkspaceConfigName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("LoadWorkspace error = %v, want ErrInvalidConfig", err)
	}
}

func Test_LoadWorkspace_rejects_invalid_mcp_transport_from_disk(t *testing.T) {
	// Given: an on-disk config that wires an MCP server with an unrecognized
	// transport — the "bad MCP transport" rejection contract.
	root := t.TempDir()
	content := `{
  "model_command": ["echo", "review"],
  "mcp_servers": {"alpha": {"transport": "grpc", "command": ["node", "server.js"]}}
}`
	if err := os.WriteFile(filepath.Join(root, WorkspaceConfigName), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	_, err := LoadWorkspace(context.Background(), root)

	// Then
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("LoadWorkspace error = %v, want ErrInvalidConfig for bad transport", err)
	}
}

func extensionKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
