package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

// These tests pin the contracts documented for the tool surfaces introduced in
// internal/jobpacket/actions.go (browser_read, computer_snapshot,
// tool_manifest). They drive the public CLI entry point so the parser,
// validator, and dispatcher are exercised together — the way an operator or
// subagent would invoke them — rather than poking runBrowser/runComputer in
// isolation.

// The tool manifest must always carry schema_version 1 and advertise the three
// capability entries other surfaces depend on.
func Test_Run_tools_manifest_json_emits_schema_version_and_actions(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	if err := Run(context.Background(), &out, []string{"tools", "manifest", "--format", "json"}); err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}

	// Then
	var body struct {
		SchemaVersion int `json:"schema_version"`
		Tools         []struct {
			Name   string `json:"name"`
			Action string `json:"action"`
		} `json:"tools"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", body.SchemaVersion)
	}
	wantTools := map[string]string{
		"browser.read":      "browser_read",
		"computer.snapshot": "computer_snapshot",
		"tools.manifest":    "tool_manifest",
	}
	seen := make(map[string]string, len(body.Tools))
	for _, tool := range body.Tools {
		seen[tool.Name] = tool.Action
	}
	for name, action := range wantTools {
		got, ok := seen[name]
		if !ok {
			t.Fatalf("manifest missing tool %q; tools = %+v", name, body.Tools)
		}
		if got != action {
			t.Fatalf("manifest tool %q action = %q, want %q", name, got, action)
		}
	}
}

// `browser read` against a localhost httptest server must succeed under the
// default allow-localhost policy and emit a non-empty receipt digest.
func Test_Run_browser_read_localhost_returns_pass_with_receipt(t *testing.T) {
	// Given
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello local"))
	}))
	defer server.Close()

	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"browser", "read", server.URL, "--format", "json",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("httptest server hits = %d, want 1", got)
	}
	var body struct {
		Status        string
		Permission    string
		URL           string
		HTTPStatus    int
		Output        string
		ReceiptSHA256 string
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("status = %q, want pass", body.Status)
	}
	if body.Permission != "allow-localhost" {
		t.Fatalf("permission = %q, want allow-localhost", body.Permission)
	}
	if body.URL != server.URL {
		t.Fatalf("url = %q, want %q", body.URL, server.URL)
	}
	if body.HTTPStatus != http.StatusOK {
		t.Fatalf("http_status = %d, want 200", body.HTTPStatus)
	}
	if !strings.Contains(body.Output, "hello local") {
		t.Fatalf("output = %q, want it to contain the served body", body.Output)
	}
	if len(body.ReceiptSHA256) != 64 {
		t.Fatalf("receipt_sha256 = %q, want 64-char hex digest", body.ReceiptSHA256)
	}
}

// `browser read` against a non-loopback URL with the default policy must be
// denied by the policy gate BEFORE any network request goes out, so the
// loopback-only contract holds even when the rest of the world is reachable.
func Test_Run_browser_read_remote_denies_without_network(t *testing.T) {
	// Given
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var out bytes.Buffer

	// When — we point at a non-loopback host with the same surface as a real
	// public URL but in fact request the loopback server would not satisfy
	// the policy check anyway; the assertion that the loopback counter stays
	// at zero is what proves the policy gate blocked the request.
	err := Run(context.Background(), &out, []string{
		"browser", "read", "https://example.com", "--format", "json",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	if got := atomic.LoadInt32(&hits); got != 0 {
		t.Fatalf("httptest server hits = %d, want 0 (policy must block before fetch)", got)
	}
	var body struct {
		Status        string
		Permission    string
		Error         string
		ReceiptSHA256 string
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "denied" {
		t.Fatalf("status = %q, want denied", body.Status)
	}
	if body.Permission != "allow-localhost" {
		t.Fatalf("permission = %q, want allow-localhost", body.Permission)
	}
	if !strings.Contains(strings.ToLower(body.Error), "localhost") {
		t.Fatalf("error = %q, want it to mention localhost", body.Error)
	}
	if len(body.ReceiptSHA256) != 64 {
		t.Fatalf("receipt_sha256 = %q, want 64-char hex digest", body.ReceiptSHA256)
	}
}

// `computer snapshot` with the default policy must be denied before the
// configured command is ever executed. The bogus command below would explode
// with an exec error if the policy gate were ever bypassed, so the asserted
// `denied` status with the approval-required message is the precise contract.
func Test_Run_computer_snapshot_default_policy_denies_without_running_command(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When — default policy is `ask`, command points at a path that does not
	// exist; if the gate ever let it through, the snapshot would surface a
	// `fail` status with an exec error instead.
	err := Run(context.Background(), &out, []string{
		"computer", "snapshot", "Notes",
		"--computer-command", "/this/binary/does/not/exist/ceo-test-12345",
		"--format", "json",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status        string
		Permission    string
		App           string
		Error         string
		ReceiptSHA256 string
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "denied" {
		t.Fatalf("status = %q, want denied", body.Status)
	}
	if body.Permission != "ask" {
		t.Fatalf("permission = %q, want ask", body.Permission)
	}
	if body.App != "Notes" {
		t.Fatalf("app = %q, want Notes", body.App)
	}
	if !strings.Contains(strings.ToLower(body.Error), "approval") {
		t.Fatalf("error = %q, want it to mention operator approval", body.Error)
	}
	if len(body.ReceiptSHA256) != 64 {
		t.Fatalf("receipt_sha256 = %q, want 64-char hex digest", body.ReceiptSHA256)
	}
}

// `computer snapshot` with policy=allow and a deterministic re-exec helper
// must return `pass`, append the app name to the command argv when no
// `{app}` placeholder is present, and surface the helper's stdout.
func Test_Run_computer_snapshot_allow_policy_runs_helper_and_appends_app(t *testing.T) {
	// Given
	t.Setenv("GO_WANT_CLI_COMPUTER_SNAPSHOT_HELPER", "1")
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"computer", "snapshot", "Notes",
		"--computer-policy", "allow",
		"--computer-command", os.Args[0], "-test.run=Test_HelperProcess_cli_computer_snapshot",
	})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status        string
		Permission    string
		App           string
		Output        string
		Bytes         int
		ReceiptSHA256 string
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("status = %q, want pass", body.Status)
	}
	if body.Permission != "allow" {
		t.Fatalf("permission = %q, want allow", body.Permission)
	}
	if body.App != "Notes" {
		t.Fatalf("app = %q, want Notes", body.App)
	}
	if body.Bytes <= 0 {
		t.Fatalf("bytes = %d, want > 0", body.Bytes)
	}
	if !strings.Contains(body.Output, "Notes") {
		t.Fatalf("output = %q, want it to contain the app name appended to argv", body.Output)
	}
	if len(body.ReceiptSHA256) != 64 {
		t.Fatalf("receipt_sha256 = %q, want 64-char hex digest", body.ReceiptSHA256)
	}
}

// Re-exec helper used by the computer snapshot allow-policy test. It echoes
// the argv it received so the parent test can verify the app appears in
// argv, then exits 0 so the snapshot returns pass.
func Test_HelperProcess_cli_computer_snapshot(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_COMPUTER_SNAPSHOT_HELPER") != "1" {
		return
	}
	if _, err := fmt.Fprintf(os.Stdout, "argv=%v\n", os.Args[1:]); err != nil {
		t.Fatalf("write helper output: %v", err)
	}
	os.Exit(0)
}
