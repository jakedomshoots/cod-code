package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"ceoharness/internal/history"
)

type historyQuery struct {
	workspaceDir   string
	verdict        string
	task           string
	limit          int
	summaryOnly    bool
	since          string
	until          string
	provider       string
	recommendation string
	topProviders   int
}

type historyReport struct {
	HistoryPath string       `json:"history_path"`
	History     []historyRow `json:"history"`
	Verdict     string       `json:"verdict_filter,omitempty"`
	Task        string       `json:"task_filter,omitempty"`
	Limit       int          `json:"limit,omitempty"`
	Since       string       `json:"since,omitempty"`
	Until       string       `json:"until,omitempty"`
}

type historySummary struct {
	TotalCount                    int            `json:"total_count"`
	VerdictCounts                 map[string]int `json:"verdict_counts"`
	TaskKindCounts                map[string]int `json:"task_kind_counts,omitempty"`
	RiskLevelCounts               map[string]int `json:"risk_level_counts,omitempty"`
	SubagentCount                 int            `json:"subagent_count"`
	ReusedSubagentCount           int            `json:"reused_subagent_count"`
	SubagentAttemptCount          int            `json:"subagent_attempt_count"`
	SubagentRetryCount            int            `json:"subagent_retry_count"`
	SubagentRetriedCount          int            `json:"subagent_retried_count"`
	SubagentRetryExhaustedCount   int            `json:"subagent_retry_exhausted_count"`
	SubagentNoProgressStopCount   int            `json:"subagent_no_progress_stop_count"`
	CheckCount                    int            `json:"check_count"`
	PatchCount                    int            `json:"patch_count"`
	CLIPatchCount                 int            `json:"cli_patch_count"`
	ModelPatchCount               int            `json:"model_patch_count"`
	CheckFixCount                 int            `json:"check_fix_count"`
	ProviderErrorCount            int            `json:"provider_error_count"`
	ProviderUnauthorizedCount     int            `json:"provider_unauthorized_count"`
	ProviderRateLimitedCount      int            `json:"provider_rate_limited_count"`
	ProviderUnavailableCount      int            `json:"provider_unavailable_count"`
	ProviderEstimatedCostMicroUSD int64          `json:"provider_estimated_cost_microusd"`
	ProviderCostOverBudgetCount   int            `json:"provider_cost_over_budget_count"`
	HumanJudgmentCount            int            `json:"human_judgment_count"`
	HumanVerdictCounts            map[string]int `json:"human_verdict_counts"`
	RecoveryStateCounts           map[string]int `json:"recovery_state_counts"`
	RetryableCount                int            `json:"retryable_count"`
}

type historySummaryReport struct {
	HistoryPath string         `json:"history_path"`
	Summary     historySummary `json:"summary"`
	Verdict     string         `json:"verdict_filter,omitempty"`
	Task        string         `json:"task_filter,omitempty"`
	Limit       int            `json:"limit,omitempty"`
	Since       string         `json:"since,omitempty"`
	Until       string         `json:"until,omitempty"`
}

type jobReport struct {
	HistoryPath    string                 `json:"history_path"`
	Job            history.Entry          `json:"job"`
	HumanJudgment  *history.HumanJudgment `json:"human_judgment,omitempty"`
	HumanJudgePath string                 `json:"human_judgment_path,omitempty"`
}

func runHistory(ctx context.Context, out io.Writer, query historyQuery) error {
	store, entries, err := readHistoryEntries(ctx, query)
	if err != nil {
		return err
	}
	judgments, err := readHumanJudgmentsForHistory(ctx, store, entries)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if query.summaryOnly {
		if err := encoder.Encode(historySummaryReport{
			HistoryPath: store.Path(),
			Summary:     summarizeHistory(entries, judgments),
			Verdict:     query.verdict,
			Task:        query.task,
			Limit:       query.limit,
			Since:       query.since,
			Until:       query.until,
		}); err != nil {
			return fmt.Errorf("write history summary report: %w", err)
		}
		return nil
	}
	if err := encoder.Encode(historyReport{
		HistoryPath: store.Path(),
		History:     historyRowsWithJudgments(entries, judgments),
		Verdict:     query.verdict,
		Task:        query.task,
		Limit:       query.limit,
		Since:       query.since,
		Until:       query.until,
	}); err != nil {
		return fmt.Errorf("write history report: %w", err)
	}
	return nil
}

func summarizeHistory(entries []history.Entry, judgments map[string]history.HumanJudgment) historySummary {
	summary := historySummary{
		TotalCount:          len(entries),
		VerdictCounts:       map[string]int{},
		TaskKindCounts:      map[string]int{},
		RiskLevelCounts:     map[string]int{},
		HumanVerdictCounts:  map[string]int{},
		RecoveryStateCounts: map[string]int{},
	}
	for _, entry := range entries {
		verdict := strings.TrimSpace(entry.Verdict)
		if verdict == "" {
			verdict = "unknown"
		}
		summary.VerdictCounts[verdict]++
		if taskKind := strings.TrimSpace(entry.TaskKind); taskKind != "" {
			summary.TaskKindCounts[taskKind]++
		}
		if riskLevel := strings.TrimSpace(entry.RiskLevel); riskLevel != "" {
			summary.RiskLevelCounts[riskLevel]++
		}
		summary.SubagentCount += entry.SubagentCount
		summary.ReusedSubagentCount += entry.ReusedSubagentCount
		summary.SubagentAttemptCount += entry.SubagentAttemptCount
		summary.SubagentRetryCount += entry.SubagentRetryCount
		summary.SubagentRetriedCount += entry.SubagentRetriedCount
		summary.SubagentRetryExhaustedCount += entry.SubagentRetryExhaustedCount
		summary.SubagentNoProgressStopCount += entry.SubagentNoProgressStopCount
		summary.CheckCount += entry.CheckCount
		summary.PatchCount += entry.PatchCount
		summary.CLIPatchCount += entry.CLIPatchCount
		summary.ModelPatchCount += entry.ModelPatchCount
		summary.CheckFixCount += entry.CheckFixCount
		summary.ProviderErrorCount += entry.ProviderErrorCount
		summary.ProviderUnauthorizedCount += entry.ProviderUnauthorizedCount
		summary.ProviderRateLimitedCount += entry.ProviderRateLimitedCount
		summary.ProviderUnavailableCount += entry.ProviderUnavailableCount
		summary.ProviderEstimatedCostMicroUSD += entry.ProviderEstimatedCostMicroUSD
		if entry.ProviderCostOverBudget {
			summary.ProviderCostOverBudgetCount++
		}
		if judgment, ok := judgments[entry.ID]; ok {
			summary.HumanJudgmentCount++
			summary.HumanVerdictCounts[judgment.Verdict]++
		}
		judgment, judged := judgments[entry.ID]
		recovery := buildRecoveryView(entry, judgment, judged)
		summary.RecoveryStateCounts[recovery.State]++
		if recovery.Retryable {
			summary.RetryableCount++
		}
	}
	return summary
}

func parseHistoryRange(sinceRaw string, untilRaw string) (history.TimeRange, error) {
	since, err := parseHistoryTime("--since", sinceRaw)
	if err != nil {
		return history.TimeRange{}, err
	}
	until, err := parseHistoryTime("--until", untilRaw)
	if err != nil {
		return history.TimeRange{}, err
	}
	return history.TimeRange{Since: since, Until: until}, nil
}

func parseHistoryTime(flag string, raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be RFC3339 timestamp: %w", flag, err)
	}
	return parsed, nil
}
