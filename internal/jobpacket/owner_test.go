package jobpacket

import "testing"

func Test_OwnerForPacket_uses_task_kind_primary_agent(t *testing.T) {
	tests := []struct {
		name  string
		task  string
		owner string
	}{
		{name: "coding", task: "Fix a failing test", owner: "coder"},
		{name: "planning", task: "Plan roadmap", owner: "planner"},
		{name: "research", task: "Research agent harness docs", owner: "researcher"},
		{name: "mixed", task: "Research and implement API docs", owner: "coder"},
		{name: "database risk", task: "Implement database migration fix", owner: "coder"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			packet, err := Build(test.task)
			if err != nil {
				t.Fatalf("Build returned error: %v", err)
			}

			// When
			owner := OwnerForPacket(packet)

			// Then
			if owner != test.owner {
				t.Fatalf("owner = %q, want %q", owner, test.owner)
			}
		})
	}
}
