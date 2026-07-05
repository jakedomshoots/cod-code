package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_Run_prints_version_when_version_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--version"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "cod dev" {
		t.Fatalf("version output = %q, want cod dev", got)
	}
}

func Test_Run_prints_build_metadata_when_version_metadata_is_set(t *testing.T) {
	// Given
	oldVersion := Version
	oldCommit := Commit
	oldBuildDate := BuildDate
	Version = "1.2.3"
	Commit = "abc123"
	BuildDate = "2026-07-01"
	t.Cleanup(func() {
		Version = oldVersion
		Commit = oldCommit
		BuildDate = oldBuildDate
	})
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--version"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	want := "cod 1.2.3 commit=abc123 built=2026-07-01"
	if got := strings.TrimSpace(out.String()); got != want {
		t.Fatalf("version output = %q, want %s", got, want)
	}
}
