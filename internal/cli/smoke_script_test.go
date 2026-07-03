package cli

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func Test_SmokeScript_runs_core_health_checks(t *testing.T) {
	// Given
	script := filepath.Join("..", "..", "scripts", "smoke.sh")

	// When
	cmd := exec.Command("sh", script)
	output, err := cmd.CombinedOutput()
	// Then
	if err != nil {
		t.Fatalf("smoke script failed: %v\n%s", err, string(output))
	}
}
