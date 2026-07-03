package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_start_initializes_workspace_and_prints_operator_next_steps(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--start", root, "--format", "text"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{"Start: pass", "Workspace: " + root, "Config:", "Doctor: pass", "Next:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("start output missing %q:\n%s", want, text)
		}
	}
	if _, err := config.LoadWorkspace(context.Background(), root); err != nil {
		t.Fatalf("LoadWorkspace returned error: %v", err)
	}
}

func Test_Run_inbox_prints_review_queue_with_details_as_text(t *testing.T) {
	// Given
	root := t.TempDir()
	saveReviewDetailJob(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--inbox"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{"Review queue: 1 job", "Action: answer subagent questions", "Question:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("inbox output missing %q:\n%s", want, text)
		}
	}
}

func Test_Run_provider_wizard_writes_openai_provider_config(t *testing.T) {
	// Given
	root := t.TempDir()
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--provider-wizard", "openai",
		"--http-model", "gpt-5",
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
	if cfg.CEOProvider != "main" || cfg.ProviderPolicy.DefaultProvider != "main" {
		t.Fatalf("provider routes = ceo:%q default:%q, want main/main", cfg.CEOProvider, cfg.ProviderPolicy.DefaultProvider)
	}
	provider := cfg.Providers["main"].HTTP
	if provider.Model != "gpt-5" || provider.APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("provider = %#v, want OpenAI preset with gpt-5", provider)
	}
	if !strings.Contains(out.String(), "Provider wizard: openai") {
		t.Fatalf("wizard text missing heading:\n%s", out.String())
	}
}

func Test_Run_init_demo_repo_creates_runnable_golden_repo(t *testing.T) {
	// Given
	root := filepath.Join(t.TempDir(), "golden")
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--init-demo-repo", root, "--format", "text"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	for _, name := range []string{"README.md", "app.txt", "Makefile", config.WorkspaceConfigName} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("expected demo file %s: %v", name, err)
		}
	}
	if !strings.Contains(out.String(), "Golden demo repo: created") {
		t.Fatalf("demo repo text missing created status:\n%s", out.String())
	}
}

func Test_Run_tui_prints_operator_dashboard(t *testing.T) {
	// Given
	root := t.TempDir()
	saveReviewDetailJob(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--tui"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	text := out.String()
	for _, want := range []string{"CEO Harness TUI", "Inbox: 1", "Next:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("tui output missing %q:\n%s", want, text)
		}
	}
}
