package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_rejects_unknown_provider_health_recommendation_filter(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--recommendation", "bad"})

	// Then
	if err == nil {
		t.Fatalf("Run returned nil error, want invalid recommendation error")
	}
	if !strings.Contains(err.Error(), "--recommendation must be one of avoid, watch, healthy, unknown") {
		t.Fatalf("Run error = %v, want recommendation validation error", err)
	}
}
