package jobpacket

import "testing"

func Test_BuildWithSubagents_preserves_custom_stage(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "ux_reviewer", Role: "review UX", Stage: 3},
	}

	// When
	packet, err := BuildWithSubagents("Review checkout UX", subagents)

	// Then
	if err != nil {
		t.Fatalf("BuildWithSubagents returned error: %v", err)
	}
	if packet.Subagents[0].Stage != 3 {
		t.Fatalf("Stage = %d, want 3", packet.Subagents[0].Stage)
	}
}

func Test_BuildWithSubagents_rejects_invalid_custom_stage(t *testing.T) {
	// Given
	subagents := []Subagent{
		{Name: "ux_reviewer", Role: "review UX", Stage: 4},
	}

	// When
	_, err := BuildWithSubagents("Review checkout UX", subagents)

	// Then
	if err == nil {
		t.Fatal("expected invalid stage error")
	}
}
