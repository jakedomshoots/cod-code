package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_StrictChecksScript_runsShellSyntaxFallbackWhenShellcheckMissing(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	temp := t.TempDir()
	bin := filepath.Join(temp, "bin")
	gopath := filepath.Join(temp, "gopath")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatalf("make fake bin: %v", err)
	}

	writeExecutable(t, filepath.Join(bin, "go"), `#!/bin/sh
set -eu
if [ "$1" = "env" ] && [ "$2" = "GOBIN" ]; then
  exit 0
fi
if [ "$1" = "env" ] && [ "$2" = "GOPATH" ]; then
  printf '%s\n' "$FAKE_GOPATH"
  exit 0
fi
printf '%s\n' "unexpected fake go command" >&2
exit 1
`)
	writeExecutable(t, filepath.Join(bin, "gofumpt"), "#!/bin/sh\nexit 0\n")
	writeExecutable(t, filepath.Join(bin, "golangci-lint"), "#!/bin/sh\nexit 0\n")
	writeExecutable(t, filepath.Join(bin, "nilaway"), "#!/bin/sh\nexit 0\n")

	cmd := exec.Command("sh", filepath.Join(root, "scripts", "strict-checks.sh"))
	cmd.Dir = root
	cmd.Env = append(
		cmd.Environ(),
		"PATH="+bin+":/bin:/usr/bin",
		"FAKE_GOPATH="+gopath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("strict-checks failed: %v\n%s", err, string(output))
	}
	body := string(output)
	if !strings.Contains(body, "strict-checks: shellcheck unavailable; ran sh -n on shell scripts") {
		t.Fatalf("strict-checks output missing shell syntax fallback:\n%s", body)
	}
	if strings.Contains(body, "shellcheck skipped") {
		t.Fatalf("strict-checks still reports shell lint skipped:\n%s", body)
	}
}

func writeExecutable(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable %s: %v", path, err)
	}
}
