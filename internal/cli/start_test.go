package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// Test_Run_start_text_next_steps_expose_boring_first_run_path pins the start
// report's "Next:" section to the recommended first-run command path. When the
// workspace is provided, each step must reference that workspace so the user can
// copy/paste the command directly. Anything else (a different task text, a
// different plan-only smoke command, a missing --format text) breaks the
// onboarding contract.
func Test_Run_start_text_next_steps_expose_boring_first_run_path(t *testing.T) {
	// Given
	root := t.TempDir()
	workspace := workspaceArg(root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--start", root, "--format", "text"})
	if err != nil {
		t.Fatalf("Run --start returned error: %v\n%s", err, out.String())
	}

	// Then
	text := out.String()
	nextIndex := strings.Index(text, "Next:\n")
	if nextIndex < 0 {
		t.Fatalf("start text missing Next: section:\n%s", text)
	}
	nextBlock := text[nextIndex:]

	wantLines := []string{
		"ceo-packet oauth doctor --format text",
		"ceo-packet oauth init kimi --workspace " + workspace + " --format text",
		"ceo-packet run --workspace " + workspace + " --check go test ./... -- \"Fix one real task\"",
		"ceo-packet production-status --workspace " + workspace + " --format text",
	}
	for _, want := range wantLines {
		if !strings.Contains(nextBlock, want) {
			t.Fatalf("start Next: missing %q:\n%s", want, text)
		}
	}
}
