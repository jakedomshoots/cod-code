package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const JobLogPath = "ceo-artifacts/jobs.jsonl"

var (
	ErrEntryNotFound    = errors.New("history entry not found")
	ErrInvalidCreatedAt = errors.New("invalid history created_at")
	ErrInvalidTimeRange = errors.New("invalid history time range")
)

type Entry struct {
	ID                            string           `json:"id"`
	CreatedAt                     string           `json:"created_at,omitempty"`
	Task                          string           `json:"task"`
	TaskKind                      string           `json:"task_kind,omitempty"`
	RiskLevel                     string           `json:"risk_level,omitempty"`
	RiskAreas                     []string         `json:"risk_areas,omitempty"`
	Verdict                       string           `json:"verdict"`
	LifecycleState                string           `json:"lifecycle_state,omitempty"`
	LifecycleEvents               []LifecycleEvent `json:"lifecycle_events,omitempty"`
	RunLedger                     *RunLedger       `json:"run_ledger,omitempty"`
	ChangedFiles                  []string         `json:"changed_files"`
	ExecutionPlanStepCount        int              `json:"execution_plan_step_count,omitempty"`
	ExecutionPlanNextAction       string           `json:"execution_plan_next_action,omitempty"`
	SubagentCount                 int              `json:"subagent_count"`
	ReusedSubagentCount           int              `json:"reused_subagent_count,omitempty"`
	SubagentAttemptCount          int              `json:"subagent_attempt_count,omitempty"`
	SubagentRetryCount            int              `json:"subagent_retry_count,omitempty"`
	SubagentRetriedCount          int              `json:"subagent_retried_count,omitempty"`
	SubagentRetryExhaustedCount   int              `json:"subagent_retry_exhausted_count,omitempty"`
	SubagentNoProgressStopCount   int              `json:"subagent_no_progress_stop_count,omitempty"`
	CheckCount                    int              `json:"check_count"`
	PatchCount                    int              `json:"patch_count"`
	CLIPatchCount                 int              `json:"cli_patch_count,omitempty"`
	ModelPatchCount               int              `json:"model_patch_count,omitempty"`
	CheckFixCount                 int              `json:"check_fix_count,omitempty"`
	ProviderErrorCount            int              `json:"provider_error_count,omitempty"`
	ProviderUnauthorizedCount     int              `json:"provider_unauthorized_count,omitempty"`
	ProviderRateLimitedCount      int              `json:"provider_rate_limited_count,omitempty"`
	ProviderUnavailableCount      int              `json:"provider_unavailable_count,omitempty"`
	ProviderEstimatedCostMicroUSD int64            `json:"provider_estimated_cost_microusd,omitempty"`
	ProviderCostBudgetMicroUSD    int64            `json:"provider_cost_budget_microusd,omitempty"`
	ProviderCostOverBudget        bool             `json:"provider_cost_over_budget,omitempty"`
	ProviderHealth                []ProviderHealth `json:"provider_health,omitempty"`
}

type Store struct {
	root string
	now  func() time.Time
}

func New(root string) (Store, error) {
	return NewWithClock(root, func() time.Time { return time.Now().UTC() })
}

func NewWithClock(root string, now func() time.Time) (Store, error) {
	cleanRoot := strings.TrimSpace(root)
	if cleanRoot == "" {
		return Store{}, errors.New("history root is required")
	}
	if now == nil {
		return Store{}, errors.New("history clock is required")
	}
	return Store{root: cleanRoot, now: now}, nil
}

func (s Store) Path() string {
	return JobLogPath
}

func (s Store) Append(ctx context.Context, entry Entry) (stored Entry, err error) {
	if err := ctx.Err(); err != nil {
		return Entry{}, err
	}
	if strings.TrimSpace(entry.ID) == "" {
		entries, err := s.ReadAll(ctx)
		if err != nil {
			return Entry{}, fmt.Errorf("read history before append: %w", err)
		}
		entry.ID = fmt.Sprintf("job-%06d", len(entries)+1)
	}
	if strings.TrimSpace(entry.CreatedAt) == "" {
		entry.CreatedAt = s.now().UTC().Format(time.RFC3339Nano)
	}
	path := s.fullPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Entry{}, fmt.Errorf("create history dir: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return Entry{}, fmt.Errorf("open history: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close history: %w", closeErr)
		}
	}()

	stored = entry
	if err := json.NewEncoder(file).Encode(stored); err != nil {
		return Entry{}, fmt.Errorf("encode history entry: %w", err)
	}
	return stored, nil
}

func (s Store) ReadAll(ctx context.Context) (entries []Entry, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries = []Entry{}
	file, err := os.Open(s.fullPath())
	if errors.Is(err, os.ErrNotExist) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open history: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close history: %w", closeErr)
		}
	}()

	decoder := json.NewDecoder(file)
	for {
		var entry Entry
		if err := decoder.Decode(&entry); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("decode history entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s Store) fullPath() string {
	return filepath.Join(s.root, JobLogPath)
}
