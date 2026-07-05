package ceo

import (
	"context"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/config"
	"ceoharness/internal/history"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

type Report struct {
	SchemaVersion          int                            `json:"schema_version"`
	JobPacket              jobpacket.Packet               `json:"job_packet"`
	JobOwner               string                         `json:"job_owner"`
	LifecycleState         LifecycleState                 `json:"lifecycle_state"`
	LifecycleEvents        []LifecycleEvent               `json:"lifecycle_events"`
	VerificationContract   VerificationContract           `json:"verification_contract"`
	ProviderRouteDecisions []config.ProviderRouteDecision `json:"provider_route_decisions,omitempty"`
	ContextTrace           []ContextTraceEntry            `json:"context_trace,omitempty"`
	RunLedger              RunLedger                      `json:"run_ledger"`
	RunManifest            RunManifest                    `json:"run_manifest"`
	RunEvents              []RunEvent                     `json:"run_events"`
	Continuation           *ContinuationContext           `json:"continuation,omitempty"`
	WorkspaceBrief         *workspace.Brief               `json:"workspace_brief,omitempty"`
	Resume                 *ResumeContext                 `json:"resume,omitempty"`
	SubagentResults        []subagent.Result              `json:"subagent_results"`
	ChangedFiles           []string                       `json:"changed_files"`
	CheckResults           []checkrunner.Result           `json:"check_results"`
	VerificationSummary    VerificationSummary            `json:"verification_summary"`
	ExecutionPlan          ExecutionPlan                  `json:"execution_plan"`
	PatchResults           []workspace.ReplaceTextResult  `json:"patch_results"`
	PatchPreviews          []workspace.ReplaceTextResult  `json:"patch_previews"`
	PatchAudit             []PatchAuditEntry              `json:"patch_audit"`
	PatchApproval          *PatchApproval                 `json:"patch_approval,omitempty"`
	CEODelegation          *CEODelegation                 `json:"ceo_delegation,omitempty"`
	CEOReview              *CEOReview                     `json:"ceo_review,omitempty"`
	RetryHistory           []RetryHistoryEntry            `json:"retry_history,omitempty"`
	HistoryPath            string                         `json:"history_path,omitempty"`
	JobID                  string                         `json:"job_id,omitempty"`
	Verdict                string                         `json:"verdict"`
}

type CEODelegation struct {
	Source            string               `json:"source"`
	ModelSource       string               `json:"model_source,omitempty"`
	ProviderName      string               `json:"provider_name,omitempty"`
	SelectedSubagents []string             `json:"selected_subagents"`
	NewSubagents      []jobpacket.Subagent `json:"new_subagents,omitempty"`
	Assignments       map[string]string    `json:"assignments,omitempty"`
	Summary           string               `json:"summary"`
	PromptBytes       int                  `json:"prompt_bytes"`
}

type CEOReview struct {
	Source             string `json:"source"`
	ModelSource        string `json:"model_source,omitempty"`
	ProviderName       string `json:"provider_name,omitempty"`
	RecommendedVerdict string `json:"recommended_verdict"`
	Summary            string `json:"summary"`
	PromptBytes        int    `json:"prompt_bytes"`
}

type ResumeContext struct {
	JobID     string   `json:"job_id"`
	Questions []string `json:"questions,omitempty"`
	Answers   []string `json:"answers,omitempty"`
}

type ContinuationContext struct {
	JobID               string            `json:"job_id"`
	ReusedSubagentCount int               `json:"reused_subagent_count,omitempty"`
	ReusableResults     []subagent.Result `json:"-"`
	UseSavedDelegation  bool              `json:"-"`
	SavedDelegation     *CEODelegation    `json:"-"`
}

type PatchRequest struct {
	Path    string
	Old     string
	New     string
	Content string
}

type JobRequest struct {
	Task                            string
	WorkspaceDir                    string
	ArtifactRoot                    string
	CheckCommand                    []string
	CheckCommands                   [][]string
	CheckEnv                        []string
	ResearchCommand                 []string
	ToolCommandTimeoutMS            int
	BrowserPolicy                   string
	BrowserCommand                  []string
	ComputerPolicy                  string
	ComputerCommand                 []string
	CheckAttempts                   int
	CheckBackoffMS                  int
	CheckFixAttempts                int
	CEORevisionAttempts             int
	MaxCEOIterations                int
	MaxSubagents                    int
	SubagentConcurrency             int
	MaxToolRequests                 int
	Subagents                       []jobpacket.Subagent
	SubagentAttempts                int
	SubagentBackoffMS               int
	NoProgressStop                  int
	ProviderCostBudgetMicroUSD      int64
	ProviderHealthPolicy            history.ProviderHealthPolicy
	ProviderHealthAvoidedRouteCount int
	ProviderHealthAvoidedProviders  []string
	ProviderRouteDecisions          []config.ProviderRouteDecision
	MaxContextBytes                 int
	MaxSubagentOutputBytes          int
	Continuation                    *ContinuationContext
	WorkspaceBriefMaxFiles          int
	WorkspaceBriefExcludes          []string
	Resume                          *ResumeContext
	Patches                         []PatchRequest
	ScorerFailedChecks              []RepairFailureDetail
	ApprovedPreviewDigest           string
	DryRun                          bool
	ApplyModelPatches               bool
	PreviewModelPatches             bool
	MaxModelPatches                 int
}

type SubagentRunner interface {
	Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error)
}

type Runtime struct {
	runner           SubagentRunner
	checks           checkrunner.Runner
	ceoReviewer      model.Client
	ceoReviewerRoute subagent.RouteMetadata
}
