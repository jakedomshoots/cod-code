package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func Test_Run_includes_lifecycle_fields_when_task_passes(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--check", "go", "version", "--", "Fix", "a", "failing", "test"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	var body struct {
		LifecycleState  string `json:"lifecycle_state"`
		LifecycleEvents []struct {
			State string `json:"state"`
		} `json:"lifecycle_events"`
		RunEvents []struct {
			Kind           string `json:"kind"`
			LifecycleState string `json:"lifecycle_state"`
		} `json:"run_events"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.LifecycleState != "passed" {
		t.Fatalf("LifecycleState = %q, want passed", body.LifecycleState)
	}
	if !jsonLifecycleStatesInclude(body.LifecycleEvents, "checking", "reviewing", "passed") {
		t.Fatalf("LifecycleEvents = %+v, want checking/reviewing/passed", body.LifecycleEvents)
	}
	if !jsonRunEventLifecycleIncludes(body.RunEvents, "check", "checking") {
		t.Fatalf("RunEvents = %+v, want check lifecycle checking", body.RunEvents)
	}
	if !jsonRunEventLifecycleIncludes(body.RunEvents, "verdict", "passed") {
		t.Fatalf("RunEvents = %+v, want verdict lifecycle passed", body.RunEvents)
	}
}

func Test_Run_events_format_prints_lifecycle_state_on_each_event(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--format", "events", "--check", "go", "version", "--", "Fix", "a", "failing", "test"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	events := decodeLifecycleEventLines(t, out.Bytes())
	if !jsonRunEventLifecycleIncludes(events, "check", "checking") {
		t.Fatalf("events = %+v, want check lifecycle checking", events)
	}
	if !jsonRunEventLifecycleIncludes(events, "verdict", "passed") {
		t.Fatalf("events = %+v, want verdict lifecycle passed", events)
	}
}

func decodeLifecycleEventLines(t *testing.T, payload []byte) []struct {
	Kind           string `json:"kind"`
	LifecycleState string `json:"lifecycle_state"`
} {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(payload))
	events := []struct {
		Kind           string `json:"kind"`
		LifecycleState string `json:"lifecycle_state"`
	}{}
	for {
		var event struct {
			Kind           string `json:"kind"`
			LifecycleState string `json:"lifecycle_state"`
		}
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("events output must be JSONL: %v\n%s", err, string(payload))
		}
		events = append(events, event)
	}
	return events
}

func jsonLifecycleStatesInclude(events []struct {
	State string `json:"state"`
}, want ...string) bool {
	next := 0
	for _, event := range events {
		if next < len(want) && event.State == want[next] {
			next++
		}
	}
	return next == len(want)
}

func jsonRunEventLifecycleIncludes(events []struct {
	Kind           string `json:"kind"`
	LifecycleState string `json:"lifecycle_state"`
}, kind string, state string) bool {
	for _, event := range events {
		if event.Kind == kind && event.LifecycleState == state {
			return true
		}
	}
	return false
}
