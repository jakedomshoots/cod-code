package toolmanifest

import (
	"sort"
	"testing"
)

func Test_Default_ExtensionSchemas_exposes_skills_and_mcp_servers(t *testing.T) {
	// Given
	m := Default()

	// Then
	if m.ExtensionSchemas == nil {
		t.Fatalf("Default().ExtensionSchemas = nil, want non-nil map exposing skills and mcp_servers")
	}
	if _, ok := m.ExtensionSchemas["skills"]; !ok {
		t.Fatalf("Default().ExtensionSchemas missing key %q; keys = %v", "skills", keysAny(m.ExtensionSchemas))
	}
	if _, ok := m.ExtensionSchemas["mcp_servers"]; !ok {
		t.Fatalf("Default().ExtensionSchemas missing key %q; keys = %v", "mcp_servers", keysAny(m.ExtensionSchemas))
	}
}

func Test_Default_ExtensionSchemas_requires_path_on_skills(t *testing.T) {
	// The skills schema is what callers consume to validate a skill entry. The
	// shape must require "path" — otherwise a schema-only consumer would admit
	// a skill with no path, diverging from the config validator which rejects
	// that case.
	m := Default()
	raw, ok := m.ExtensionSchemas["skills"]
	if !ok {
		t.Fatalf("Default().ExtensionSchemas missing key %q; keys = %v", "skills", keysAny(m.ExtensionSchemas))
	}
	skillsSchema, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("ExtensionSchemas[skills] type = %T, want map[string]any", raw)
	}
	additional, ok := skillsSchema["additionalProperties"].(map[string]any)
	if !ok {
		t.Fatalf("skills schema missing additionalProperties object; got %v", skillsSchema)
	}
	required, ok := additional["required"].([]string)
	if !ok {
		t.Fatalf("skills.additionalProperties.required type = %T, want []string", additional["required"])
	}
	if !containsString(required, "path") {
		t.Fatalf("skills schema required = %v, want to contain \"path\"", required)
	}
}

func Test_Default_ExtensionSchemas_lists_allowed_transports_for_mcp_servers(t *testing.T) {
	// The mcp_servers schema is what callers consume to validate an MCP entry.
	// It must constrain "transport" to exactly "stdio" and "http" — matching
	// the set the config validator accepts. Any additional transport the
	// schema permits but the validator rejects would create a false-positive
	// admission; any missing transport would leave callers unable to express
	// the allowed set.
	m := Default()
	raw, ok := m.ExtensionSchemas["mcp_servers"]
	if !ok {
		t.Fatalf("Default().ExtensionSchemas missing key %q; keys = %v", "mcp_servers", keysAny(m.ExtensionSchemas))
	}
	mcpSchema, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("ExtensionSchemas[mcp_servers] type = %T, want map[string]any", raw)
	}
	additional, ok := mcpSchema["additionalProperties"].(map[string]any)
	if !ok {
		t.Fatalf("mcp_servers schema missing additionalProperties object; got %v", mcpSchema)
	}
	properties, ok := additional["properties"].(map[string]any)
	if !ok {
		t.Fatalf("mcp_servers.additionalProperties.properties type = %T, want map[string]any", additional["properties"])
	}
	transportRaw, ok := properties["transport"]
	if !ok {
		t.Fatalf("mcp_servers schema missing transport property; got %v", keysAny(properties))
	}
	transport, ok := transportRaw.(map[string]any)
	if !ok {
		t.Fatalf("mcp_servers.transport type = %T, want map[string]any", transportRaw)
	}
	enumRaw, ok := transport["enum"]
	if !ok {
		t.Fatalf("mcp_servers.transport missing enum; got %v", transport)
	}
	enum, ok := enumRaw.([]string)
	if !ok {
		t.Fatalf("mcp_servers.transport.enum type = %T, want []string", enumRaw)
	}
	sort.Strings(enum)
	want := []string{"http", "stdio"}
	if len(enum) != len(want) {
		t.Fatalf("mcp_servers.transport.enum = %v, want %v", enum, want)
	}
	for i, allowed := range want {
		if enum[i] != allowed {
			t.Fatalf("mcp_servers.transport.enum[%d] = %q, want %q (full enum = %v)", i, enum[i], allowed, enum)
		}
	}
}

func Test_Default_ExtensionSchemas_exposes_command_and_url_on_mcp_servers(t *testing.T) {
	// The mcp_servers schema must surface both "command" and "url" so callers
	// can validate either shape (stdio launch vs http endpoint).
	m := Default()
	raw, ok := m.ExtensionSchemas["mcp_servers"]
	if !ok {
		t.Fatalf("Default().ExtensionSchemas missing key %q; keys = %v", "mcp_servers", keysAny(m.ExtensionSchemas))
	}
	mcpSchema, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("ExtensionSchemas[mcp_servers] type = %T, want map[string]any", raw)
	}
	additional, ok := mcpSchema["additionalProperties"].(map[string]any)
	if !ok {
		t.Fatalf("mcp_servers schema missing additionalProperties object; got %v", mcpSchema)
	}
	properties, ok := additional["properties"].(map[string]any)
	if !ok {
		t.Fatalf("mcp_servers.additionalProperties.properties type = %T, want map[string]any", additional["properties"])
	}
	for _, key := range []string{"command", "url"} {
		if _, ok := properties[key]; !ok {
			t.Fatalf("mcp_servers schema missing property %q; got %v", key, keysAny(properties))
		}
	}
}

func keysAny(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
