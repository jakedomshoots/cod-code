package eval

import (
	"fmt"
	"path/filepath"
	"strings"
)

func buildLocalAgentSpec(id string, ceoBinary string, task localAgentTaskSpec) (localAgentSpec, error) {
	switch id {
	case "ceo_harness":
		return buildCEOHarnessSpec(id, ceoBinary, task), nil
	case "codex_cli":
		return buildCodexCLISpec(id, task), nil
	case "claude_code":
		return buildClaudeCodeSpec(id, task), nil
	case "aider":
		return buildAiderSpec(id, task), nil
	case "opencode":
		return buildOpenCodeSpec(id, task), nil
	case "goose":
		return buildGooseSpec(id, task), nil
	case "pi":
		return buildPiSpec(id, task), nil
	case "oh_my_pi":
		return buildOhMyPiSpec(id, task), nil
	default:
		return localAgentSpec{}, fmt.Errorf("%w: unknown local agent %q", ErrInvalidCompetitor, id)
	}
}

func buildCEOHarnessSpec(id string, ceoBinary string, task localAgentTaskSpec) localAgentSpec {
	binary := strings.TrimSpace(ceoBinary)
	if binary == "" {
		binary = filepath.Join(".", "bin", "ceo-packet")
	}
	args := []string{"--plan-only", "--format", "json", task.prompt}
	expectedOutput := "plan_only"
	if task.name == localAgentTaskEditFile {
		args = []string{"--write-policy", "trusted-local", "--replace", "app.txt", "old", "new", "--format", "json", task.prompt}
		expectedOutput = `"verdict"`
	}
	return localAgentSpec{
		id:             id,
		name:           "CEO Harness",
		binary:         binary,
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Build CEO Harness with `make build` before running local comparisons.",
	}
}

func buildCodexCLISpec(id string, task localAgentTaskSpec) localAgentSpec {
	sandbox := "read-only"
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		sandbox = "workspace-write"
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "OpenAI Codex CLI",
		binary:         "codex",
		args:           []string{"exec", "--ephemeral", "--ignore-user-config", "--ignore-rules", "--sandbox", sandbox, "--skip-git-repo-check", task.prompt},
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install and authenticate Codex CLI before real comparison runs.",
	}
}

func buildOpenCodeSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"run", "--pure", "--format", "json", task.prompt}
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		args = []string{"run", "--pure", "--auto", "--format", "json", task.prompt}
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "OpenCode",
		binary:         "opencode",
		args:           args,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install and authenticate OpenCode before real comparison runs.",
	}
}

func buildPiSpec(id string, task localAgentTaskSpec) localAgentSpec {
	args := []string{"--no-session", "--no-tools", "--offline", "-p", task.prompt}
	env := []string{"PI_OFFLINE=1"}
	expectedOutput := localAgentMarker
	if task.name == localAgentTaskEditFile {
		args = []string{"--no-session", "--approve", "-p", task.prompt}
		env = nil
		expectedOutput = ""
	}
	return localAgentSpec{
		id:             id,
		name:           "Pi",
		binary:         "pi",
		args:           args,
		env:            env,
		expectedOutput: expectedOutput,
		expectedFile:   task.expectedFile,
		setupHint:      "Install Pi and configure a provider key before real comparison runs.",
	}
}
