package jobpacket

import "testing"

func Test_BuildWithSubagents_preserves_custom_subagent_context_budget(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "scanner", Role: "inspect scope", MaxContextBytes: 512},
	}

	// When
	packet, err := BuildWithSubagents("Fix a failing test", subagents)

	// Then
	if err != nil {
		t.Fatalf("BuildWithSubagents returned error: %v", err)
	}
	if packet.Subagents[0].MaxContextBytes != 512 {
		t.Fatalf("MaxContextBytes = %d, want 512", packet.Subagents[0].MaxContextBytes)
	}
}

func Test_BuildWithSubagents_rejects_negative_subagent_context_budget(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "scanner", Role: "inspect scope", MaxContextBytes: -1},
	}

	// When
	_, err := BuildWithSubagents("Fix a failing test", subagents)

	// Then
	if err == nil {
		t.Fatal("expected negative subagent context budget error")
	}
}
