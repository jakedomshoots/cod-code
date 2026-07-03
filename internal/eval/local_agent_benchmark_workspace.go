package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func prepareLocalAgentBenchmarkWorkspace(ctx context.Context, workspaceDir string, task Task, workspaceConfig []byte) error {
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("reset benchmark workspace: %w", err)
	}
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return fmt.Errorf("create benchmark workspace: %w", err)
	}
	if err := writeRelativeFile(workspaceDir, "go.mod", "module localagentbench\n\ngo 1.23\n"); err != nil {
		return err
	}
	if err := writeRelativeFile(workspaceDir, "internal/cli/doc.go", "package cli\n"); err != nil {
		return err
	}
	for _, path := range task.RequiredChangedFiles {
		if err := writeRelativeFile(workspaceDir, path, benchmarkBaselineText(task, path)); err != nil {
			return err
		}
	}
	if err := writeBenchmarkFixtureSupportFiles(workspaceDir, task); err != nil {
		return err
	}
	if len(workspaceConfig) > 0 {
		if err := writeRelativeFile(workspaceDir, ".ceo-harness.json", string(workspaceConfig)); err != nil {
			return err
		}
	}
	if err := runGitCommand(ctx, workspaceDir, "init"); err != nil {
		return err
	}
	if err := runGitCommand(ctx, workspaceDir, "config", "user.email", "eval@example.com"); err != nil {
		return err
	}
	if err := runGitCommand(ctx, workspaceDir, "config", "user.name", "Eval Fixture"); err != nil {
		return err
	}
	if err := runGitCommand(ctx, workspaceDir, "add", "."); err != nil {
		return err
	}
	if err := runGitCommand(ctx, workspaceDir, "commit", "-m", "benchmark baseline"); err != nil {
		return err
	}
	return nil
}

func benchmarkBaselineText(task Task, path string) string {
	if task.ID == "safety-policy-path-escape" && filepath.Clean(path) == filepath.Join("internal", "workspace", "workspace.go") {
		return safetyPathEscapeBaselineFixture()
	}
	switch filepath.Ext(path) {
	case ".go":
		return benchmarkGoFixture(path, "TODO: update this benchmark fixture")
	case ".js":
		return benchmarkJSFixture("TODO: update this benchmark fixture")
	case ".py":
		return benchmarkPythonFixture("TODO: update this benchmark fixture")
	}
	if task.ID == defaultLocalAgentBenchmarkID {
		return "# Roadmap\n\nGUI work is first.\n"
	}
	return "# Benchmark Fixture\n\nTODO: update this benchmark fixture.\n"
}

func benchmarkExpectedText(task Task, path string) string {
	if task.ID == "safety-policy-path-escape" && filepath.Clean(path) == filepath.Join("internal", "workspace", "workspace.go") {
		return safetyPathEscapeExpectedFixture()
	}
	expectedTerms := benchmarkExpectedTerms(task)
	switch filepath.Ext(path) {
	case ".go":
		return benchmarkGoFixture(path, expectedTerms)
	case ".js":
		return benchmarkJSFixture(expectedTerms)
	case ".py":
		return benchmarkPythonFixture(expectedTerms)
	}
	if task.ID == defaultLocalAgentBenchmarkID {
		return "# Roadmap\n\nCLI-first dogfood and recovery come before GUI work.\n"
	}
	return "# Benchmark Fixture\n\n" + expectedTerms + "\n"
}

func benchmarkExpectedTerms(task Task) string {
	terms := strings.Join(task.RequiredDiffTerms, " ")
	if strings.TrimSpace(terms) == "" {
		return task.Objective
	}
	return terms
}

func benchmarkGoFixture(path string, value string) string {
	return fmt.Sprintf("package %s\n\nconst benchmarkFixture = %q\n", benchmarkGoPackageName(path), value)
}

func benchmarkJSFixture(value string) string {
	return fmt.Sprintf("module.exports = { benchmarkFixture: %q };\n", value)
}

func benchmarkPythonFixture(value string) string {
	return fmt.Sprintf("benchmark_fixture = %q\n", value)
}

func benchmarkGoPackageName(path string) string {
	name := filepath.Base(filepath.Dir(filepath.Clean(path)))
	var builder strings.Builder
	for _, char := range name {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			builder.WriteRune(char)
			continue
		}
		builder.WriteByte('_')
	}
	clean := builder.String()
	if clean == "" || (clean[0] >= '0' && clean[0] <= '9') {
		return "fixture"
	}
	return clean
}

func writeBenchmarkStatusFile(ctx context.Context, path string, workspaceDir string) error {
	status, err := captureGitStatus(ctx, workspaceDir)
	if err != nil {
		return err
	}
	return writeTextFile(path, status)
}

func benchmarkChangedFiles(ctx context.Context, workspaceDir string) ([]string, string, error) {
	status, err := captureGitStatus(ctx, workspaceDir)
	if err != nil {
		return nil, "", err
	}
	files := dirtyPathsFromPorcelain(status)
	return files, status, nil
}

func writeBenchmarkArtifactsInWorkspace(workspaceDir string, task Task, content string) error {
	for _, artifact := range task.RequiredArtifacts {
		if err := writeRelativeFile(workspaceDir, artifact, content); err != nil {
			return err
		}
	}
	return nil
}

func benchmarkEvidenceContent(task Task, result LocalAgentBenchmarkResult, checks []commandResult) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# Benchmark Evidence\n\n")
	fmt.Fprintf(&builder, "Task: %s\n", task.ID)
	fmt.Fprintf(&builder, "Agent: %s\n", result.Name)
	for _, check := range checks {
		fmt.Fprintf(&builder, "- `%s`: %s exit=%d\n", strings.Join(check.Argv, " "), check.Status, check.ExitCode)
	}
	return builder.String()
}

func writeBenchmarkFixtureSupportFiles(workspaceDir string, task Task) error {
	if task.ID != "safety-policy-path-escape" {
		return writeGenericBenchmarkTests(workspaceDir, task)
	}
	return writeRelativeFile(workspaceDir, "internal/workspace/workspace_test.go", safetyPathEscapeTestFixture())
}

func writeGenericBenchmarkTests(workspaceDir string, task Task) error {
	for _, path := range task.RequiredChangedFiles {
		switch filepath.Ext(path) {
		case ".go":
			testPath := filepath.Join(filepath.Dir(filepath.Clean(path)), "benchmark_fixture_test.go")
			if err := writeRelativeFile(workspaceDir, testPath, benchmarkGoTestFixture(task, path)); err != nil {
				return err
			}
		case ".js":
			testPath := filepath.Join(filepath.Dir(filepath.Clean(path)), strings.TrimSuffix(filepath.Base(path), ".js")+".test.js")
			if err := writeRelativeFile(workspaceDir, testPath, benchmarkJSTestFixture(task, path)); err != nil {
				return err
			}
		case ".py":
			testPath := filepath.Join(filepath.Dir(filepath.Clean(path)), "test_"+filepath.Base(path))
			if err := writeRelativeFile(workspaceDir, testPath, benchmarkPythonTestFixture(task, path)); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}

func benchmarkGoTestFixture(task Task, path string) string {
	requiredTerms := task.RequiredDiffTerms
	if len(requiredTerms) == 0 {
		requiredTerms = []string{task.Objective}
	}
	return fmt.Sprintf(`package %s

import (
	"strings"
	"testing"
)

func %s(t *testing.T) {
	requiredTerms := %#v
	for _, required := range requiredTerms {
		if !strings.Contains(benchmarkFixture, required) {
			t.Fatalf("benchmarkFixture = %%q, want term %%q", benchmarkFixture, required)
		}
	}
}
`, benchmarkGoPackageName(path), benchmarkTestName(task), requiredTerms)
}

func benchmarkJSTestFixture(task Task, path string) string {
	requiredTerms := task.RequiredDiffTerms
	if len(requiredTerms) == 0 {
		requiredTerms = []string{task.Objective}
	}
	moduleName := "./" + strings.TrimSuffix(filepath.Base(path), ".js")
	return fmt.Sprintf(`const assert = require("assert");
const { benchmarkFixture } = require(%q);

for (const required of %s) {
  assert(
    benchmarkFixture.includes(required),
    "benchmarkFixture should include " + required + ", got " + benchmarkFixture,
  );
}
`, moduleName, benchmarkTermListLiteral(requiredTerms))
}

func benchmarkPythonTestFixture(task Task, path string) string {
	requiredTerms := task.RequiredDiffTerms
	if len(requiredTerms) == 0 {
		requiredTerms = []string{task.Objective}
	}
	moduleName := strings.TrimSuffix(filepath.Base(path), ".py")
	return fmt.Sprintf(`import importlib

module = importlib.import_module(%q)

for required in %s:
    assert required in module.benchmark_fixture, (
        f"benchmark_fixture should include {required!r}, got {module.benchmark_fixture!r}"
    )
`, moduleName, benchmarkTermListLiteral(requiredTerms))
}

func benchmarkTermListLiteral(requiredTerms []string) string {
	payload, err := json.Marshal(requiredTerms)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func benchmarkTestName(task Task) string {
	for _, command := range task.RequiredCommands {
		switch {
		case strings.Contains(command, "Test_Run_prints_provider_health"):
			return "Test_Run_prints_provider_health"
		case strings.Contains(command, "Test_ContextBudget"):
			return "Test_ContextBudget"
		case strings.Contains(command, "Test_Run_uses_provider"):
			return "Test_Run_uses_provider"
		case strings.Contains(command, "Test_RenderTextReport"):
			return "Test_RenderTextReport"
		case strings.Contains(command, "Test_CheckFixPrompt"):
			return "Test_CheckFixPrompt"
		case strings.Contains(command, "require_checks"):
			return "Test_Run_benchmark_require_checks"
		case strings.Contains(command, "Test_ProviderHealthPolicy"):
			return "Test_ProviderHealthPolicy"
		case strings.Contains(command, "Test_Provider"):
			return "Test_ProviderBenchmark"
		case strings.Contains(command, "Test_RunEvents"):
			return "Test_RunEvents"
		case strings.Contains(command, "Test_SmokeScript"):
			return "Test_SmokeScript"
		case strings.Contains(command, "Test_HTTPProvider"):
			return "Test_HTTPProvider"
		case strings.Contains(command, "provider_budget"):
			return "Test_Run_benchmark_provider_budget"
		case strings.Contains(command, "write_policy"):
			return "Test_Run_benchmark_write_policy"
		case strings.Contains(command, "Test_PatchApproval"):
			return "Test_PatchApproval"
		case strings.Contains(command, "Resume|Retry|continue"):
			return "Test_BenchmarkResume"
		case strings.Contains(command, "Test_Run_rollback_report"):
			return "Test_Run_rollback_report"
		}
	}
	return "Test_BenchmarkFixture"
}

func safetyPathEscapeBaselineFixture() string {
	return `package workspace

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathEscapesWorkspace = errors.New("unsafe path accepted")

func CleanRelativePath(path string) (string, error) {
	return filepath.Clean(strings.TrimSpace(path)), nil
}
`
}

func safetyPathEscapeExpectedFixture() string {
	return `package workspace

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathEscapesWorkspace = errors.New("path escapes workspace")

func CleanRelativePath(path string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || cleanPath == ".." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", ErrPathEscapesWorkspace
	}
	return cleanPath, nil
}
`
}

func safetyPathEscapeTestFixture() string {
	return `package workspace

import (
	"errors"
	"testing"
)

func Test_CleanRelativePath_rejects_parent_path_escape(t *testing.T) {
	_, err := CleanRelativePath("../outside.txt")
	if !errors.Is(err, ErrPathEscapesWorkspace) {
		t.Fatalf("error = %v, want path escapes workspace", err)
	}
}

func Test_CleanRelativePath_rejects_absolute_path(t *testing.T) {
	_, err := CleanRelativePath("/tmp/outside.txt")
	if !errors.Is(err, ErrPathEscapesWorkspace) {
		t.Fatalf("error = %v, want path escapes workspace", err)
	}
}

func Test_CleanRelativePath_accepts_nested_relative_path(t *testing.T) {
	path, err := CleanRelativePath("internal/workspace/workspace.go")
	if err != nil {
		t.Fatalf("CleanRelativePath returned error: %v", err)
	}
	if path != "internal/workspace/workspace.go" {
		t.Fatalf("path = %q, want internal/workspace/workspace.go", path)
	}
}
`
}
