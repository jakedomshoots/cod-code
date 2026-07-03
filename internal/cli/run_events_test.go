package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func Test_Run_prints_compact_run_events(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		RunEvents []struct {
			Index int    `json:"index"`
			Kind  string `json:"kind"`
		} `json:"run_events"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if len(body.RunEvents) == 0 {
		t.Fatal("run_events is empty")
	}
	if body.RunEvents[0].Index != 1 || body.RunEvents[0].Kind != "job_packet" {
		t.Fatalf("first run event = %+v, want indexed job_packet", body.RunEvents[0])
	}
	if body.RunEvents[len(body.RunEvents)-1].Kind != "verdict" {
		t.Fatalf("last run event = %+v, want verdict", body.RunEvents[len(body.RunEvents)-1])
	}
}

func Test_Run_prints_events_jsonl_when_events_format_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{"--format", "events", "Fix", "a", "failing", "test"}

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	type eventLine struct {
		Index     int             `json:"index"`
		Kind      string          `json:"kind"`
		Status    string          `json:"status"`
		AgentName string          `json:"agent_name"`
		JobPacket json.RawMessage `json:"job_packet"`
	}
	decoder := json.NewDecoder(bytes.NewReader(out.Bytes()))
	events := []eventLine{}
	for {
		var event eventLine
		decodeErr := decoder.Decode(&event)
		if decodeErr == io.EOF {
			break
		}
		if decodeErr != nil {
			t.Fatalf("events output must be JSONL: %v\n%s", decodeErr, out.String())
		}
		if len(event.JobPacket) > 0 {
			t.Fatalf("events output included full report job_packet field:\n%s", out.String())
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		t.Fatal("events output is empty")
	}
	if events[0].Index != 1 || events[0].Kind != "job_packet" || events[0].Status != "ready" {
		t.Fatalf("first event = %+v, want ready job_packet", events[0])
	}
	last := events[len(events)-1]
	if last.Kind != "verdict" || last.Status != "pass" || last.AgentName != "ceo" {
		t.Fatalf("last event = %+v, want passing CEO verdict", last)
	}
}
