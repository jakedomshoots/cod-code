package ceo

import (
	"ceoharness/internal/config"
	"ceoharness/internal/history"
)

type RunLedger = history.RunLedger

type RunLedgerInput struct {
	Owner                  string
	Verdict                string
	NextAction             string
	VerificationContract   VerificationContract
	ChangedFiles           []string
	ProviderRouteDecisions []config.ProviderRouteDecision
}

func NewRunLedger(input RunLedgerInput) RunLedger {
	return RunLedger{
		Owner:                input.Owner,
		Verdict:              input.Verdict,
		NextAction:           input.NextAction,
		VerificationStatus:   input.VerificationContract.Status,
		RequiredCheckCount:   input.VerificationContract.RequiredCheckCount,
		CheckAttemptCount:    input.VerificationContract.CheckAttemptCount,
		ChangedFileCount:     len(input.ChangedFiles),
		ChangedFiles:         append([]string(nil), input.ChangedFiles...),
		ProviderRouteCount:   len(input.ProviderRouteDecisions),
		ProviderRouteReasons: providerRouteReasons(input.ProviderRouteDecisions),
	}
}

func providerRouteReasons(decisions []config.ProviderRouteDecision) []string {
	reasons := make([]string, 0, len(decisions))
	seen := map[string]struct{}{}
	for _, decision := range decisions {
		if decision.Reason == "" {
			continue
		}
		if _, ok := seen[decision.Reason]; ok {
			continue
		}
		seen[decision.Reason] = struct{}{}
		reasons = append(reasons, decision.Reason)
	}
	return reasons
}
