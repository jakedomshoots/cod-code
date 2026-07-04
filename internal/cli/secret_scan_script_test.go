package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_SecretScanScript_blocksSecretsInReleaseFiles(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	scanRoot := t.TempDir()
	writeText(t, filepath.Join(scanRoot, "README.md"), "OPENAI_API_KEY=sk-proj-liveleakvalue1234567890\n")

	cmd := exec.Command("sh", filepath.Join(repoRoot, "scripts", "secret-scan.sh"), "--root", scanRoot)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("secret scan unexpectedly passed:\n%s", string(output))
	}
	body := string(output)
	if !strings.Contains(body, "README.md") || !strings.Contains(body, "possible secret") {
		t.Fatalf("secret scan output missing file and reason:\n%s", body)
	}
}

func Test_SecretScanScript_allowsPlaceholdersAndTestFixtures(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	scanRoot := t.TempDir()
	writeText(t, filepath.Join(scanRoot, "README.md"), "export OPENAI_API_KEY=...\n")
	writeText(t, filepath.Join(scanRoot, "internal", "cli", "context_trace_test.go"), "const fake = \"OPENAI_API_KEY=sk-proj-testfixture1234567890\"\n")

	cmd := exec.Command("sh", filepath.Join(repoRoot, "scripts", "secret-scan.sh"), "--root", scanRoot)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("secret scan should allow placeholders and test fixtures: %v\n%s", err, string(output))
	}
	if !strings.Contains(string(output), "secret-scan ok") {
		t.Fatalf("secret scan output missing success marker:\n%s", string(output))
	}
}

func writeText(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
