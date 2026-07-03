package jobpacket

import "testing"

func Test_BuildWithOptions_uses_configured_context_budget(t *testing.T) {
	// Given
	task := "Fix a failing test"

	// When
	packet, err := BuildWithOptions(BuildOptions{
		Task:            task,
		MaxContextBytes: 1024,
	})
	// Then
	if err != nil {
		t.Fatalf("BuildWithOptions returned error: %v", err)
	}
	if packet.ContextPolicy.MaxBytes != 1024 {
		t.Fatalf("MaxBytes = %d, want configured budget", packet.ContextPolicy.MaxBytes)
	}
}

func Test_BuildWithOptions_rejects_negative_context_budget(t *testing.T) {
	// Given
	req := BuildOptions{Task: "Fix a failing test", MaxContextBytes: -1}

	// When
	_, err := BuildWithOptions(req)

	// Then
	if err == nil {
		t.Fatal("expected error for negative context budget")
	}
}
