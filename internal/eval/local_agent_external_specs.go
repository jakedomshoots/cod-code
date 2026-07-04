package eval

func buildClaudeCodeSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"--safe-mode", "--no-session-persistence", "--permission-mode", "plan", "--print", task.prompt}
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		args = []string{"--safe-mode", "--no-session-persistence", "--permission-mode", "bypassPermissions", "--output-format", "json", "--print", task.prompt}
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "Claude Code",
		binary:         "claude",
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install and authenticate Claude Code before benchmark runs.",
	}
}

func aiderHistoryArgs() []string {
	return []string{
		"--input-history-file", "/dev/null",
		"--chat-history-file", "/dev/null",
		"--llm-history-file", "/dev/null",
		"--no-restore-chat-history",
	}
}

func buildAiderSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"--no-git", "--no-gitignore", "--no-auto-commits", "--no-pretty", "--no-stream", "--no-analytics", "--no-check-update", "--yes-always"}
	args = append(args, aiderHistoryArgs()...)
	args = append(args, "--message", task.prompt)
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		args = append(args[:len(args)-2], "--file", "app.txt", "--message", task.prompt)
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "Aider",
		binary:         "aider",
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install Aider and configure a provider before benchmark runs.",
	}
}

func buildGooseSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"run", "--no-session", "--quiet", "--text", task.prompt}
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "Goose",
		binary:         "goose",
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install Goose and configure a provider before benchmark runs.",
	}
}

func buildOhMyPiSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"--no-session", "--no-tools", "--no-rules", "--no-skills", "--max-time", "45", "--print", task.prompt}
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		args = []string{"--no-session", "--auto-approve", "--approval-mode", "yolo", "--no-rules", "--no-skills", "--max-time", "240", "--print", task.prompt}
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "Oh My Pi",
		binary:         "omp",
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install Oh My Pi and configure a provider before benchmark runs.",
	}
}

func buildClaudeCodeBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, prompt string) localAgentSpec {
	args := []string{"--safe-mode", "--no-session-persistence", "--permission-mode", "bypassPermissions", "--output-format", "json"}
	args = appendAgentModelArgs(args, req.AgentModels, id)
	args = append(args, "--print", prompt)
	return localAgentSpec{id: id, name: "Claude Code", binary: "claude", args: args, setupHint: "Install and authenticate Claude Code before benchmark runs."}
}

func buildAiderBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, task Task, prompt string) localAgentSpec {
	args := []string{"--git", "--no-gitignore", "--skip-sanity-check-repo", "--no-auto-commits", "--no-dirty-commits", "--no-pretty", "--no-stream", "--no-analytics", "--no-check-update", "--yes-always", "--map-tokens", "0"}
	args = append(args, aiderHistoryArgs()...)
	args = appendAgentModelArgs(args, req.AgentModels, id)
	for _, path := range task.RequiredChangedFiles {
		args = append(args, "--file", path)
	}
	args = append(args, "--message", prompt)
	return localAgentSpec{id: id, name: "Aider", binary: "aider", args: args, setupHint: "Install Aider and configure a provider before benchmark runs."}
}

func buildGooseBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, prompt string) localAgentSpec {
	args := []string{"run", "--no-session", "--quiet", "--max-turns", "20"}
	args = appendAgentModelArgs(args, req.AgentModels, id)
	args = append(args, "--text", prompt)
	return localAgentSpec{id: id, name: "Goose", binary: "goose", args: args, setupHint: "Install Goose and configure a provider before benchmark runs."}
}

func buildOhMyPiBenchmarkSpec(id string, req LocalAgentBenchmarkRequest, prompt string) localAgentSpec {
	args := []string{"--no-session", "--auto-approve", "--approval-mode", "yolo", "--no-rules", "--no-skills", "--max-time", "240"}
	args = appendAgentModelArgs(args, req.AgentModels, id)
	args = append(args, "--print", prompt)
	return localAgentSpec{id: id, name: "Oh My Pi", binary: "omp", args: args, setupHint: "Install Oh My Pi and configure a provider before benchmark runs."}
}
