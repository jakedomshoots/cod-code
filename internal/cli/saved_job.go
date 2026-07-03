package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
)

func runJobLookup(ctx context.Context, out io.Writer, workspaceDir string, jobID string) error {
	store, err := history.New(workspaceDir)
	if err != nil {
		return err
	}
	jobID, err = resolveSavedJobID(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	entry, err := store.FindByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("find history job: %w", err)
	}
	judgment, judgmentPath, err := findHumanJudgment(ctx, store, jobID)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jobReport{
		HistoryPath:    store.Path(),
		Job:            entry,
		HumanJudgment:  judgment,
		HumanJudgePath: judgmentPath,
	}); err != nil {
		return fmt.Errorf("write job report: %w", err)
	}
	return nil
}

func runJobReportLookup(ctx context.Context, out io.Writer, workspaceDir string, jobID string) error {
	store, err := history.New(workspaceDir)
	if err != nil {
		return err
	}
	jobID, err = resolveSavedJobID(ctx, workspaceDir, jobID)
	if err != nil {
		return err
	}
	snapshot, err := store.ReadReportSnapshotWithMetadata(ctx, jobID)
	if err != nil {
		return fmt.Errorf("find job report: %w", err)
	}
	report := snapshot.Payload
	if snapshot.Metadata.Legacy {
		report, err = reportPayloadWithCompatibility(snapshot)
		if err != nil {
			return err
		}
	}
	if _, err := out.Write(report); err != nil {
		return fmt.Errorf("write job report: %w", err)
	}
	if !strings.HasSuffix(string(report), "\n") {
		if _, err := io.WriteString(out, "\n"); err != nil {
			return fmt.Errorf("write job report newline: %w", err)
		}
	}
	return nil
}

func reportPayloadWithCompatibility(snapshot history.ReportSnapshot) ([]byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(snapshot.Payload, &payload); err != nil {
		return nil, fmt.Errorf("decode legacy job report: %w", err)
	}
	compatibility := struct {
		Status               string `json:"status"`
		Warning              string `json:"warning"`
		AssumedSchemaVersion int    `json:"assumed_schema_version"`
		ReaderSchemaVersion  int    `json:"reader_schema_version"`
	}{
		Status:               "legacy",
		Warning:              snapshot.Metadata.Warning,
		AssumedSchemaVersion: snapshot.Metadata.SchemaVersion,
		ReaderSchemaVersion:  history.ReportSchemaVersion,
	}
	compatibilityJSON, err := json.Marshal(compatibility)
	if err != nil {
		return nil, fmt.Errorf("encode schema compatibility: %w", err)
	}
	payload["schema_compatibility"] = compatibilityJSON
	report, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode legacy job report: %w", err)
	}
	return report, nil
}

func optionsWithRerunTask(ctx context.Context, opts options) (options, error) {
	if strings.TrimSpace(opts.rerunJobID) == "" {
		return opts, nil
	}
	if strings.TrimSpace(opts.task) != "" {
		return options{}, fmt.Errorf("--rerun cannot be combined with task text")
	}
	jobID, err := resolveFailedJobID(ctx, opts.workspaceDir, opts.rerunJobID)
	if err != nil {
		return options{}, err
	}
	store, err := history.New(opts.workspaceDir)
	if err != nil {
		return options{}, err
	}
	entry, err := store.FindByID(ctx, jobID)
	if err != nil {
		return options{}, fmt.Errorf("find rerun job: %w", err)
	}
	if strings.TrimSpace(entry.Task) == "" {
		return options{}, fmt.Errorf("rerun job task is empty")
	}
	opts.task = entry.Task
	opts.rerunContextJobID = jobID
	loaded, err := loadCompactJobContext(ctx, opts.workspaceDir, jobID)
	if err != nil {
		return options{}, err
	}
	if len(loaded.Context.FailedChecks) > 0 {
		opts.task = taskWithPriorJobContext(opts.task, loaded.Context)
		opts.scorerFailedChecks = repairFailureDetailsFromCompactChecks(loaded.Context.FailedChecks)
	}
	return opts, nil
}

func resolveFailedJobID(ctx context.Context, workspaceDir string, rawID string) (string, error) {
	cleanID := strings.TrimSpace(rawID)
	switch strings.ToLower(cleanID) {
	case "latest", "last":
		return latestFailedJobID(ctx, workspaceDir)
	default:
		return cleanID, nil
	}
}

func latestFailedJobID(ctx context.Context, workspaceDir string) (string, error) {
	store, err := history.New(workspaceDir)
	if err != nil {
		return "", err
	}
	entries, err := store.ReadByVerdict(ctx, "fail")
	if err != nil {
		return "", err
	}
	for index := len(entries) - 1; index >= 0; index-- {
		if id := strings.TrimSpace(entries[index].ID); id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("no failed jobs in history")
}

func repairFailureDetailsFromCompactChecks(checks []compactCheckResult) []ceo.RepairFailureDetail {
	details := make([]ceo.RepairFailureDetail, 0, len(checks))
	for _, check := range checks {
		if check.Status == "pass" {
			continue
		}
		name := "command:" + strings.Join(check.Command, " ")
		message := strings.TrimSpace(check.FailureExcerpt)
		if message == "" {
			message = fmt.Sprintf("exit code %d", check.ExitCode)
		}
		details = append(details, ceo.RepairFailureDetail{
			Name:    name,
			Status:  check.Status,
			Message: message,
		})
	}
	return details
}

func findHumanJudgment(ctx context.Context, store history.Store, jobID string) (*history.HumanJudgment, string, error) {
	judgment, err := store.ReadHumanJudgment(ctx, jobID)
	if err != nil {
		if errors.Is(err, history.ErrEntryNotFound) {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("find human judgment: %w", err)
	}
	path, err := history.HumanJudgmentPath(jobID)
	if err != nil {
		return nil, "", err
	}
	return &judgment, path, nil
}
