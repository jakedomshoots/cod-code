package adapter

func SupportedTools() []Tool {
	return []Tool{
		{ID: ToolCodex, DisplayName: "Codex CLI", EnvVar: "CEO_CODEX_ADAPTER_COMMAND", SetupDoc: "docs/adapters/codex.md"},
		{ID: ToolClaude, DisplayName: "Claude Code", EnvVar: "CEO_CLAUDE_ADAPTER_COMMAND", SetupDoc: "docs/adapters/claude.md"},
		{ID: ToolOpenCode, DisplayName: "OpenCode", EnvVar: "CEO_OPENCODE_ADAPTER_COMMAND", SetupDoc: "docs/adapters/opencode.md"},
		{ID: ToolAider, DisplayName: "Aider", EnvVar: "CEO_AIDER_ADAPTER_COMMAND", SetupDoc: "docs/adapters/aider.md"},
		{ID: ToolGoose, DisplayName: "Goose", EnvVar: "CEO_GOOSE_ADAPTER_COMMAND", SetupDoc: "docs/adapters/goose.md"},
	}
}

func ToolByID(id ToolID) (Tool, bool) {
	for _, tool := range SupportedTools() {
		if tool.ID == id {
			return tool, true
		}
	}
	return Tool{}, false
}
