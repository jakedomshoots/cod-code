package ceo

import (
	"fmt"
	"strings"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

const maxRunEventSummaryBytes = 180

type RunEvent struct {
	Index                  int            `json:"index"`
	Kind                   string         `json:"kind"`
	Status                 string         `json:"status"`
	LifecycleState         LifecycleState `json:"lifecycle_state,omitempty"`
	AgentName              string         `json:"agent_name,omitempty"`
	Stage                  int            `json:"stage,omitempty"`
	RouteCount             int            `json:"route_count,omitempty"`
	ProviderNames          []string       `json:"provider_names,omitempty"`
	ProviderName           string         `json:"provider_name,omitempty"`
	ProviderFallbackFrom   string         `json:"provider_fallback_from,omitempty"`
	ProviderFallbackReason string         `json:"provider_fallback_reason,omitempty"`
	Action                 string         `json:"action,omitempty"`
	Path                   string         `json:"path,omitempty"`
	Query                  string         `json:"query,omitempty"`
	URL                    string         `json:"url,omitempty"`
	App                    string         `json:"app,omitempty"`
	Tool                   string         `json:"tool,omitempty"`
	Permission             string         `json:"permission,omitempty"`
	CheckIndex             int            `json:"check_index,omitempty"`
	Attempt                int            `json:"attempt,omitempty"`
	Source                 string         `json:"source,omitempty"`
	Digest                 string         `json:"digest,omitempty"`
	PreviewCount           int            `json:"preview_count,omitempty"`
	Summary                string         `json:"summary,omitempty"`
}

type runEventsInput struct {
	Packet                          jobpacket.Packet
	Delegation                      *CEODelegation
	WorkspaceBrief                  *workspace.Brief
	ProviderHealthAvoidedRouteCount int
	ProviderHealthAvoidedProviders  []string
	Results                         []subagent.Result
	Checks                          []checkrunner.Result
	PatchAudit                      []PatchAuditEntry
	PatchPreviewEvents              []PatchPreviewEvent
	PatchApproval                   *PatchApproval
	CEOReview                       *CEOReview
	LifecycleEvents                 []LifecycleEvent
	Verdict                         string
}

type runEventBuilder struct {
	events         []RunEvent
	lifecycleState LifecycleState
}

func buildRunEvents(input runEventsInput) []RunEvent {
	builder := runEventBuilder{lifecycleState: LifecycleCreated}
	lifecycle := newRunEventLifecycleCursor(input.LifecycleEvents)
	builder.setLifecycle(lifecycle.stateFor(LifecyclePlanning))
	builder.add(RunEvent{
		Kind:   "job_packet",
		Status: "ready",
		Summary: fmt.Sprintf(
			"%s/%s, %d subagent(s)",
			input.Packet.TaskProfile.Kind,
			input.Packet.TaskProfile.RiskLevel,
			len(input.Packet.Subagents),
		),
	})
	if input.ProviderHealthAvoidedRouteCount > 0 {
		builder.add(RunEvent{
			Kind:          "provider_health_route",
			Status:        "rerouted",
			RouteCount:    input.ProviderHealthAvoidedRouteCount,
			ProviderNames: append([]string(nil), input.ProviderHealthAvoidedProviders...),
			Source:        "provider_health",
			Summary:       providerHealthRouteSummary(input.ProviderHealthAvoidedRouteCount, input.ProviderHealthAvoidedProviders),
		})
	}
	if input.WorkspaceBrief != nil {
		builder.add(RunEvent{
			Kind:    "workspace_brief",
			Status:  "ready",
			Summary: fmt.Sprintf("%d file(s), %d shown", input.WorkspaceBrief.FileCount, len(input.WorkspaceBrief.Files)),
		})
	}
	if input.Delegation != nil {
		builder.setLifecycle(lifecycle.stateFor(LifecycleDelegated))
		builder.add(RunEvent{
			Kind:    "ceo_delegation",
			Status:  "ready",
			Source:  input.Delegation.Source,
			Summary: input.Delegation.Summary,
		})
	}
	for _, result := range input.Results {
		builder.setLifecycle(lifecycle.stateFor(LifecycleDelegated))
		builder.addSubagentEvents(result)
	}
	for _, preview := range input.PatchPreviewEvents {
		builder.setLifecycle(lifecycle.stateFor(LifecyclePatchPreviewed))
		builder.add(RunEvent{
			Kind:    "patch_preview",
			Status:  "ready",
			Path:    preview.Path,
			Source:  preview.Source,
			Summary: patchPreviewEventSummary(preview.Source),
		})
	}
	if input.PatchApproval != nil {
		builder.setLifecycle(lifecycle.stateFor(LifecyclePatchPreviewed))
		builder.add(RunEvent{
			Kind:         "patch_approval",
			Status:       input.PatchApproval.Status,
			Source:       "patch_approval",
			Digest:       input.PatchApproval.PreviewDigest,
			PreviewCount: input.PatchApproval.PreviewCount,
			Summary:      patchApprovalEventSummary(input.PatchApproval),
		})
	}
	for _, entry := range input.PatchAudit {
		builder.setLifecycle(lifecycle.stateFor(LifecyclePatchApplied))
		builder.add(RunEvent{
			Kind:      "patch",
			Status:    "applied",
			AgentName: entry.AgentName,
			Path:      entry.Path,
			Source:    entry.Source,
			Summary:   "patch applied",
		})
	}
	for _, check := range input.Checks {
		builder.setLifecycle(lifecycle.stateFor(LifecycleChecking))
		builder.add(RunEvent{
			Kind:       "check",
			Status:     check.Status,
			CheckIndex: check.CheckIndex,
			Attempt:    check.Attempt,
			Summary:    strings.Join(check.Argv, " "),
		})
	}
	if input.CEOReview != nil {
		builder.setLifecycle(lifecycle.stateFor(LifecycleReviewing))
		builder.add(RunEvent{
			Kind:      "ceo_review",
			Status:    input.CEOReview.RecommendedVerdict,
			AgentName: "ceo",
			Source:    input.CEOReview.Source,
			Summary:   input.CEOReview.Summary,
		})
	}
	builder.setLifecycle(lifecycle.finalState())
	builder.add(RunEvent{
		Kind:      "verdict",
		Status:    input.Verdict,
		AgentName: "ceo",
		Summary:   "CEO final verdict",
	})
	return builder.events
}

func patchPreviewEventSummary(source string) string {
	cleanSource := strings.TrimSpace(source)
	if cleanSource == "" {
		cleanSource = "patch"
	}
	return cleanSource + " patch previewed"
}

func patchApprovalEventSummary(approval *PatchApproval) string {
	return fmt.Sprintf("%d patch preview(s) %s with digest %s", approval.PreviewCount, approval.Status, approval.PreviewDigest)
}

func (b *runEventBuilder) add(event RunEvent) {
	event.Index = len(b.events) + 1
	if event.LifecycleState == "" {
		event.LifecycleState = b.lifecycleState
	}
	event.Summary = trimRunEventSummary(event.Summary)
	b.events = append(b.events, event)
}

func (b *runEventBuilder) setLifecycle(state LifecycleState) {
	if state != "" {
		b.lifecycleState = state
	}
}

func trimRunEventSummary(summary string) string {
	clean := strings.Join(strings.Fields(summary), " ")
	if len(clean) <= maxRunEventSummaryBytes {
		return clean
	}
	return clean[:maxRunEventSummaryBytes] + "..."
}

func providerHealthRouteSummary(routeCount int, providerNames []string) string {
	if len(providerNames) == 0 {
		return fmt.Sprintf("%d route(s) moved away from avoided provider(s)", routeCount)
	}
	return fmt.Sprintf("%d route(s) moved away from avoided provider(s): %s", routeCount, strings.Join(providerNames, ", "))
}
