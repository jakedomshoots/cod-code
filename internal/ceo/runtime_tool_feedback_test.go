package ceo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
)

func Test_RunJob_browser_read_httptest_localhost_propagates_permission_and_receipt_through_tool_feedback(t *testing.T) {
	// Given — fake subagent issues a browser_read tool request, against an httptest localhost URL.
	expected := "<html>hello localhost</html>"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(expected))
	}))
	defer server.Close()

	runner := browserReadRunner{url: server.URL}
	req := JobRequest{
		Task: "Inspect localhost page",
		Subagents: []jobpacket.Subagent{
			{
				Name:           "browser",
				Role:           "browser",
				Stage:          1,
				AllowedActions: []jobpacket.Action{jobpacket.ActionBrowserRead},
			},
		},
		// BrowserPolicy omitted -> defaults to allow-localhost.
	}

	// When
	result := runtimeExecute(t, runner, req, "browser")

	// Then
	if len(result.ToolResults) != 1 {
		t.Fatalf("ToolResults length = %d, want 1", len(result.ToolResults))
	}
	tr := result.ToolResults[0]
	if tr.Status != "pass" {
		t.Fatalf("Status = %q, want pass: err=%q", tr.Status, tr.Error)
	}
	if tr.Permission != "allow-localhost" {
		t.Fatalf("Permission = %q, want allow-localhost", tr.Permission)
	}
	if tr.URL != server.URL {
		t.Fatalf("URL = %q, want %q", tr.URL, server.URL)
	}
	if tr.ReceiptSHA256 == "" {
		t.Fatal("ReceiptSHA256 is empty after RunJob integration")
	}
	if _, err := hex.DecodeString(tr.ReceiptSHA256); err != nil || len(tr.ReceiptSHA256) != sha256.Size*2 {
		t.Fatalf("ReceiptSHA256 = %q, want hex sha256 digest", tr.ReceiptSHA256)
	}
	if !strings.Contains(tr.Output, "hello localhost") {
		t.Fatalf("Output = %q, want http body content", tr.Output)
	}
	if result.ToolFeedbackPasses != 1 {
		t.Fatalf("ToolFeedbackPasses = %d, want 1", result.ToolFeedbackPasses)
	}
}

func runtimeExecute(t *testing.T, runner SubagentRunner, req JobRequest, agentName string) subagent.Result {
	t.Helper()
	runtime := NewRuntimeWithSubagentRunner(runner)
	report, err := runtime.RunJob(context.Background(), req)
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	for _, result := range report.SubagentResults {
		if result.AgentName == agentName {
			return result
		}
	}
	t.Fatalf("subagent %q not found in %d results", agentName, len(report.SubagentResults))
	return subagent.Result{}
}

type browserReadRunner struct {
	url string
}

func (r browserReadRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	if len(packet.ToolResults) > 0 {
		return subagent.Result{
			AgentName:          packet.AgentName,
			Role:               packet.Role,
			Status:             "pass",
			Attempts:           1,
			ContextReceived:    packet.ContextMode,
			Summary:            "browser read body: " + packet.ToolResults[0].Output,
			ToolFeedbackPasses: 1,
			Evidence:           []string{"browser read feedback"},
		}, nil
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		Summary:         "request browser read",
		ToolRequests: []subagent.ToolRequest{
			{Action: "browser_read", URL: r.url},
		},
		Evidence: []string{"requested browser read"},
	}, nil
}
