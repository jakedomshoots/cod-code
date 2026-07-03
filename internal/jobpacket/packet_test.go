package jobpacket

import (
	"encoding/json"
	"testing"
)

func Test_Build_returns_lean_ceo_packet_when_task_is_valid(t *testing.T) {
	// Given
	task := "Fix a failing test in a small repo"

	// When
	packet, err := Build(task)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if packet.Task != task {
		t.Fatalf("Task = %q, want %q", packet.Task, task)
	}
	if packet.CEO.Authority != "final" {
		t.Fatalf("CEO authority = %q, want final", packet.CEO.Authority)
	}
	if packet.ContextPolicy.Mode != "lean" {
		t.Fatalf("context mode = %q, want lean", packet.ContextPolicy.Mode)
	}
	if packet.ContextPolicy.MaxBytes != 4096 {
		t.Fatalf("context max bytes = %d, want 4096", packet.ContextPolicy.MaxBytes)
	}
	if len(packet.Subagents) != 3 {
		t.Fatalf("subagent count = %d, want 3", len(packet.Subagents))
	}
	if len(packet.Subagents[0].AllowedActions) == 0 {
		t.Fatal("expected subagent allowed actions")
	}
	if len(packet.EvidenceRequired) == 0 {
		t.Fatal("expected evidence requirements")
	}
	if _, err := json.Marshal(packet); err != nil {
		t.Fatalf("packet must marshal to JSON: %v", err)
	}
}

func Test_Build_assigns_role_scoped_allowed_actions(t *testing.T) {
	// Given
	task := "Research auth bug and implement fix"

	// When
	packet, err := Build(task)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	assertSubagentActions(t, packet.Subagents, "coder", []Action{ActionReadWorkspace, ActionSearchWorkspace, ActionProposePatch})
	assertSubagentActions(t, packet.Subagents, "security", []Action{ActionReadWorkspace, ActionSearchWorkspace, ActionRunChecks})
	assertSubagentActions(t, packet.Subagents, "reviewer", []Action{ActionReadWorkspace, ActionRunChecks, ActionVerifyEvidence})
}

func Test_Build_rejects_empty_task_when_task_is_blank(t *testing.T) {
	// Given
	task := "   "

	// When
	_, err := Build(task)

	// Then
	if err == nil {
		t.Fatal("expected error for blank task")
	}
}

func assertSubagentActions(t *testing.T, subagents []Subagent, name string, want []Action) {
	t.Helper()
	for _, subagent := range subagents {
		if subagent.Name != name {
			continue
		}
		if len(subagent.AllowedActions) != len(want) {
			t.Fatalf("%s allowed actions = %#v, want %#v", name, subagent.AllowedActions, want)
		}
		for index, wantAction := range want {
			if subagent.AllowedActions[index] != wantAction {
				t.Fatalf("%s allowed action[%d] = %q, want %q", name, index, subagent.AllowedActions[index], wantAction)
			}
		}
		return
	}
	t.Fatalf("missing subagent %q", name)
}
