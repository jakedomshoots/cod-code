package jobpacket

import (
	"reflect"
	"testing"
)

func Test_IsKnownAction_recognizes_browser_computer_and_manifest_actions(t *testing.T) {
	// Given
	known := []Action{
		ActionReadWorkspace,
		ActionSearchWorkspace,
		ActionNetworkResearch,
		ActionBrowserRead,
		ActionComputerSnapshot,
		ActionToolManifest,
		ActionProposePatch,
		ActionRunChecks,
		ActionVerifyEvidence,
	}

	// When / Then
	for _, action := range known {
		if !IsKnownAction(action) {
			t.Fatalf("IsKnownAction(%q) = false, want true", action)
		}
	}
}

func Test_IsKnownAction_rejects_unknown_and_blank_actions(t *testing.T) {
	// Given
	unknown := []Action{
		"",
		" ",
		"browser_write",
		"computer_click",
		"execute_bash",
		"BROWSER_READ",
	}

	// When / Then
	for _, action := range unknown {
		if IsKnownAction(action) {
			t.Fatalf("IsKnownAction(%q) = true, want false", action)
		}
	}
}

func Test_NormalizeActions_trims_dedupes_and_preserves_browser_compute_and_manifest(t *testing.T) {
	// Given
	input := []Action{
		" browser_read ",
		"browser_read",
		ActionBrowserRead,
		" computer_snapshot",
		"tool_manifest",
	}

	// When
	normalized, ok := NormalizeActions(input)

	// Then
	if !ok {
		t.Fatalf("NormalizeActions(%v) returned ok=false, want true", input)
	}
	want := []Action{ActionBrowserRead, ActionComputerSnapshot, ActionToolManifest}
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("normalized = %v, want %v", normalized, want)
	}
}

func Test_NormalizeActions_returns_empty_and_ok_when_input_is_empty(t *testing.T) {
	// When
	got, ok := NormalizeActions(nil)

	// Then
	if !ok {
		t.Fatalf("NormalizeActions(nil) returned ok=false, want true")
	}
	if len(got) != 0 {
		t.Fatalf("normalized length = %d, want 0", len(got))
	}
}

func Test_NormalizeActions_rejects_unknown_actions(t *testing.T) {
	// Given
	input := []Action{ActionBrowserRead, "launch_missiles"}

	// When
	got, ok := NormalizeActions(input)

	// Then
	if ok {
		t.Fatalf("NormalizeActions(%v) returned ok=true, want false for unknown action", input)
	}
	if got != nil {
		t.Fatalf("normalized = %v, want nil when an unknown action appears", got)
	}
}
