package ceo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/toolmanifest"
)

func runToolRequest(t *testing.T, req JobRequest, request subagent.ToolRequest) subagent.ToolResult {
	t.Helper()
	runtime := NewRuntime()
	state := toolRequestState{Request: req}
	result := subagent.Result{
		AgentName:      "scanner",
		Role:           "browser",
		Status:         "pass",
		AllowedActions: jobpacket.ActionStrings(allowedActionsForTest(req)),
	}
	return runtime.runSubagentToolRequest(context.Background(), result, request, state)
}

func allowedActionsForTest(req JobRequest) []jobpacket.Action {
	return []jobpacket.Action{
		jobpacket.ActionBrowserRead,
		jobpacket.ActionComputerSnapshot,
		jobpacket.ActionToolManifest,
	}
}

func Test_RunSubagentToolRequest_browser_read_allow_localhost_returns_receipt_and_permission(t *testing.T) {
	// Given
	body := "<html><head><title>allowed</title></head><body>hello</body></html>"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	// When
	tr := runToolRequest(t, JobRequest{ /* BrowserPolicy omitted -> defaults to allow-localhost */ }, subagent.ToolRequest{
		Action:   string(jobpacket.ActionBrowserRead),
		URL:      server.URL,
		MaxBytes: 4096,
	})

	// Then
	if tr.Status != "pass" {
		t.Fatalf("Status = %q, want pass (default policy is allow-localhost): output=%q err=%q", tr.Status, tr.Output, tr.Error)
	}
	if tr.Permission != "allow-localhost" {
		t.Fatalf("Permission = %q, want allow-localhost (browser defaults to allow-localhost)", tr.Permission)
	}
	if tr.URL != server.URL {
		t.Fatalf("URL = %q, want %q", tr.URL, server.URL)
	}
	if tr.ReceiptSHA256 == "" {
		t.Fatal("ReceiptSHA256 is empty; expected a sha256 digest")
	}
	if _, err := hex.DecodeString(tr.ReceiptSHA256); err != nil || len(tr.ReceiptSHA256) != sha256.Size*2 {
		t.Fatalf("ReceiptSHA256 = %q, want %d-char hex sha256 digest", tr.ReceiptSHA256, sha256.Size*2)
	}
	if !strings.Contains(tr.Output, "hello") {
		t.Fatalf("Output = %q, want http body content", tr.Output)
	}
}

func Test_RunSubagentToolRequest_browser_read_blocks_non_localhost_under_localhost_policy(t *testing.T) {
	// Given — explicitly pin the policy to allow-localhost so a non-loopback URL is denied.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("served on 127.0.0.1 — must not be reachable"))
	}))
	defer server.Close()

	// When
	tr := runToolRequest(t, JobRequest{BrowserPolicy: "allow-localhost"}, subagent.ToolRequest{
		Action: string(jobpacket.ActionBrowserRead),
		URL:    server.URL,
	})

	// Then — httptest.NewServer binds to 127.0.0.1, so this URL IS localhost reachable.
	if tr.Status != "pass" {
		t.Fatalf("Status = %q, want pass for 127.0.0.1 under allow-localhost: err=%q", tr.Status, tr.Error)
	}
	if tr.Permission != "allow-localhost" {
		t.Fatalf("Permission = %q, want allow-localhost", tr.Permission)
	}

	// And explicitly verify the deny branch for a non-loopback host that allow-localhost rejects.
	denied := runToolRequest(t, JobRequest{BrowserPolicy: "allow-localhost"}, subagent.ToolRequest{
		Action: string(jobpacket.ActionBrowserRead),
		URL:    "http://example.com/",
	})
	if denied.Status != "denied" {
		t.Fatalf("Status = %q, want denied for non-localhost under allow-localhost: err=%q", denied.Status, denied.Error)
	}
	if denied.Permission != "allow-localhost" {
		t.Fatalf("Permission = %q, want allow-localhost even on deny", denied.Permission)
	}
	if denied.ReceiptSHA256 == "" {
		t.Fatal("ReceiptSHA256 is empty on denied result; expected denied digest")
	}
}

func Test_RunSubagentToolRequest_computer_snapshot_default_policy_denies_without_invoking_command(t *testing.T) {
	// Given — default ComputerPolicy ("") normalizes to "ask", and the configured command is harmless.
	request := subagent.ToolRequest{
		Action: string(jobpacket.ActionComputerSnapshot),
		App:    "Finder",
	}
	req := JobRequest{
		// ComputerPolicy omitted -> defaults to ask.
		ComputerCommand: []string{"sh", "-c", "echo SHOULD_NOT_RUN >/dev/null; exit 99"},
	}

	// When
	tr := runToolRequest(t, req, request)

	// Then
	if tr.Status != "denied" {
		t.Fatalf("Status = %q, want denied under default ask policy: err=%q", tr.Status, tr.Error)
	}
	if tr.Permission != "ask" {
		t.Fatalf("Permission = %q, want ask (computer default policy)", tr.Permission)
	}
	if tr.App != "Finder" {
		t.Fatalf("App = %q, want Finder", tr.App)
	}
	if !strings.Contains(tr.Error, "explicit operator approval") {
		t.Fatalf("Error = %q, want explicit operator approval wording", tr.Error)
	}
	if tr.ReceiptSHA256 == "" {
		t.Fatal("ReceiptSHA256 is empty on denied result")
	}
	if _, err := hex.DecodeString(tr.ReceiptSHA256); err != nil || len(tr.ReceiptSHA256) != sha256.Size*2 {
		t.Fatalf("ReceiptSHA256 = %q, want hex sha256 digest", tr.ReceiptSHA256)
	}
	// The command must not have been invoked — Status="denied" occurs before exec.CommandContext.
	if tr.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0 because the command was not executed", tr.ExitCode)
	}
	if strings.Contains(tr.Output, "SHOULD_NOT_RUN") {
		t.Fatalf("Output = %q, must not contain command stdout (command must not run)", tr.Output)
	}
}

func Test_RunSubagentToolRequest_tool_manifest_returns_manifest_json(t *testing.T) {
	// Given
	wantNames := map[string]string{
		"browser.read":      string(jobpacket.ActionBrowserRead),
		"computer.snapshot": string(jobpacket.ActionComputerSnapshot),
		"tools.manifest":    string(jobpacket.ActionToolManifest),
	}

	// When
	tr := runToolRequest(t, JobRequest{}, subagent.ToolRequest{
		Action: string(jobpacket.ActionToolManifest),
	})

	// Then
	if tr.Status != "pass" {
		t.Fatalf("Status = %q, want pass: err=%q", tr.Status, tr.Error)
	}
	if tr.Tool != "tools.manifest" {
		t.Fatalf("Tool = %q, want tools.manifest", tr.Tool)
	}
	if tr.Bytes <= 0 {
		t.Fatalf("Bytes = %d, want >0", tr.Bytes)
	}
	var got toolmanifest.Manifest
	if err := json.Unmarshal([]byte(tr.Output), &got); err != nil {
		t.Fatalf("manifest JSON must unmarshal: %v\n%s", err, tr.Output)
	}
	if got.SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", got.SchemaVersion)
	}
	if got.Name != "ceo-packet-tools" {
		t.Fatalf("Name = %q, want ceo-packet-tools", got.Name)
	}
	toolByName := map[string]toolmanifest.Tool{}
	for _, tool := range got.Tools {
		toolByName[tool.Name] = tool
	}
	for name, action := range wantNames {
		tool, ok := toolByName[name]
		if !ok {
			t.Fatalf("manifest missing tool %q (advertised tools: %v)", name, toolNames(got.Tools))
		}
		if tool.Action != action {
			t.Fatalf("manifest tool %q action = %q, want %q", name, tool.Action, action)
		}
	}
}

func toolNames(tools []toolmanifest.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}
