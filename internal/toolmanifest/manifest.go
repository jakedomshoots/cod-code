package toolmanifest

import "ceoharness/internal/jobpacket"

type Manifest struct {
	SchemaVersion    int            `json:"schema_version"`
	Name             string         `json:"name"`
	Tools            []Tool         `json:"tools"`
	ExtensionSchemas map[string]any `json:"extension_schemas"`
}

type Tool struct {
	Name          string         `json:"name"`
	Action        string         `json:"action"`
	Description   string         `json:"description"`
	DefaultPolicy string         `json:"default_policy"`
	Permission    string         `json:"permission"`
	RequestSchema map[string]any `json:"request_schema"`
}

func Default() Manifest {
	return Manifest{
		SchemaVersion: 1,
		Name:          "ceo-packet-tools",
		Tools: []Tool{
			{
				Name:          "browser.read",
				Action:        string(jobpacket.ActionBrowserRead),
				Description:   "Fetch an HTTP(S) page for local web QA or approved public browsing and return bounded text plus a receipt digest.",
				DefaultPolicy: "allow-localhost",
				Permission:    "browser_read",
				RequestSchema: map[string]any{"type": "object", "required": []string{"action", "url"}, "properties": map[string]any{"action": map[string]any{"const": string(jobpacket.ActionBrowserRead)}, "url": map[string]any{"type": "string", "format": "uri"}, "max_bytes": map[string]any{"type": "integer", "minimum": 1}}},
			},
			{
				Name:          "computer.snapshot",
				Action:        string(jobpacket.ActionComputerSnapshot),
				Description:   "Run the configured desktop accessibility snapshot command for an approved app and return bounded output plus a receipt digest.",
				DefaultPolicy: "ask",
				Permission:    "computer_read",
				RequestSchema: map[string]any{"type": "object", "required": []string{"action", "app"}, "properties": map[string]any{"action": map[string]any{"const": string(jobpacket.ActionComputerSnapshot)}, "app": map[string]any{"type": "string"}, "max_bytes": map[string]any{"type": "integer", "minimum": 1}}},
			},
			{
				Name:          "tools.manifest",
				Action:        string(jobpacket.ActionToolManifest),
				Description:   "Return the built-in browser/computer tool capability manifest for skill and MCP pack routing.",
				DefaultPolicy: "allow",
				Permission:    "read_tools",
				RequestSchema: map[string]any{"type": "object", "required": []string{"action"}, "properties": map[string]any{"action": map[string]any{"const": string(jobpacket.ActionToolManifest)}}},
			},
		},
		ExtensionSchemas: map[string]any{
			"skills": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type":     "object",
					"required": []string{"path"},
					"properties": map[string]any{
						"path":            map[string]any{"type": "string"},
						"description":     map[string]any{"type": "string"},
						"allowed_actions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"mcp_servers": map[string]any{
				"type": "object",
				"additionalProperties": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"transport":   map[string]any{"enum": []string{"stdio", "http"}},
						"command":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"url":         map[string]any{"type": "string"},
						"permissions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
		},
	}
}
