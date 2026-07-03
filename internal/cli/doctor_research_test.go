package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_Run_prints_research_command_doctor_check_when_research_command_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	script := filepath.Join("..", "..", "examples", "research-command.sh")

	// When
	err := Run(context.Background(), &out, []string{"--doctor", "--research-command", "sh", script})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	found := false
	for _, check := range body.Checks {
		if check.Name == "research_command" && check.Status == "pass" && check.Source == "flag" && check.Error == "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Checks = %#v, want passing research_command check", body.Checks)
	}
}

func Test_Run_prints_workspace_source_for_configured_research_command_doctor_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	script, err := filepath.Abs(filepath.Join("..", "..", "examples", "research-command.sh"))
	if err != nil {
		t.Fatalf("research script path: %v", err)
	}
	configJSON := `{"research_command":["sh",` + strconv.Quote(script) + `]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	requireDoctorCheckSource(t, body.Checks, "research_command", "workspace")
}
