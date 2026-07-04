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

func Test_Run_oauthList_printsCliLoginProviders(t *testing.T) {
	var out bytes.Buffer

	err := Run(context.Background(), &out, []string{"oauth", "list", "--format", "json"})

	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		Providers []struct {
			Name         string `json:"name"`
			AuthType     string `json:"auth_type"`
			TokenStorage bool   `json:"token_storage"`
			InitReady    bool   `json:"init_ready"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.Providers) < 5 {
		t.Fatalf("providers length = %d, want Kimi, Codex, Claude, OpenCode, and Goose", len(body.Providers))
	}
	seen := map[string]bool{}
	for _, provider := range body.Providers {
		if provider.AuthType != "cli_oauth" {
			t.Fatalf("provider %s auth_type = %q, want cli_oauth", provider.Name, provider.AuthType)
		}
		if provider.TokenStorage {
			t.Fatalf("provider %s stores tokens in harness, want false", provider.Name)
		}
		seen[provider.Name] = true
	}
	for _, name := range []string{"kimi", "codex", "claude", "opencode", "goose"} {
		if !seen[name] {
			t.Fatalf("providers = %+v, want %s", body.Providers, name)
		}
	}
}

func Test_Run_oauthInit_writesCommandProviderWithoutSecretEnv(t *testing.T) {
	for _, provider := range []string{"kimi", "codex", "claude", "opencode", "goose"} {
		t.Run(provider, func(t *testing.T) {
			root := t.TempDir()
			var out bytes.Buffer

			err := Run(context.Background(), &out, []string{
				"oauth", "init", provider,
				"--workspace", root,
				"--format", "json",
			})

			if err != nil {
				t.Fatalf("Run returned error: %v\n%s", err, out.String())
			}
			cfg, err := config.LoadWorkspace(context.Background(), root)
			if err != nil {
				t.Fatalf("LoadWorkspace returned error: %v", err)
			}
			mainProvider, ok := cfg.Providers["main"]
			if !ok {
				t.Fatalf("providers = %#v, want main", cfg.Providers)
			}
			if len(mainProvider.ModelCommand) != 2 || mainProvider.ModelCommand[0] != "sh" {
				t.Fatalf("main model_command = %#v, want sh script", mainProvider.ModelCommand)
			}
			wantScript := provider + "-model-command.sh"
			if filepath.Base(mainProvider.ModelCommand[1]) != wantScript {
				t.Fatalf("main model_command script = %q, want %s", mainProvider.ModelCommand[1], wantScript)
			}
			if len(mainProvider.EnvVars) != 0 || !mainProvider.HTTP.IsZero() {
				t.Fatalf("main provider = %#v, want no env vars and no HTTP key config", mainProvider)
			}
			if cfg.CEOProvider != "main" || cfg.ProviderPolicy.DefaultProvider != "main" || cfg.ProviderPolicy.FallbackProvider != "main" {
				t.Fatalf("routing = ceo %q policy %#v, want all main", cfg.CEOProvider, cfg.ProviderPolicy)
			}
			if !strings.Contains(out.String(), `"token_storage": false`) {
				t.Fatalf("output = %q, want no token storage field", out.String())
			}
		})
	}
}

func Test_Run_oauthDoctor_checksCliPresenceWithoutModelCalls(t *testing.T) {
	root := t.TempDir()
	writeFakeOAuthCLI(t, root, "kimi", "kimi version 1.2.3")
	t.Setenv("PATH", root)
	var out bytes.Buffer

	err := Run(context.Background(), &out, []string{"oauth", "doctor", "--format", "json"})

	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Providers []struct {
			Name    string `json:"name"`
			Status  string `json:"status"`
			Version string `json:"version,omitempty"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	statuses := map[string]string{}
	versions := map[string]string{}
	for _, provider := range body.Providers {
		statuses[provider.Name] = provider.Status
		versions[provider.Name] = provider.Version
	}
	if statuses["kimi"] != "ready" || !strings.Contains(versions["kimi"], "kimi version 1.2.3") {
		t.Fatalf("kimi doctor = status %q version %q, want ready with fake version", statuses["kimi"], versions["kimi"])
	}
	if statuses["codex"] != "missing_cli" {
		t.Fatalf("codex doctor status = %q, want missing_cli", statuses["codex"])
	}
}

func Test_Run_oauthRejectsUnknownProvider(t *testing.T) {
	err := Run(context.Background(), &bytes.Buffer{}, []string{"oauth", "init", "missing", "--workspace", t.TempDir()})

	if err == nil {
		t.Fatal("expected unknown OAuth provider error")
	}
	if !strings.Contains(err.Error(), `unknown oauth provider "missing"`) {
		t.Fatalf("error = %q, want unknown provider guidance", err.Error())
	}
}

func writeFakeOAuthCLI(t *testing.T, dir string, name string, version string) {
	t.Helper()
	path := filepath.Join(dir, name)
	body := "#!/bin/sh\nprintf '%s\\n' " + shellQuote(version) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
}
