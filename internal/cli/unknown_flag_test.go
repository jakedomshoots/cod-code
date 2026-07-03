package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_rejects_unknown_flag(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--versoin"})

	// Then
	if err == nil {
		t.Fatal("expected unknown flag error")
	}
	if !strings.Contains(err.Error(), "unknown flag --versoin") {
		t.Fatalf("error = %q, want unknown flag", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty output", out.String())
	}
}
