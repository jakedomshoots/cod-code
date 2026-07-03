package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"ceoharness/internal/history"
)

func Test_Run_prints_plan_only_preview_without_running_models(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	modelLogPath := filepath.Join(root, "model-ran.log")
	cheapCommand := "cat >/dev/null; printf cheap-route >> " + strconv.Quote(modelLogPath) + "; printf cheap-route"
	premiumCommand := "cat >/dev/null; printf premium-route >> " + strconv.Quote(modelLogPath) + "; printf premium-route"
	configJSON := `{"providers":{"cheap":{"model_command":["sh","-c",` + strconv.Quote(cheapCommand) + `]},"premium":{"model_command":["sh","-c",` + strconv.Quote(premiumCommand) + `]}},"provider_policy":{"default_provider":"cheap","fallback_provider":"premium"},"check_command":["go","test","./..."],"max_context_bytes":2048,"max_tool_requests":2}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := history.New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	_, err = store.Append(context.Background(), history.Entry{
		Task:    "bad cheap run",
		Verdict: "fail",
		ProviderHealth: []history.ProviderHealth{
			{ProviderName: "cheap", ModelSource: "command", AttemptCount: 1, FailCount: 1, ErrorCount: 1},
		},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--plan-only", "Plan", "roadmap"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Mode      string `json:"mode"`
		JobOwner  string `json:"job_owner"`
		JobPacket struct {
			TaskProfile struct {
				Kind string `json:"kind"`
			} `json:"task_profile"`
			ContextPolicy struct {
				MaxBytes int `json:"max_bytes"`
			} `json:"context_policy"`
			Subagents []struct {
				Name string `json:"name"`
			} `json:"subagents"`
		} `json:"job_packet"`
		ProviderRoutes                  map[string]string `json:"provider_routes"`
		ProviderHealthAvoidedRouteCount int               `json:"provider_health_avoided_route_count"`
		ProviderHealthAvoidedProviders  []string          `json:"provider_health_avoided_providers"`
		CheckCommandCount               int               `json:"check_command_count"`
		VerificationContract            struct {
			Status             string   `json:"status"`
			RequiredCheckCount int      `json:"required_check_count"`
			RequiredChecks     []string `json:"required_checks"`
		} `json:"verification_contract"`
		MaxToolRequests int `json:"max_tool_requests"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Mode != "plan_only" {
		t.Fatalf("Mode = %q, want plan_only", body.Mode)
	}
	if body.JobOwner != "planner" {
		t.Fatalf("JobOwner = %q, want planner", body.JobOwner)
	}
	if body.JobPacket.TaskProfile.Kind != "planning" {
		t.Fatalf("task kind = %q, want planning", body.JobPacket.TaskProfile.Kind)
	}
	if len(body.JobPacket.Subagents) != 2 || body.JobPacket.Subagents[0].Name != "planner" || body.JobPacket.Subagents[1].Name != "reviewer" {
		t.Fatalf("subagents = %#v, want planner/reviewer", body.JobPacket.Subagents)
	}
	if body.ProviderRoutes["planner"] != "premium" || body.ProviderRoutes["reviewer"] != "premium" {
		t.Fatalf("ProviderRoutes = %#v, want planner/reviewer premium", body.ProviderRoutes)
	}
	if body.ProviderHealthAvoidedRouteCount != 2 {
		t.Fatalf("ProviderHealthAvoidedRouteCount = %d, want 2", body.ProviderHealthAvoidedRouteCount)
	}
	if len(body.ProviderHealthAvoidedProviders) != 1 || body.ProviderHealthAvoidedProviders[0] != "cheap" {
		t.Fatalf("ProviderHealthAvoidedProviders = %#v, want [cheap]", body.ProviderHealthAvoidedProviders)
	}
	if body.CheckCommandCount != 1 {
		t.Fatalf("CheckCommandCount = %d, want 1", body.CheckCommandCount)
	}
	if body.VerificationContract.Status != "pending" || body.VerificationContract.RequiredCheckCount != 1 {
		t.Fatalf("VerificationContract = %#v, want one pending check", body.VerificationContract)
	}
	if len(body.VerificationContract.RequiredChecks) != 1 || body.VerificationContract.RequiredChecks[0] != "go test ./..." {
		t.Fatalf("RequiredChecks = %#v, want go test ./...", body.VerificationContract.RequiredChecks)
	}
	if body.JobPacket.ContextPolicy.MaxBytes != 2048 || body.MaxToolRequests != 2 {
		t.Fatalf("limits = context %d tools %d, want 2048 and 2", body.JobPacket.ContextPolicy.MaxBytes, body.MaxToolRequests)
	}
	if _, err := os.Stat(modelLogPath); err == nil {
		t.Fatal("model command ran during plan-only preview")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat model log: %v", err)
	}
}

func Test_Run_plan_only_reports_ceo_provider_from_workspace_config(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	configJSON := `{"providers":{"main":{"model_command":["echo","review"]}},"ceo_provider":"main"}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--plan-only", "Fix", "bug"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		CEOProvider        string `json:"ceo_provider"`
		CEOProviderPresent bool   `json:"ceo_provider_present"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.CEOProvider != "main" || !body.CEOProviderPresent {
		t.Fatalf("CEO provider = %q present=%v, want main true", body.CEOProvider, body.CEOProviderPresent)
	}
}

func Test_Run_plan_only_uses_max_subagents_flag_when_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--plan-only", "--max-subagents", "7", "Research", "payment", "database", "migration", "and", "deploy", "the", "fix"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		JobPacket struct {
			MaxSubagents int `json:"max_subagents"`
			Subagents    []struct {
				Name string `json:"name"`
			} `json:"subagents"`
		} `json:"job_packet"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.JobPacket.MaxSubagents != 7 || len(body.JobPacket.Subagents) != 7 {
		t.Fatalf("job packet = %#v, want seven delegated subagents", body.JobPacket)
	}
	if body.JobPacket.Subagents[1].Name != "researcher" || body.JobPacket.Subagents[4].Name != "database" {
		t.Fatalf("subagents = %#v, want full mixed-risk crew", body.JobPacket.Subagents)
	}
}

func Test_Run_plan_only_previews_continue_job_reuse(t *testing.T) {
	// Given
	root := t.TempDir()
	if err := Run(context.Background(), &bytes.Buffer{}, []string{"--workspace", root, "Fix", "auth", "bug"}); err != nil {
		t.Fatalf("seed Run returned error: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--plan-only", "--continue-job", "job-000001"})

	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Continuation struct {
			JobID                  string `json:"job_id"`
			UseSavedDelegation     bool   `json:"use_saved_delegation"`
			PlannedSubagentCount   int    `json:"planned_subagent_count"`
			ReusableSubagentCount  int    `json:"reusable_subagent_count"`
			SavedDelegationPresent bool   `json:"saved_delegation_present"`
		} `json:"continuation"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, out.String())
	}
	if body.Continuation.JobID != "job-000001" {
		t.Fatalf("Continuation job = %q, want job-000001", body.Continuation.JobID)
	}
	if !body.Continuation.UseSavedDelegation {
		t.Fatal("UseSavedDelegation = false, want true")
	}
	if body.Continuation.PlannedSubagentCount != 3 || body.Continuation.ReusableSubagentCount != 3 {
		t.Fatalf("Continuation counts = %+v, want three planned and reusable subagents", body.Continuation)
	}
	if body.Continuation.SavedDelegationPresent {
		t.Fatal("SavedDelegationPresent = true, want false for a default static seed run")
	}
}
