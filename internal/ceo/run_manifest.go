package ceo

import "ceoharness/internal/jobpacket"

type RunManifest struct {
	SchemaVersion                   int      `json:"schema_version"`
	TaskBytes                       int      `json:"task_bytes"`
	TaskKind                        string   `json:"task_kind"`
	RiskLevel                       string   `json:"risk_level"`
	RiskAreas                       []string `json:"risk_areas,omitempty"`
	ContextMode                     string   `json:"context_mode"`
	MaxContextBytes                 int      `json:"max_context_bytes"`
	SubagentConcurrency             int      `json:"subagent_concurrency,omitempty"`
	MaxToolRequests                 int      `json:"max_tool_requests,omitempty"`
	MaxOutputBytes                  int      `json:"max_subagent_output_bytes,omitempty"`
	NoProgressStop                  int      `json:"no_progress_stop,omitempty"`
	MaxCEOIterations                int      `json:"max_ceo_iterations,omitempty"`
	CEOIterationCount               int      `json:"ceo_iteration_count,omitempty"`
	CEOIterationExhausted           bool     `json:"ceo_iteration_exhausted,omitempty"`
	DryRun                          bool     `json:"dry_run,omitempty"`
	SubagentCount                   int      `json:"subagent_count"`
	ReusedSubagentCount             int      `json:"reused_subagent_count,omitempty"`
	CheckAttemptCount               int      `json:"check_attempt_count"`
	ChangedFileCount                int      `json:"changed_file_count"`
	PatchCount                      int      `json:"patch_count"`
	ProviderHealthAvoidedRouteCount int      `json:"provider_health_avoided_route_count,omitempty"`
	ProviderHealthAvoidedProviders  []string `json:"provider_health_avoided_providers,omitempty"`
	Verdict                         string   `json:"verdict"`
}

type runManifestInput struct {
	Packet                          jobpacket.Packet
	SubagentCount                   int
	ReusedSubagentCount             int
	ChangedFileCount                int
	CheckAttemptCount               int
	PatchCount                      int
	SubagentConcurrency             int
	MaxToolRequests                 int
	MaxOutputBytes                  int
	NoProgressStop                  int
	MaxCEOIterations                int
	CEOIterationCount               int
	CEOIterationExhausted           bool
	DryRun                          bool
	ProviderHealthAvoidedRouteCount int
	ProviderHealthAvoidedProviders  []string
	Verdict                         string
}

func buildRunManifest(input runManifestInput) RunManifest {
	return RunManifest{
		SchemaVersion:                   1,
		TaskBytes:                       len(input.Packet.Task),
		TaskKind:                        input.Packet.TaskProfile.Kind,
		RiskLevel:                       input.Packet.TaskProfile.RiskLevel,
		RiskAreas:                       append([]string(nil), input.Packet.TaskProfile.RiskAreas...),
		ContextMode:                     input.Packet.ContextPolicy.Mode,
		MaxContextBytes:                 input.Packet.ContextPolicy.MaxBytes,
		SubagentConcurrency:             input.SubagentConcurrency,
		MaxToolRequests:                 input.MaxToolRequests,
		MaxOutputBytes:                  input.MaxOutputBytes,
		NoProgressStop:                  input.NoProgressStop,
		MaxCEOIterations:                input.MaxCEOIterations,
		CEOIterationCount:               input.CEOIterationCount,
		CEOIterationExhausted:           input.CEOIterationExhausted,
		DryRun:                          input.DryRun,
		ProviderHealthAvoidedRouteCount: input.ProviderHealthAvoidedRouteCount,
		ProviderHealthAvoidedProviders:  append([]string(nil), input.ProviderHealthAvoidedProviders...),
		SubagentCount:                   input.SubagentCount,
		ReusedSubagentCount:             input.ReusedSubagentCount,
		CheckAttemptCount:               input.CheckAttemptCount,
		ChangedFileCount:                input.ChangedFileCount,
		PatchCount:                      input.PatchCount,
		Verdict:                         input.Verdict,
	}
}
