package cli

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_InstallLocalScript_installs_binary_to_prefix(t *testing.T) {
	// Given
	root := filepath.Join("..", "..")
	script := filepath.Join("scripts", "install-local.sh")
	prefix := t.TempDir()

	// When
	cmd := exec.Command("sh", script)
	cmd.Dir = root
	cmd.Env = append(cmd.Environ(), "PREFIX="+prefix, "VERSION=0.2.0-test", "COMMIT=script-test")
	output, err := cmd.CombinedOutput()
	// Then
	if err != nil {
		t.Fatalf("install script failed: %v\n%s", err, string(output))
	}
	binary := filepath.Join(prefix, "bin", "cod")
	versionOutput, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("installed binary failed: %v\n%s", err, string(versionOutput))
	}
	versionText := string(versionOutput)
	if !strings.Contains(versionText, "cod 0.2.0-test commit=script-test") {
		t.Fatalf("version output = %q, want installed metadata", versionText)
	}
}
