package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/config"
)

func Test_Run_limits_tool_requests_with_flag(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("GO_WANT_CLI_TOOL_POLICY_MODEL", "1")
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_tool_policy_model",
		"--",
		"--max-tool-requests",
		"1",
		"Fix",
		"a",
		"bug",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName   string `json:"agent_name"`
			ToolResults []struct {
				Action string `json:"action"`
				Status string `json:"status"`
				Error  string `json:"error"`
			} `json:"tool_results"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	scanner := body.SubagentResults[0]
	if scanner.AgentName != "scanner" {
		t.Fatalf("first agent = %q, want scanner", scanner.AgentName)
	}
	if len(scanner.ToolResults) != 2 {
		t.Fatalf("ToolResults length = %d, want 2", len(scanner.ToolResults))
	}
	if scanner.ToolResults[0].Status != "pass" {
		t.Fatalf("ToolResults[0] = %+v, want pass", scanner.ToolResults[0])
	}
	if scanner.ToolResults[1].Status != "skipped" || !strings.Contains(scanner.ToolResults[1].Error, "tool request limit") {
		t.Fatalf("ToolResults[1] = %+v, want skipped limit", scanner.ToolResults[1])
	}
}

func Test_Run_uses_workspace_max_tool_requests_default(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("GO_WANT_CLI_TOOL_POLICY_MODEL", "1")
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello needle"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	configBody := `{"max_tool_requests":1,"model_command":["` + os.Args[0] + `","-test.run=Test_HelperProcess_cli_tool_policy_model"]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "bug"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			ToolResults []struct {
				Status string `json:"status"`
			} `json:"tool_results"`
		} `json:"subagent_results"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SubagentResults[0].ToolResults[1].Status != "skipped" {
		t.Fatalf("ToolResults = %+v, want second tool skipped", body.SubagentResults[0].ToolResults)
	}
}

func Test_Run_writes_max_tool_requests_when_init_config_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace",
		root,
		"--init-config",
		"--max-tool-requests",
		"2",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	cfg, loadErr := config.LoadWorkspace(context.Background(), root)
	if loadErr != nil {
		t.Fatalf("LoadWorkspace returned error: %v", loadErr)
	}
	if cfg.MaxToolRequests != 2 {
		t.Fatalf("MaxToolRequests = %d, want 2", cfg.MaxToolRequests)
	}
	var body struct {
		MaxToolRequests int `json:"max_tool_requests"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxToolRequests != 2 {
		t.Fatalf("MaxToolRequests report = %d, want 2", body.MaxToolRequests)
	}
}

func Test_Run_prints_max_tool_requests_config_check(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"max_tool_requests":3}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--config-check"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		MaxToolRequests int `json:"max_tool_requests"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.MaxToolRequests != 3 {
		t.Fatalf("MaxToolRequests = %d, want 3", body.MaxToolRequests)
	}
}

func Test_HelperProcess_cli_tool_policy_model(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_TOOL_POLICY_MODEL") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	switch {
	case strings.Contains(text, "agent: scanner") && strings.Contains(text, "tool_results:"):
		os.Stdout.WriteString(`{"summary":"used capped tool results"}`)
	case strings.Contains(text, "agent: scanner"):
		os.Stdout.WriteString(`{"summary":"need tools","tool_requests":[{"action":"read_workspace","path":"app.txt"},{"action":"search_workspace","query":"needle"}]}`)
	default:
		os.Stdout.WriteString(`{"summary":"ok"}`)
	}
	os.Exit(0)
}
