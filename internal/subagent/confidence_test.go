package subagent

import "testing"

func Test_ParseModelOutput_reads_confidence_when_json_includes_confidence(t *testing.T) {
	// Given
	payload := `{"summary":"ready","confidence":0.42}`

	// When
	output, err := ParseModelOutput(payload)

	// Then
	if err != nil {
		t.Fatalf("ParseModelOutput returned error: %v", err)
	}
	if output.Confidence == nil || *output.Confidence != 0.42 {
		t.Fatalf("Confidence = %v, want 0.42", output.Confidence)
	}
}

func Test_ParseModelOutput_rejects_confidence_outside_unit_range(t *testing.T) {
	// Given
	payload := `{"summary":"not valid","confidence":1.2}`

	// When
	_, err := ParseModelOutput(payload)

	// Then
	if err == nil {
		t.Fatal("expected invalid confidence error")
	}
}
