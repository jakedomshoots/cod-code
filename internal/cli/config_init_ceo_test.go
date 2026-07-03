package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_writes_ceo_model_command_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	args := []string{
		"--workspace",
		root,
		"--init-config",
		"--ceo-model-command",
		"python3",
		"-c",
		"print(\"review\")",
		"--",
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
	if len(cfg.CEOModelCommand) != 3 {
		t.Fatalf("CEOModelCommand length = %d, want 3", len(cfg.CEOModelCommand))
	}
	var body struct {
		CEOModelCommandArgc int `json:"ceo_model_command_argc"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOModelCommandArgc != 3 {
		t.Fatalf("CEOModelCommandArgc = %d, want 3", body.CEOModelCommandArgc)
	}
}
