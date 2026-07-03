package model

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"testing"
	"time"
)

func Test_CommandClient_Complete_sends_prompt_to_command_stdin(t *testing.T) {
	// Given
	client, err := NewCommandClient(CommandSpec{
		Argv: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_model_command",
		},
		Env: []string{"GO_WANT_MODEL_HELPER=echo"},
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	response, err := client.Complete(context.Background(), Request{
		Prompt: "hello model",
	})

	// Then
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	if response.Text != "model saw: hello model" {
		t.Fatalf("Text = %q, want command stdout", response.Text)
	}
	if response.PromptBytes != len("hello model") {
		t.Fatalf("PromptBytes = %d, want prompt length", response.PromptBytes)
	}
}

func Test_CommandClient_Complete_times_out_command_when_timeout_is_set(t *testing.T) {
	// Given
	client, err := NewCommandClient(CommandSpec{
		Argv: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_model_command",
		},
		Env:       []string{"GO_WANT_MODEL_HELPER=block"},
		TimeoutMS: 1,
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	_, err = client.Complete(context.Background(), Request{
		Prompt: "hello model",
	})

	// Then
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Complete error = %v, want context deadline exceeded", err)
	}
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("Complete error = %T, want *CommandError", err)
	}
	if commandErr.Kind != CommandErrorKindTimeout {
		t.Fatalf("CommandError kind = %q, want %q", commandErr.Kind, CommandErrorKindTimeout)
	}
}

func Test_CommandClient_Complete_kills_shell_process_group_when_timeout_expires(t *testing.T) {
	// Given
	if runtime.GOOS == "windows" {
		t.Skip("shell process-group cancellation is Unix-specific")
	}
	client, err := NewCommandClient(CommandSpec{
		Argv:      []string{"sh", "-c", `printf "{\"summary\":\"model said ok\"}"; sleep 5`},
		TimeoutMS: 50,
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	startedAt := time.Now()
	_, err = client.Complete(context.Background(), Request{
		Prompt: "hello model",
	})
	elapsed := time.Since(startedAt)

	// Then
	if elapsed > 2*time.Second {
		t.Fatalf("Complete elapsed = %s, want prompt timeout under 2s", elapsed)
	}
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("Complete error = %T, want *CommandError", err)
	}
	if commandErr.Kind != CommandErrorKindTimeout {
		t.Fatalf("CommandError kind = %q, want %q", commandErr.Kind, CommandErrorKindTimeout)
	}
}

func Test_CommandClient_Complete_rejects_stdout_over_output_limit(t *testing.T) {
	// Given
	client, err := NewCommandClient(CommandSpec{
		Argv: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_model_command",
		},
		Env:            []string{"GO_WANT_MODEL_HELPER=large"},
		MaxOutputBytes: 8,
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	_, err = client.Complete(context.Background(), Request{
		Prompt: "hello model",
	})

	// Then
	if !errors.Is(err, ErrCommandOutputTooLarge) {
		t.Fatalf("Complete error = %v, want ErrCommandOutputTooLarge", err)
	}
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("Complete error = %T, want *CommandError", err)
	}
	if commandErr.Kind != CommandErrorKindOutputTooLarge {
		t.Fatalf("CommandError kind = %q, want %q", commandErr.Kind, CommandErrorKindOutputTooLarge)
	}
}

func Test_CommandClient_Complete_classifies_command_exit_failure(t *testing.T) {
	// Given
	client, err := NewCommandClient(CommandSpec{
		Argv: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_model_command",
		},
		Env: []string{"GO_WANT_MODEL_HELPER=fail"},
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	_, err = client.Complete(context.Background(), Request{
		Prompt: "hello model",
	})

	// Then
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("Complete error = %T, want *CommandError", err)
	}
	if commandErr.Kind != CommandErrorKindFailed || commandErr.ExitCode != 7 || commandErr.Stderr != "model broke\n" {
		t.Fatalf("CommandError = %#v, want failed exit code and stderr", commandErr)
	}
}

func Test_HelperProcess_model_command(t *testing.T) {
	switch os.Getenv("GO_WANT_MODEL_HELPER") {
	case "echo":
	case "block":
		select {}
	case "large":
		os.Stdout.WriteString("12345678901234567890")
		os.Exit(0)
	case "fail":
		os.Stderr.WriteString("model broke\n")
		os.Exit(7)
	default:
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	os.Stdout.WriteString("model saw: " + string(prompt))
	os.Exit(0)
}
