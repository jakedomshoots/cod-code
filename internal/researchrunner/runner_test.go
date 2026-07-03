package researchrunner

import (
	"context"
	"os"
	"strings"
	"testing"
)

func Test_Runner_Run_returns_failed_result_when_command_times_out(t *testing.T) {
	// Given
	runner := NewRunner()
	cmd := Command{
		Argv:      []string{os.Args[0], "-test.run=Test_HelperProcess_research_block"},
		Env:       []string{"GO_WANT_RESEARCH_HELPER=block"},
		Query:     "agent harness docs",
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
	if !strings.Contains(result.Error, "context deadline exceeded") {
		t.Fatalf("Error = %q, want context deadline exceeded", result.Error)
	}
}

func Test_HelperProcess_research_block(t *testing.T) {
	if os.Getenv("GO_WANT_RESEARCH_HELPER") != "block" {
		return
	}
	select {}
}
