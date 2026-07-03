package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_prints_subagent_stages_when_task_uses_native_roles(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"Fix", "a", "failing", "test"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		SubagentResults []struct {
			AgentName string `json:"agent_name"`
			Stage     int    `json:"stage"`
		} `json:"subagent_results"`
		ExecutionPlan struct {
			Steps []struct {
				Owner string `json:"owner"`
				Stage int    `json:"stage"`
			} `json:"steps"`
		} `json:"execution_plan"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if len(body.SubagentResults) != 3 {
		t.Fatalf("subagent count = %d, want 3", len(body.SubagentResults))
	}
	for index, wantStage := range []int{1, 2, 3} {
		if body.SubagentResults[index].Stage != wantStage {
			t.Fatalf("subagent stage[%d] = %d, want %d", index, body.SubagentResults[index].Stage, wantStage)
		}
		if body.ExecutionPlan.Steps[index].Stage != wantStage {
			t.Fatalf("plan stage[%d] = %d, want %d", index, body.ExecutionPlan.Steps[index].Stage, wantStage)
		}
	}
}
