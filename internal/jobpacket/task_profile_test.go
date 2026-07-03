package jobpacket

import "testing"

func Test_Build_adds_task_profile_when_task_mixes_research_and_coding(t *testing.T) {
	// Given
	task := "Research auth bug and implement a fix"

	// When
	packet, err := Build(task)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if packet.TaskProfile.Kind != "mixed" {
		t.Fatalf("Kind = %q, want mixed", packet.TaskProfile.Kind)
	}
	if packet.TaskProfile.RiskLevel != "high" {
		t.Fatalf("RiskLevel = %q, want high", packet.TaskProfile.RiskLevel)
	}
	assertRiskAreas(t, packet.TaskProfile.RiskAreas, []string{"security"})
}

func Test_Build_adds_task_profile_when_task_is_planning(t *testing.T) {
	// Given
	task := "Plan the roadmap for the CLI harness"

	// When
	packet, err := Build(task)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if packet.TaskProfile.Kind != "planning" {
		t.Fatalf("Kind = %q, want planning", packet.TaskProfile.Kind)
	}
	if packet.TaskProfile.RiskLevel != "low" {
		t.Fatalf("RiskLevel = %q, want low", packet.TaskProfile.RiskLevel)
	}
	assertRiskAreas(t, packet.TaskProfile.RiskAreas, nil)
}

func Test_Build_adds_task_profile_risk_areas_when_task_has_multiple_high_risk_domains(t *testing.T) {
	// Given
	task := "Research payment database migration and deploy the fix"

	// When
	packet, err := Build(task)
	// Then
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if packet.TaskProfile.RiskLevel != "high" {
		t.Fatalf("RiskLevel = %q, want high", packet.TaskProfile.RiskLevel)
	}
	assertRiskAreas(t, packet.TaskProfile.RiskAreas, []string{"billing", "database", "release"})
}

func assertRiskAreas(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("RiskAreas = %#v, want %#v", got, want)
	}
	for index, wantArea := range want {
		if got[index] != wantArea {
			t.Fatalf("RiskAreas[%d] = %q, want %q", index, got[index], wantArea)
		}
	}
}
