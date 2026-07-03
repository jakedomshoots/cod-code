package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_example_adapter_commands_when_init_example_adapters_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{"--workspace", root, "--init-config", "--init-example-adapters"}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	requireExampleCommand(t, cfg.ModelCommand, "command-model.sh")
	requireExampleCommand(t, cfg.CEOModelCommand, "ceo-model.sh")
	requireExampleCommand(t, cfg.ResearchCommand, "research-command.sh")

	var body struct {
		ExampleAdapters     bool `json:"example_adapters"`
		ModelCommandArgc    int  `json:"model_command_argc"`
		CEOModelCommandArgc int  `json:"ceo_model_command_argc"`
		ResearchCommandArgc int  `json:"research_command_argc"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if !body.ExampleAdapters || body.ModelCommandArgc != 2 || body.CEOModelCommandArgc != 2 || body.ResearchCommandArgc != 2 {
		t.Fatalf("init report = %#v, want example adapter argc values", body)
	}
}

func Test_Run_rejects_init_example_adapters_without_init_config(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--init-example-adapters", "Fix", "tests"})

	// Then
	if err == nil {
		t.Fatal("expected init example adapters usage error")
	}
	if !strings.Contains(err.Error(), "--init-example-adapters requires --init-config") {
		t.Fatalf("error = %q, want init-config guidance", err.Error())
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty output", out.String())
	}
}

func requireExampleCommand(t *testing.T, command []string, scriptName string) {
	t.Helper()
	if len(command) != 2 {
		t.Fatalf("%s command = %#v, want sh plus script path", scriptName, command)
	}
	if command[0] != "sh" {
		t.Fatalf("%s command[0] = %q, want sh", scriptName, command[0])
	}
	if filepath.Base(command[1]) != scriptName {
		t.Fatalf("%s command path = %q, want %s", scriptName, command[1], scriptName)
	}
	if !filepath.IsAbs(command[1]) {
		t.Fatalf("%s command path = %q, want absolute path", scriptName, command[1])
	}
}
