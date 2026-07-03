package model

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_CommandClient_Complete_exposes_request_metadata_to_command_env(t *testing.T) {
	// Given
	client, err := NewCommandClient(CommandSpec{
		Argv: []string{
			os.Args[0],
			"-test.run=Test_HelperProcess_model_metadata_command",
		},
		Env: []string{"GO_WANT_MODEL_METADATA_HELPER=1"},
	})
	if err != nil {
		t.Fatalf("NewCommandClient returned error: %v", err)
	}

	// When
	response, err := client.Complete(context.Background(), Request{
		Prompt: "hello model",
		Metadata: RequestMetadata{
			Kind:        "subagent",
			AgentName:   "scanner",
			AgentRole:   "inspect scope",
			ContextMode: "lean",
		},
	})
	// Then
	if err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}
	for _, want := range []string{
		"kind=subagent",
		"agent=scanner",
		"role=inspect scope",
		"context=lean",
	} {
		if !strings.Contains(response.Text, want) {
			t.Fatalf("Text = %q, want %q", response.Text, want)
		}
	}
}

func Test_HelperProcess_model_metadata_command(t *testing.T) {
	if os.Getenv("GO_WANT_MODEL_METADATA_HELPER") != "1" {
		return
	}
	if _, err := io.ReadAll(os.Stdin); err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	os.Stdout.WriteString("kind=" + os.Getenv("CEO_MODEL_REQUEST_KIND") + "\n")
	os.Stdout.WriteString("agent=" + os.Getenv("CEO_AGENT_NAME") + "\n")
	os.Stdout.WriteString("role=" + os.Getenv("CEO_AGENT_ROLE") + "\n")
	os.Stdout.WriteString("context=" + os.Getenv("CEO_CONTEXT_MODE") + "\n")
	os.Exit(0)
}
