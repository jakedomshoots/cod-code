package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
)

func Test_Run_prints_doctor_event_details_when_events_format_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	script := filepath.Join("..", "..", "examples", "research-command.sh")

	// When
	err := Run(context.Background(), &out, []string{"--doctor", "--format", "events", "--research-command", "sh", script})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	type eventLine struct {
		Kind       string `json:"kind"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Source     string `json:"source"`
		Verdict    string `json:"verdict"`
		PatchCount int    `json:"patch_count"`
		CheckCount int    `json:"check_count"`
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
			t.Fatalf("doctor events output must be JSONL: %v\n%s", decodeErr, out.String())
		}
		events = append(events, event)
	}
	var goldenSeen bool
	var researchSeen bool
	for _, event := range events {
		if event.Name == "golden_demo" && event.PatchCount == 1 && event.CheckCount == 1 {
			goldenSeen = true
		}
		if event.Name == "research_command" && event.Source == "flag" && event.Verdict == "pass" {
			researchSeen = true
		}
	}
	if !goldenSeen || !researchSeen {
		t.Fatalf("events = %#v, want golden and research doctor events", events)
	}
}
