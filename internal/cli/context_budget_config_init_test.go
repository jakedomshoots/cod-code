package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_max_context_bytes_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--max-context-bytes",
		"1024",
	}

	// When
	err := Run(context.Background(), &out, args)
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.MaxContextBytes != 1024 {
		t.Fatalf("MaxContextBytes = %d, want 1024", cfg.MaxContextBytes)
	}
	var body struct {
		MaxContextBytes int `json:"max_context_bytes"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxContextBytes != 1024 {
		t.Fatalf("report MaxContextBytes = %d, want 1024", body.MaxContextBytes)
	}
}

func Test_Run_writes_workspace_brief_excludes_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--workspace-brief-exclude",
		"generated",
		"--workspace-brief-exclude",
		"*.lock",
	}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if len(cfg.WorkspaceBriefExcludes) != 2 || cfg.WorkspaceBriefExcludes[0] != "generated" || cfg.WorkspaceBriefExcludes[1] != "*.lock" {
		t.Fatalf("WorkspaceBriefExcludes = %#v, want generated and *.lock", cfg.WorkspaceBriefExcludes)
	}
	var body struct {
		WorkspaceBriefExcludeCount int `json:"workspace_brief_exclude_count"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.WorkspaceBriefExcludeCount != 2 {
		t.Fatalf("WorkspaceBriefExcludeCount = %d, want 2", body.WorkspaceBriefExcludeCount)
	}
}
