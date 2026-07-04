package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Test_ProductionLocalGateScript_passesWhenOnlyPublicBlockersRemain(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	dist := filepath.Join(tmp, "dist")
	outputDir := filepath.Join(tmp, "production-local-gate")

	release := exec.Command("sh", filepath.Join(root, "scripts", "release-local.sh"))
	release.Dir = root
	release.Env = append(release.Environ(), "DIST="+dist, "VERSION=0.2.0-test", "COMMIT=abc123")
	output, err := release.CombinedOutput()
	if err != nil {
		t.Fatalf("release-local failed: %v\n%s", err, string(output))
	}

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-local-gate.sh"),
		"--dist", dist,
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("production-local-gate failed: %v\n%s", err, string(output))
	}
	body := string(output)
	for _, want := range []string{
		"production-local-gate: pass local_production_ready=true public_production_ready=false",
		"production-local-gate: blocked_count=",
		"production-local-gate: checklist_actions=",
		"production-local-gate: production_actions=",
		"production-local-gate: runnable_commands=",
		"production-local-gate: blocked_commands=",
		"production-local-gate: action_reasons=",
		"production-local-gate: finalizer_setup_actions=",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("production-local-gate output missing %q:\n%s", want, body)
		}
	}
	for _, path := range []string{
		filepath.Join(outputDir, "summary.json"),
		filepath.Join(outputDir, "index.md"),
		filepath.Join(outputDir, "launch-checklist.md"),
		filepath.Join(outputDir, "production-readiness.stdout.txt"),
		filepath.Join(outputDir, "production-readiness.stderr.txt"),
		filepath.Join(outputDir, "production-actions.json"),
		filepath.Join(outputDir, "production-actions.commands.sh"),
		filepath.Join(outputDir, "production-actions.stderr.txt"),
		filepath.Join(outputDir, "production-status.json"),
		filepath.Join(outputDir, "production-status.stderr.txt"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected evidence file %s: %v", path, err)
		}
	}
	commands := readTextFile(t, filepath.Join(outputDir, "production-actions.commands.sh"))
	if !strings.Contains(commands, "# blocked command:") {
		t.Fatalf("production action command artifact should comment blocked commands:\n%s", commands)
	}
	if !strings.Contains(commands, " reason: ") {
		t.Fatalf("production action command artifact should include blocker reasons:\n%s", commands)
	}
	actions := readTextFile(t, filepath.Join(outputDir, "production-actions.json"))
	for _, want := range []string{
		`"path":`,
		`"action_reason":`,
		`"action_state":`,
		`"action_state_counts":`,
		`"required_action_count":`,
		`"runnable_command_count":`,
		`"blocked_command_count":`,
		`"evidence_declared_match_count":`,
		`"evidence_declared_mismatch_count": 0`,
	} {
		if !strings.Contains(actions, want) {
			t.Fatalf("production action artifact missing %q:\n%s", want, actions)
		}
	}
	status := readTextFile(t, filepath.Join(outputDir, "production-status.json"))
	for _, want := range []string{
		`"finalizer_next_actions":`,
		`"matches_declared": true`,
		`"setup_sha256":`,
		`"setup_matches_declared": true`,
		`"setup_required_action_count":`,
		`"evidence_declared_mismatch_count": 0`,
	} {
		if !strings.Contains(status, want) {
			t.Fatalf("production status artifact missing %q:\n%s", want, status)
		}
	}
}

func Test_ProductionLocalGateScript_failsWhenLocalReadinessIsFalse(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	tmp := t.TempDir()
	outputDir := filepath.Join(tmp, "production-local-gate")

	cmd := exec.Command(
		"sh",
		filepath.Join(root, "scripts", "production-local-gate.sh"),
		"--dist", filepath.Join(root, "dist"),
		"--evidence-root", filepath.Join(tmp, "empty-evidence"),
		"--output-dir", outputDir,
	)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("production-local-gate unexpectedly passed with missing dist:\n%s", string(output))
	}
	if !strings.Contains(string(output), "production-local-gate: fail local_production_ready=false") {
		t.Fatalf("production-local-gate output missing local failure:\n%s", string(output))
	}
}
