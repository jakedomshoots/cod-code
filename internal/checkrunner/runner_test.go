package checkrunner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func Test_Runner_Run_returns_passed_result_when_command_exits_zero(t *testing.T) {
	// Given
	runner := NewRunner()
	cmd := Command{
		Argv: []string{os.Args[0], "-test.run=Test_HelperProcess_pass"},
		Env:  []string{"GO_WANT_HELPER_PROCESS=pass"},
	}

	// When
	result, err := runner.Run(context.Background(), cmd)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("Status = %q, want pass", result.Status)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "helper pass\n" {
		t.Fatalf("Stdout = %q, want helper pass", result.Stdout)
	}
	if result.DurationMS < 0 {
		t.Fatalf("DurationMS = %d, want nonnegative duration", result.DurationMS)
	}
}

func Test_Runner_Run_returns_failed_result_when_command_exits_nonzero(t *testing.T) {
	// Given
	runner := NewRunner()
	cmd := Command{
		Argv: []string{os.Args[0], "-test.run=Test_HelperProcess_fail"},
		Env:  []string{"GO_WANT_HELPER_PROCESS=fail"},
	}

	// When
	result, err := runner.Run(context.Background(), cmd)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Status != "fail" {
		t.Fatalf("Status = %q, want fail", result.Status)
	}
	if result.ExitCode == 0 {
		t.Fatal("expected nonzero exit code")
	}
	if result.Stderr != "helper fail\n" {
		t.Fatalf("Stderr = %q, want helper fail", result.Stderr)
	}
}

func Test_Runner_Run_uses_working_directory_when_configured(t *testing.T) {
	// Given
	runner := NewRunner()
	root := t.TempDir()
	cmd := Command{
		Argv:    []string{os.Args[0], "-test.run=Test_HelperProcess_pwd"},
		Env:     []string{"GO_WANT_HELPER_PROCESS=pwd"},
		WorkDir: root,
	}

	// When
	result, err := runner.Run(context.Background(), cmd)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("Status = %q, want pass", result.Status)
	}
	wantRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve root: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != wantRoot {
		t.Fatalf("Stdout = %q, want working directory %q", result.Stdout, wantRoot)
	}
}

func Test_Runner_Run_returns_failed_result_when_command_times_out(t *testing.T) {
	// Given
	runner := NewRunner()
	cmd := Command{
		Argv:      []string{os.Args[0], "-test.run=Test_HelperProcess_block"},
		Env:       []string{"GO_WANT_HELPER_PROCESS=block"},
		TimeoutMS: 1,
	}

	// When
	result, err := runner.Run(context.Background(), cmd)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Status != "fail" || result.ExitCode != -1 {
		t.Fatalf("result = %#v, want timeout failure", result)
	}
	if !strings.Contains(result.Stderr, "context deadline exceeded") {
		t.Fatalf("Stderr = %q, want context deadline exceeded", result.Stderr)
	}
}

func Test_Runner_Run_cancels_shell_process_group_when_timeout_expires(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("shell process-group cancellation is Unix-specific")
	}
	runner := NewRunner()
	cmd := Command{
		Argv:      []string{"sh", "-c", `printf "PASS\n"; sleep 10`},
		TimeoutMS: 250,
	}

	// When
	startedAt := time.Now()
	result, err := runner.Run(context.Background(), cmd)
	elapsed := time.Since(startedAt)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Run elapsed = %s, want prompt timeout under 2s", elapsed)
	}
	if result.Status == "pass" {
		t.Fatalf("Status = %q with stdout %q, want timeout failure", result.Status, result.Stdout)
	}
	if result.ExitCode != -1 {
		t.Fatalf("ExitCode = %d, want timeout exit code -1", result.ExitCode)
	}
	if !strings.Contains(result.Stdout, "PASS") {
		t.Fatalf("Stdout = %q, want misleading PASS output captured", result.Stdout)
	}
}

func Test_Runner_Run_cancels_shell_process_group_when_parent_context_expires(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("shell process-group cancellation is Unix-specific")
	}
	runner := NewRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	cmd := Command{Argv: []string{"sh", "-c", `printf "PASS\n"; sleep 10`}}

	// When
	startedAt := time.Now()
	result, err := runner.Run(ctx, cmd)
	elapsed := time.Since(startedAt)

	// Then
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run error = %v, want context deadline exceeded", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Run elapsed = %s, want prompt parent-context cancellation under 2s", elapsed)
	}
	if result.Status == "pass" {
		t.Fatalf("Status = %q, want no passing result after parent cancellation", result.Status)
	}
}

func Test_HelperProcess_pass(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "pass" {
		return
	}
	os.Stdout.WriteString("helper pass\n")
	os.Exit(0)
}

func Test_HelperProcess_fail(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "fail" {
		return
	}
	os.Stderr.WriteString("helper fail\n")
	os.Exit(7)
}

func Test_HelperProcess_pwd(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "pwd" {
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	os.Stdout.WriteString(wd + "\n")
	os.Exit(0)
}

func Test_HelperProcess_block(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "block" {
		return
	}
	select {}
}
