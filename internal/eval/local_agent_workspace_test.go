package eval

import (
	"slices"
	"strings"
	"testing"
)

// Test_LocalAgentCommand_ompInjectsCwdWithoutDisturbingArgs verifies that the
// omp binary receives a --cwd <workspace> prefix and that the agent's own args
// (including the trailing prompt) stay intact in their original order.
func Test_LocalAgentCommand_ompInjectsCwdWithoutDisturbingArgs(t *testing.T) {
	workspace := "/tmp/omp-workspace"
	args := []string{"--no-session", "--auto-approve", "--print", "hello world"}

	command := localAgentCommand("omp", args, workspace)

	want := []string{"omp", "--cwd", workspace, "--no-session", "--auto-approve", "--print", "hello world"}
	if !slices.Equal(command, want) {
		t.Fatalf("command = %+v, want %+v", command, want)
	}
}

// Test_LocalAgentCommand_existingBinariesUndisturbed pins the previously
// wired workspace injection for ceo-packet, opencode, and codex. A change
// that accidentally drops or duplicates the workspace flag for these agents
// must fail the test.
func Test_LocalAgentCommand_existingBinariesUndisturbed(t *testing.T) {
	workspace := "/tmp/ws"
	t.Run("ceo-packet uses --workspace", func(t *testing.T) {
		got := localAgentCommand("/usr/local/bin/ceo-packet", []string{"--plan-only", "ping"}, workspace)
		want := []string{"/usr/local/bin/ceo-packet", "--workspace", workspace, "--plan-only", "ping"}
		if !slices.Equal(got, want) {
			t.Fatalf("command = %+v, want %+v", got, want)
		}
	})
	t.Run("opencode uses --dir", func(t *testing.T) {
		got := localAgentCommand("opencode", []string{"run", "--pure", "--format", "json", "ping"}, workspace)
		want := []string{"opencode", "run", "--dir", workspace, "--pure", "--format", "json", "ping"}
		if !slices.Equal(got, want) {
			t.Fatalf("command = %+v, want %+v", got, want)
		}
	})
	t.Run("codex uses -C", func(t *testing.T) {
		got := localAgentCommand("codex", []string{"exec", "--ephemeral", "--sandbox", "read-only", "ping"}, workspace)
		want := []string{"codex", "exec", "-C", workspace, "--ephemeral", "--sandbox", "read-only", "ping"}
		if !slices.Equal(got, want) {
			t.Fatalf("command = %+v, want %+v", got, want)
		}
	})
	t.Run("aider passthrough with --no-git safety", func(t *testing.T) {
		got := localAgentCommand("aider", []string{"--no-git", "--no-gitignore", "--yes-always", "--message", "ping"}, workspace)
		want := []string{"aider", "--no-git", "--no-gitignore", "--yes-always", "--message", "ping"}
		if !slices.Equal(got, want) {
			t.Fatalf("command = %+v, want %+v", got, want)
		}
		if len(got) != 6 {
			t.Fatalf("command length = %d, want 6 (no workspace injection for aider)", len(got))
		}
	})
}

// Test_LocalAgentCommand_aiderNoGitMutationsSurvive is the contract guard for
// the new Aider passthrough: the safety flags must still appear in the command
// the harness executes, even though Aider has no workspace flag.
func Test_LocalAgentCommand_aiderNoGitMutationsSurvive(t *testing.T) {
	got := localAgentCommand("aider", []string{"--no-git", "--no-gitignore", "--no-auto-commits", "--yes-always", "--message", "app.txt"}, "/tmp/aider-ws")
	joined := strings.Join(got, " ")
	for _, want := range []string{"--no-git", "--no-gitignore", "--no-auto-commits", "--yes-always"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("command %q, missing Aider safety flag %q", joined, want)
		}
	}
}
