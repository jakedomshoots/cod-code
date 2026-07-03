package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func optionsWithQuickstartDefaults(opts options) options {
	if len(opts.checkCommand) == 0 {
		if command, ok := quickstartCheckCommand(opts.quickstartDir); ok {
			opts.checkCommand = command
		}
	}
	if len(opts.checkCommand) > 0 {
		opts.requireChecks = true
	}
	return opts
}

func quickstartCheckCommand(workspaceDir string) ([]string, bool) {
	if regularFileExists(filepath.Join(workspaceDir, "go.mod")) {
		return []string{"go", "test", "./..."}, true
	}
	if regularFileExists(filepath.Join(workspaceDir, "Cargo.toml")) {
		return []string{"cargo", "test"}, true
	}
	if packageJSONHasTestScript(filepath.Join(workspaceDir, "package.json")) {
		return packageTestCommand(workspaceDir), true
	}
	if pythonProjectUsesPytest(workspaceDir) {
		return pythonTestCommand(workspaceDir), true
	}
	if makefileHasTestTarget(workspaceDir) {
		return []string{"make", "test"}, true
	}
	return nil, false
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && !info.IsDir()
}

func packageJSONHasTestScript(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var packageFile struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(content, &packageFile); err != nil {
		return false
	}
	return strings.TrimSpace(packageFile.Scripts["test"]) != ""
}

func packageTestCommand(workspaceDir string) []string {
	if regularFileExists(filepath.Join(workspaceDir, "bun.lock")) || regularFileExists(filepath.Join(workspaceDir, "bun.lockb")) {
		return []string{"bun", "test"}
	}
	if regularFileExists(filepath.Join(workspaceDir, "pnpm-lock.yaml")) {
		return []string{"pnpm", "test"}
	}
	if regularFileExists(filepath.Join(workspaceDir, "yarn.lock")) {
		return []string{"yarn", "test"}
	}
	return []string{"npm", "test"}
}

func pythonProjectUsesPytest(workspaceDir string) bool {
	return regularFileExists(filepath.Join(workspaceDir, "pytest.ini")) ||
		fileContains(filepath.Join(workspaceDir, "tox.ini"), "[pytest]") ||
		fileContains(filepath.Join(workspaceDir, "pyproject.toml"), "[tool.pytest.ini_options]")
}

func pythonTestCommand(workspaceDir string) []string {
	if regularFileExists(filepath.Join(workspaceDir, "uv.lock")) {
		return []string{"uv", "run", "pytest"}
	}
	return []string{"python", "-m", "pytest"}
}

func fileContains(path string, needle string) bool {
	content, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(content), needle)
}

func makefileHasTestTarget(workspaceDir string) bool {
	return fileHasMakeTarget(filepath.Join(workspaceDir, "Makefile"), "test") ||
		fileHasMakeTarget(filepath.Join(workspaceDir, "makefile"), "test") ||
		fileHasMakeTarget(filepath.Join(workspaceDir, "GNUmakefile"), "test")
}

func fileHasMakeTarget(path string, target string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		name, _, ok := strings.Cut(trimmed, ":")
		if ok && name == target {
			return true
		}
	}
	return false
}
