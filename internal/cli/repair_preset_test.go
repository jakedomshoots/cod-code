package cli

import (
	"strings"
	"testing"
)

func Test_ParseArgs_applies_standard_repair_preset(t *testing.T) {
	// Given
	args := []string{"--repair-preset", "standard", "Fix", "the", "bug"}

	// When
	opts, err := parseArgs(args)

	// Then
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if opts.checkFixAttempts != 1 {
		t.Fatalf("checkFixAttempts = %d, want 1", opts.checkFixAttempts)
	}
	if opts.ceoRevisionAttempts != 1 {
		t.Fatalf("ceoRevisionAttempts = %d, want 1", opts.ceoRevisionAttempts)
	}
	if opts.maxCEOIterations != 3 {
		t.Fatalf("maxCEOIterations = %d, want 3", opts.maxCEOIterations)
	}
}

func Test_ParseArgs_keeps_explicit_repair_flags_when_standard_preset_is_supplied(t *testing.T) {
	// Given
	args := []string{
		"--check-fix-attempts", "2",
		"--repair-preset", "standard",
		"--ceo-revision-attempts", "3",
		"--max-ceo-iterations", "4",
		"Fix",
		"the",
		"bug",
	}

	// When
	opts, err := parseArgs(args)

	// Then
	if err != nil {
		t.Fatalf("parseArgs returned error: %v", err)
	}
	if opts.checkFixAttempts != 2 {
		t.Fatalf("checkFixAttempts = %d, want explicit 2", opts.checkFixAttempts)
	}
	if opts.ceoRevisionAttempts != 3 {
		t.Fatalf("ceoRevisionAttempts = %d, want explicit 3", opts.ceoRevisionAttempts)
	}
	if opts.maxCEOIterations != 4 {
		t.Fatalf("maxCEOIterations = %d, want explicit 4", opts.maxCEOIterations)
	}
}

func Test_ParseArgs_returns_clear_error_when_repair_preset_is_unknown(t *testing.T) {
	// Given
	args := []string{"--repair-preset", "chaos", "Fix", "the", "bug"}

	// When
	_, err := parseArgs(args)

	// Then
	if err == nil {
		t.Fatal("expected repair preset error")
	}
	if !strings.Contains(err.Error(), `--repair-preset "chaos" is not supported`) {
		t.Fatalf("error = %q, want unsupported repair preset", err.Error())
	}
}
