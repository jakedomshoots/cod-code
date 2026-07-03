package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_KimiModelCommandScript_includesRequiredFileContents(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "internal", "workspace"), 0o755); err != nil {
		t.Fatalf("create workspace dirs: %v", err)
	}
	requiredPath := filepath.Join(workspace, "internal", "workspace", "workspace.go")
	if err := os.WriteFile(requiredPath, []byte("package workspace\n\nconst benchmarkFixture = \"old\"\n"), 0o644); err != nil {
		t.Fatalf("write required file: %v", err)
	}
	testPath := filepath.Join(workspace, "internal", "workspace", "workspace_test.go")
	if err := os.WriteFile(testPath, []byte("package workspace\n\nfunc TestBenchmarkFixture(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatalf("write sibling test file: %v", err)
	}
	binDir := t.TempDir()
	capturedPrompt := filepath.Join(t.TempDir(), "prompt.txt")
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), `#!/bin/sh
while [ "$#" -gt 0 ]; do
  case "$1" in
    -p)
      shift
      printf '%s\n' "$1" > "$CAPTURED_PROMPT"
      ;;
  esac
  shift || true
done
cat <<'JSON'
{"role":"assistant","content":"{\"status\":\"needs_input\",\"summary\":\"captured\"}"}
JSON
`)

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "kimi-model-command.sh"))
	cmd.Dir = workspace
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"), "CAPTURED_PROMPT="+capturedPrompt)
	cmd.Stdin = strings.NewReader("Required changed files: internal/workspace/workspace.go.\nRequired evidence artifacts: .omo/evidence/safety-policy-path-escape.md.\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kimi model command failed: %v\n%s", err, string(output))
	}
	prompt := readTextFile(t, capturedPrompt)
	for _, want := range []string{
		"Workspace snippets:",
		"File: internal/workspace/workspace.go",
		"const benchmarkFixture = \"old\"",
		"File: internal/workspace/workspace_test.go",
		"TestBenchmarkFixture",
		"Required evidence artifacts parsed from harness prompt:",
		"- .omo/evidence/safety-policy-path-escape.md",
		".omo/evidence/safety-policy-path-escape.md",
		"ceo_delegation must include a non-empty selected_subagents array",
		"Do not contradict guard_verdict, checks, patch_results, or workspace snippets",
		"Return needs_input only for a missing user decision",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("captured prompt missing %q:\n%s", want, prompt)
		}
	}
}

func Test_KimiModelCommandScript_includesJavaScriptSiblingTestContents(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "frontend"), 0o755); err != nil {
		t.Fatalf("create frontend dir: %v", err)
	}
	requiredPath := filepath.Join(workspace, "frontend", "state.js")
	if err := os.WriteFile(requiredPath, []byte("module.exports = { benchmarkFixture: \"TODO\" };\n"), 0o644); err != nil {
		t.Fatalf("write required JS file: %v", err)
	}
	testPath := filepath.Join(workspace, "frontend", "state.test.js")
	if err := os.WriteFile(testPath, []byte("assert(benchmarkFixture.includes('optimistic update'));\n"), 0o644); err != nil {
		t.Fatalf("write sibling JS test file: %v", err)
	}
	binDir := t.TempDir()
	capturedPrompt := filepath.Join(t.TempDir(), "prompt.txt")
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), `#!/bin/sh
while [ "$#" -gt 0 ]; do
  case "$1" in
    -p)
      shift
      printf '%s\n' "$1" > "$CAPTURED_PROMPT"
      ;;
  esac
  shift || true
done
cat <<'JSON'
{"role":"assistant","content":"{\"status\":\"needs_input\",\"summary\":\"captured\"}"}
JSON
`)

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "kimi-model-command.sh"))
	cmd.Dir = workspace
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"), "CAPTURED_PROMPT="+capturedPrompt)
	cmd.Stdin = strings.NewReader("Required changed files: frontend/state.js.\nRequired evidence artifacts: .omo/evidence/cross-language-js-state-reducer.md.\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kimi model command failed: %v\n%s", err, string(output))
	}
	prompt := readTextFile(t, capturedPrompt)
	for _, want := range []string{
		"File: frontend/state.js",
		"benchmarkFixture",
		"File: frontend/state.test.js",
		"optimistic update",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("captured prompt missing %q:\n%s", want, prompt)
		}
	}
}

func Test_KimiModelCommandScript_runsKimiOutsideWorkspace(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "internal", "workspace"), 0o755); err != nil {
		t.Fatalf("create workspace dirs: %v", err)
	}
	requiredPath := filepath.Join(workspace, "internal", "workspace", "workspace.go")
	original := "package workspace\n\nconst benchmarkFixture = \"old\"\n"
	if err := os.WriteFile(requiredPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write required file: %v", err)
	}
	binDir := t.TempDir()
	kimiCWD := filepath.Join(t.TempDir(), "cwd.txt")
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), `#!/bin/sh
printf '%s\n' "$PWD" > "$KIMI_CWD"
mkdir -p internal/workspace
printf 'package workspace

const benchmarkFixture = "hacked"
' > internal/workspace/workspace.go
cat <<'JSON'
{"role":"assistant","content":"{\"status\":\"needs_input\",\"summary\":\"captured\"}"}
JSON
`)

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "kimi-model-command.sh"))
	cmd.Dir = workspace
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"), "KIMI_CWD="+kimiCWD)
	cmd.Stdin = strings.NewReader("Required changed files: internal/workspace/workspace.go.\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kimi model command failed: %v\n%s", err, string(output))
	}
	got := readTextFile(t, requiredPath)
	if got != original {
		t.Fatalf("workspace file changed by kimi cwd side effect:\n%s", got)
	}
	if strings.TrimSpace(readTextFile(t, kimiCWD)) == workspace {
		t.Fatalf("kimi ran inside scored workspace %q", workspace)
	}
}

func Test_KimiModelCommandScript_normalizesPatchShorthand(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "internal", "workspace"), 0o755); err != nil {
		t.Fatalf("create workspace dirs: %v", err)
	}
	requiredPath := filepath.Join(workspace, "internal", "workspace", "workspace.go")
	original := "package workspace\n\nconst benchmarkFixture = \"old\"\n"
	replacement := "package workspace\n\nconst benchmarkFixture = \"new\"\n"
	if err := os.WriteFile(requiredPath, []byte(original), 0o644); err != nil {
		t.Fatalf("write required file: %v", err)
	}
	binDir := t.TempDir()
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), `#!/bin/sh
cat <<'JSON'
{"role":"assistant","content":"{\"status\":\"pass\",\"summary\":\"patch ready\",\"patches\":[{\"path\":\"internal/workspace/workspace.go\",\"content\":\"package workspace\\n\\nconst benchmarkFixture = \\\"new\\\"\\n\"},{\"path\":\".omo/evidence/safety-policy-path-escape.md\",\"new\":\"# Evidence\\n\"}]}"}
JSON
`)

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "kimi-model-command.sh"))
	cmd.Dir = workspace
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Stdin = strings.NewReader("Required changed files: internal/workspace/workspace.go.\nRequired evidence artifacts: .omo/evidence/safety-policy-path-escape.md.\n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("kimi model command failed: %v\n%s", err, string(output))
	}
	var body struct {
		Patches []struct {
			Path    string `json:"path"`
			Old     string `json:"old,omitempty"`
			New     string `json:"new,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"patches"`
	}
	if err := json.Unmarshal(output, &body); err != nil {
		t.Fatalf("parse normalized output: %v\n%s", err, string(output))
	}
	if len(body.Patches) != 2 {
		t.Fatalf("patches = %+v, want two normalized patches", body.Patches)
	}
	if body.Patches[0].Path != "internal/workspace/workspace.go" || body.Patches[0].Old != original || body.Patches[0].New != replacement || body.Patches[0].Content != "" {
		t.Fatalf("patch[0] = %+v, want existing-file old/new normalization", body.Patches[0])
	}
	if body.Patches[1].Path != ".omo/evidence/safety-policy-path-escape.md" || body.Patches[1].Content != "# Evidence\n" || body.Patches[1].Old != "" || body.Patches[1].New != "" {
		t.Fatalf("patch[1] = %+v, want new-file content normalization", body.Patches[1])
	}
}

func Test_KimiModelCommandScript_rejectsPatchMissingRequiredDiffTerm(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	workspace := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workspace, "internal", "workspace"), 0o755); err != nil {
		t.Fatalf("create workspace dirs: %v", err)
	}
	requiredPath := filepath.Join(workspace, "internal", "workspace", "workspace.go")
	if err := os.WriteFile(requiredPath, []byte("package workspace\n\nfunc CleanRelativePath(path string) string { return path }\n"), 0o644); err != nil {
		t.Fatalf("write required file: %v", err)
	}
	binDir := t.TempDir()
	writeExecutableScript(t, filepath.Join(binDir, "kimi"), `#!/bin/sh
cat <<'JSON'
{"role":"assistant","content":"{\"status\":\"pass\",\"summary\":\"patch ready\",\"patches\":[{\"path\":\"internal/workspace/workspace.go\",\"old\":\"return path\",\"new\":\"return clean\"}]}"}
JSON
`)

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "kimi-model-command.sh"))
	cmd.Dir = workspace
	cmd.Env = append(cmd.Environ(), "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cmd.Stdin = strings.NewReader("Required changed files: internal/workspace/workspace.go.\nRequired diff terms: path escapes workspace.\n")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("kimi model command unexpectedly passed:\n%s", string(output))
	}
	if !strings.Contains(string(output), "required diff terms missing from patch") {
		t.Fatalf("output = %s, want required diff term rejection", string(output))
	}
}

func writeExecutableScript(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}
