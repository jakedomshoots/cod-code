package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_filters_provider_health_rollup_when_provider_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "first",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, PassCount: 1, EstimatedCostMicroUSD: 106},
				{ProviderName: "cheap", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1, UnauthorizedCount: 1},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--provider", "fast"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Provider       string `json:"provider,omitempty"`
		ProviderHealth []struct {
			ProviderName          string `json:"provider_name"`
			AttemptCount          int    `json:"attempt_count"`
			EstimatedCostMicroUSD int64  `json:"estimated_cost_microusd"`
		} `json:"provider_health"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Provider != "fast" {
		t.Fatalf("Provider = %q, want fast", body.Provider)
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	health := body.ProviderHealth[0]
	if health.ProviderName != "fast" || health.AttemptCount != 1 || health.EstimatedCostMicroUSD != 106 {
		t.Fatalf("provider health = %#v, want fast only", health)
	}
}

func Test_Run_filters_provider_health_rollup_when_recommendation_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "first",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, PassCount: 1},
				{ProviderName: "cheap", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1, UnauthorizedCount: 1},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--recommendation", "avoid"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Recommendation string `json:"recommendation,omitempty"`
		ProviderHealth []struct {
			ProviderName   string `json:"provider_name"`
			Recommendation string `json:"recommendation"`
		} `json:"provider_health"`
		Summary struct {
			AvoidCount   int `json:"avoid_count"`
			HealthyCount int `json:"healthy_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Recommendation != "avoid" {
		t.Fatalf("Recommendation = %q, want avoid", body.Recommendation)
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	health := body.ProviderHealth[0]
	if health.ProviderName != "cheap" || health.Recommendation != "avoid" {
		t.Fatalf("provider health = %#v, want cheap avoid", health)
	}
	if body.Summary.AvoidCount != 1 || body.Summary.HealthyCount != 0 {
		t.Fatalf("summary = %#v, want 1 avoid and 0 healthy", body.Summary)
	}
}

func Test_Run_filters_provider_health_rollup_when_task_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	entries := []history.Entry{
		{
			Task: "Fix checkout retry",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, PassCount: 1},
			},
		},
		{
			Task: "Refactor parser",
			ProviderHealth: []history.ProviderHealth{
				{ProviderName: "fast", ModelSource: "http", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
			},
		},
	}
	for _, entry := range entries {
		if _, err := store.Append(context.Background(), entry); err != nil {
			t.Fatalf("Append returned error: %v", err)
		}
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--provider-health", "--task", "checkout"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		TaskFilter     string `json:"task_filter,omitempty"`
		ProviderHealth []struct {
			ProviderName   string `json:"provider_name"`
			AttemptCount   int    `json:"attempt_count"`
			Recommendation string `json:"recommendation"`
		} `json:"provider_health"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.TaskFilter != "checkout" {
		t.Fatalf("TaskFilter = %q, want checkout", body.TaskFilter)
	}
	if len(body.ProviderHealth) != 1 {
		t.Fatalf("provider health length = %d, want 1: %#v", len(body.ProviderHealth), body.ProviderHealth)
	}
	health := body.ProviderHealth[0]
	if health.ProviderName != "fast" || health.AttemptCount != 1 || health.Recommendation != "healthy" {
		t.Fatalf("provider health = %#v, want filtered healthy fast row", health)
	}
}
