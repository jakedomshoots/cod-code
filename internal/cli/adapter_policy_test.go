package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_write_policy_dry_run_previews_patch_without_writing(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "dry-run",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(content) != "hello old" {
		t.Fatalf("content = %q, want unchanged", string(content))
	}
	var body struct {
		PatchApproval struct {
			Status string `json:"status"`
		} `json:"patch_approval"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.PatchApproval.Status != "previewed" {
		t.Fatalf("PatchApproval.Status = %q, want previewed", body.PatchApproval.Status)
	}
}

func Test_Run_write_policy_approved_write_requires_preview_digest(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &bytes.Buffer{}, []string{
		"--workspace", root,
		"--write-policy", "approved-write",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	if err == nil {
		t.Fatal("expected approved-write guidance error")
	}
	if !strings.Contains(err.Error(), "--write-policy approved-write requires --approve-preview") {
		t.Fatalf("error = %q, want approved-write guidance", err.Error())
	}
}

func Test_Run_init_config_uses_external_adapter_preset(t *testing.T) {
	for _, adapter := range []string{"codex", "claude", "opencode", "aider", "goose"} {
		t.Run(adapter, func(t *testing.T) {
			// Given
			root := t.TempDir()
			var out bytes.Buffer

			// When
			err := Run(context.Background(), &out, []string{
				"--workspace", root,
				"--init-config",
				"--adapter", adapter,
			})
			// Then
			if err != nil {
				t.Fatalf("Run returned error: %v\n%s", err, out.String())
			}
			cfg, err := config.LoadWorkspace(context.Background(), root)
			if err != nil {
				t.Fatalf("LoadWorkspace returned error: %v", err)
			}
			if len(cfg.ModelCommand) != 2 || !strings.Contains(cfg.ModelCommand[1], adapter+".sh") {
				t.Fatalf("ModelCommand = %#v, want %s adapter script", cfg.ModelCommand, adapter)
			}
		})
	}
}

func Test_Run_ConfigInit_usesKimiAdapterPreset_whenAdvertised(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"config", "init",
		"--workspace", root,
		"--adapter", "kimi",
		"--format", "text",
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	cfg, err := config.LoadWorkspace(context.Background(), root)
	if err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
	if len(cfg.ModelCommand) != 2 || cfg.ModelCommand[0] != "sh" || filepath.Base(cfg.ModelCommand[1]) != "kimi-model-command.sh" {
		t.Fatalf("ModelCommand = %#v, want sh scripts/kimi-model-command.sh", cfg.ModelCommand)
	}
	if !filepath.IsAbs(cfg.ModelCommand[1]) {
		t.Fatalf("Kimi adapter path = %q, want absolute path", cfg.ModelCommand[1])
	}
	if !strings.Contains(out.String(), `"adapter": "kimi"`) {
		t.Fatalf("output = %q, want adapter report", out.String())
	}
}

func Test_Run_ConfigInit_rejectsUnknownAdapterPreset(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"config", "init",
		"--workspace", root,
		"--adapter", "missing",
	})

	// Then
	if err == nil {
		t.Fatal("expected unknown adapter error")
	}
	if !strings.Contains(err.Error(), `unknown --adapter "missing"`) {
		t.Fatalf("error = %q, want unknown adapter guidance", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(root, config.WorkspaceConfigName)); !os.IsNotExist(statErr) {
		t.Fatalf("workspace config stat error = %v, want no config written", statErr)
	}
}
