package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func Test_Run_prints_execution_plan_when_task_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"Fix", "a", "failing", "test"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		ExecutionPlan struct {
			Authority string `json:"authority"`
			Mode      string `json:"mode"`
			Steps     []struct {
				Owner  string `json:"owner"`
				Status string `json:"status"`
			} `json:"steps"`
			NextAction string `json:"next_action"`
		} `json:"execution_plan"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.ExecutionPlan.Authority != "ceo" || body.ExecutionPlan.Mode != "delegated" {
		t.Fatalf("execution plan = %#v, want delegated CEO plan", body.ExecutionPlan)
	}
	if len(body.ExecutionPlan.Steps) != 4 {
		t.Fatalf("steps length = %d, want 4", len(body.ExecutionPlan.Steps))
	}
	if body.ExecutionPlan.Steps[3].Owner != "ceo" || body.ExecutionPlan.NextAction != "accept" {
		t.Fatalf("final plan = %#v, want CEO accept", body.ExecutionPlan)
	}
}
