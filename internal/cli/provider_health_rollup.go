package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/history"
)

type providerHealthRollupReport struct {
	HistoryPath    string                        `json:"history_path"`
	ProviderHealth []history.ProviderHealth      `json:"provider_health"`
	Summary        history.ProviderHealthSummary `json:"summary"`
	Verdict        string                        `json:"verdict_filter,omitempty"`
	Task           string                        `json:"task_filter,omitempty"`
	Limit          int                           `json:"limit,omitempty"`
	Since          string                        `json:"since,omitempty"`
	Until          string                        `json:"until,omitempty"`
	Provider       string                        `json:"provider,omitempty"`
	Recommendation string                        `json:"recommendation,omitempty"`
	TopProviders   int                           `json:"top_providers,omitempty"`
}

type providerHealthSummaryReport struct {
	HistoryPath    string                        `json:"history_path"`
	Summary        history.ProviderHealthSummary `json:"summary"`
	Verdict        string                        `json:"verdict_filter,omitempty"`
	Task           string                        `json:"task_filter,omitempty"`
	Limit          int                           `json:"limit,omitempty"`
	Since          string                        `json:"since,omitempty"`
	Until          string                        `json:"until,omitempty"`
	Provider       string                        `json:"provider,omitempty"`
	Recommendation string                        `json:"recommendation,omitempty"`
	TopProviders   int                           `json:"top_providers,omitempty"`
}

func runProviderHealthRollup(ctx context.Context, out io.Writer, query historyQuery) error {
	store, entries, err := readHistoryEntries(ctx, query)
	if err != nil {
		return err
	}
	policy, err := providerHealthPolicyForWorkspace(ctx, query.workspaceDir)
	if err != nil {
		return err
	}
	providerHealth := filterProviderHealth(history.AggregateProviderHealthWithPolicy(entries, policy), query.provider, query.recommendation)
	providerHealth = limitProviderHealth(providerHealth, query.topProviders)
	summary := history.SummarizeProviderHealth(providerHealth)
	report := providerHealthRollupReport{
		HistoryPath:    store.Path(),
		ProviderHealth: providerHealth,
		Summary:        summary,
		Verdict:        query.verdict,
		Task:           query.task,
		Limit:          query.limit,
		Since:          query.since,
		Until:          query.until,
		Provider:       query.provider,
		Recommendation: query.recommendation,
		TopProviders:   query.topProviders,
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if query.summaryOnly {
		if err := encoder.Encode(providerHealthSummaryReport{
			HistoryPath:    store.Path(),
			Summary:        summary,
			Verdict:        query.verdict,
			Task:           query.task,
			Limit:          query.limit,
			Since:          query.since,
			Until:          query.until,
			Provider:       query.provider,
			Recommendation: query.recommendation,
			TopProviders:   query.topProviders,
		}); err != nil {
			return fmt.Errorf("write provider health summary report: %w", err)
		}
		return nil
	}
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write provider health report: %w", err)
	}
	return nil
}

func filterProviderHealth(rows []history.ProviderHealth, providerName string, recommendation string) []history.ProviderHealth {
	cleanProvider := strings.TrimSpace(providerName)
	cleanRecommendation := strings.TrimSpace(recommendation)
	if cleanProvider == "" && cleanRecommendation == "" {
		return rows
	}
	filtered := []history.ProviderHealth{}
	for _, row := range rows {
		if cleanProvider != "" && row.ProviderName != cleanProvider {
			continue
		}
		if cleanRecommendation != "" && row.Recommendation != cleanRecommendation {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func limitProviderHealth(rows []history.ProviderHealth, limit int) []history.ProviderHealth {
	if limit < 1 || limit >= len(rows) {
		return rows
	}
	return rows[:limit]
}
