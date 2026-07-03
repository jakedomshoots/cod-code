package jobpacket

import "testing"

func Test_BuildWithSubagents_preserves_custom_subagent_provider(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "ux_reviewer", Role: "review UX", ProviderName: "premium"},
	}

	// When
	packet, err := BuildWithSubagents("Review checkout UX", subagents)
	// Then
	if err != nil {
		t.Fatalf("BuildWithSubagents returned error: %v", err)
	}
	if packet.Subagents[0].ProviderName != "premium" {
		t.Fatalf("ProviderName = %q, want premium", packet.Subagents[0].ProviderName)
	}
}
