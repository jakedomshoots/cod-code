package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func Test_Run_prints_saved_job_events_jsonl_when_job_events_flag_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	var runOut bytes.Buffer
	if err := Run(context.Background(), &runOut, []string{"--workspace", root, "Fix", "a", "failing", "test"}); err != nil {
		t.Fatalf("initial Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--job-events", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	events := decodeJobEventLines(t, out.Bytes())
	if len(events) == 0 {
		t.Fatal("job events output is empty")
	}
	if events[0].Kind != "job_packet" || events[0].Status != "ready" {
		t.Fatalf("first event = %+v, want ready job_packet", events[0])
	}
	last := events[len(events)-1]
	if last.Kind != "verdict" || last.Status != "pass" || last.AgentName != "ceo" {
		t.Fatalf("last event = %+v, want passing CEO verdict", last)
	}
}

func Test_Run_prints_saved_job_events_jsonl_when_job_events_latest_alias_is_supplied(t *testing.T) {
	// Given
	root := t.TempDir()
	var runOut bytes.Buffer
	if err := Run(context.Background(), &runOut, []string{"--workspace", root, "Fix", "first", "test"}); err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	if err := Run(context.Background(), &runOut, []string{"--workspace", root, "Fix", "latest", "test"}); err != nil {
		t.Fatalf("latest Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--job-events", "latest"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	events := decodeJobEventLines(t, out.Bytes())
	if len(events) == 0 {
		t.Fatal("job events output is empty")
	}
	last := events[len(events)-1]
	if last.Kind != "verdict" || last.Status != "pass" || last.AgentName != "ceo" {
		t.Fatalf("last event = %+v, want passing latest CEO verdict", last)
	}
}

func decodeJobEventLines(t *testing.T, payload []byte) []struct {
	Index     int    `json:"index"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	AgentName string `json:"agent_name"`
} {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(payload))
	events := []struct {
		Index     int    `json:"index"`
		Kind      string `json:"kind"`
		Status    string `json:"status"`
		AgentName string `json:"agent_name"`
	}{}
	for {
		var event struct {
			Index     int    `json:"index"`
			Kind      string `json:"kind"`
			Status    string `json:"status"`
			AgentName string `json:"agent_name"`
		}
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("job events output must be JSONL: %v\n%s", err, string(payload))
		}
		events = append(events, event)
	}
	return events
}
