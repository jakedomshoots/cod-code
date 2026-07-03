package subagent

import (
	"encoding/json"
	"strings"
	"testing"
)

func Test_RenderToolResults_compacts_large_output_for_feedback_prompts(t *testing.T) {
	// Given
	result := ToolResult{
		Action: "run_checks",
		Status: "fail",
		Output: strings.Repeat("o", 1600),
		Error:  strings.Repeat("e", 1700),
	}

	// When
	rendered := RenderToolResults([]ToolResult{result})

	// Then
	var body toolResultEnvelope
	if err := json.Unmarshal([]byte(rendered), &body); err != nil {
		t.Fatalf("rendered tool results must be JSON: %v\n%s", err, rendered)
	}
	if len(body.ToolResults) != 1 {
		t.Fatalf("ToolResults length = %d, want 1", len(body.ToolResults))
	}
	got := body.ToolResults[0]
	if len(got.Output) > maxRenderedToolTextBytes+20 {
		t.Fatalf("Output length = %d, want compact output", len(got.Output))
	}
	if len(got.Error) > maxRenderedToolTextBytes+20 {
		t.Fatalf("Error length = %d, want compact error", len(got.Error))
	}
	if !got.Truncated {
		t.Fatal("expected rendered tool result to mark truncation")
	}
}
