package jobpacket

import "testing"

func Test_BuildWithSubagents_uses_custom_subagents_when_supplied(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "planner", Role: "break down work"},
		{Name: "security", Role: "review auth risks"},
	}

	// When
	packet, err := BuildWithSubagents("Fix auth flow", subagents)

	// Then
	if err != nil {
		t.Fatalf("BuildWithSubagents returned error: %v", err)
	}
	if packet.MaxSubagents != 2 {
		t.Fatalf("MaxSubagents = %d, want 2", packet.MaxSubagents)
	}
	if packet.Subagents[0].Name != "planner" || packet.Subagents[1].Role != "review auth risks" {
		t.Fatalf("Subagents = %#v, want custom delegation", packet.Subagents)
	}
	if len(packet.Subagents[0].AllowedActions) == 0 {
		t.Fatal("expected default allowed actions on custom subagent")
	}
}

func Test_BuildWithSubagents_uses_custom_allowed_actions_when_supplied(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "coder", Role: "read only coding review", AllowedActions: []Action{ActionReadWorkspace}},
	}

	// When
	packet, err := BuildWithSubagents("Fix auth flow", subagents)

	// Then
	if err != nil {
		t.Fatalf("BuildWithSubagents returned error: %v", err)
	}
	if len(packet.Subagents[0].AllowedActions) != 1 || packet.Subagents[0].AllowedActions[0] != ActionReadWorkspace {
		t.Fatalf("AllowedActions = %#v, want read-only override", packet.Subagents[0].AllowedActions)
	}
}

func Test_BuildWithSubagents_rejects_invalid_allowed_action(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "coder", Role: "apply bounded changes", AllowedActions: []Action{"delete_workspace"}},
	}

	// When
	_, err := BuildWithSubagents("Fix auth flow", subagents)

	// Then
	if err == nil {
		t.Fatal("expected invalid allowed action error")
	}
}

func Test_BuildWithSubagents_rejects_duplicate_subagent_names(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "planner", Role: "plan"},
		{Name: "planner", Role: "review"},
	}

	// When
	_, err := BuildWithSubagents("Fix auth flow", subagents)

	// Then
	if err == nil {
		t.Fatal("expected duplicate subagent error")
	}
}
