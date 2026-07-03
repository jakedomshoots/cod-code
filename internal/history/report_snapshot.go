package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const JobReportDir = "ceo-artifacts/jobs"
const ReportSchemaVersion = 1

const legacyReportWarning = "missing schema_version; treating as legacy report compatible with schema v1"

var (
	ErrInvalidJobID          = errors.New("invalid job id")
	ErrInvalidReportSnapshot = errors.New("invalid report snapshot")
)

type ReportSnapshot struct {
	Payload  []byte
	Metadata ReportSnapshotMetadata
}

type ReportSnapshotMetadata struct {
	SchemaVersion int
	Legacy        bool
	Warning       string
}

type ReportRecoveryIssue struct {
	JobID    string `json:"job_id"`
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Guidance string `json:"guidance"`
	Error    string `json:"error,omitempty"`
}

func (s Store) SaveReportSnapshot(ctx context.Context, jobID string, report []byte) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cleanID, err := cleanSnapshotJobID(jobID)
	if err != nil {
		return "", err
	}
	if !json.Valid(report) {
		return "", ErrInvalidReportSnapshot
	}
	relativePath := reportSnapshotPath(cleanID)
	fullPath := filepath.Join(s.root, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create report snapshot dir: %w", err)
	}
	if err := os.WriteFile(fullPath, report, 0o644); err != nil {
		return "", fmt.Errorf("write report snapshot: %w", err)
	}
	return relativePath, nil
}

func (s Store) ReadReportSnapshot(ctx context.Context, jobID string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cleanID, err := cleanSnapshotJobID(jobID)
	if err != nil {
		return nil, err
	}
	report, err := os.ReadFile(filepath.Join(s.root, reportSnapshotPath(cleanID)))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read report snapshot: %w", err)
	}
	return report, nil
}

func (s Store) ReadReportSnapshotWithMetadata(ctx context.Context, jobID string) (ReportSnapshot, error) {
	report, err := s.ReadReportSnapshot(ctx, jobID)
	if err != nil {
		return ReportSnapshot{}, err
	}
	metadata, err := reportSnapshotMetadata(report)
	if err != nil {
		return ReportSnapshot{}, err
	}
	return ReportSnapshot{Payload: report, Metadata: metadata}, nil
}

func (s Store) InspectReportRecovery(ctx context.Context) ([]ReportRecoveryIssue, error) {
	entries, err := s.ReadAll(ctx)
	if err != nil {
		return nil, err
	}
	issues := []ReportRecoveryIssue{}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cleanID, err := cleanSnapshotJobID(entry.ID)
		if err != nil {
			continue
		}
		_, err = s.ReadReportSnapshotWithMetadata(ctx, cleanID)
		if err == nil && reportEntryRecoverable(entry) {
			issues = append(issues, s.reportRecoveryIssue(cleanID, "interrupted_job", "job ended before a passing final state"))
			continue
		}
		if err == nil {
			continue
		}
		if errors.Is(err, ErrEntryNotFound) {
			issues = append(issues, s.reportRecoveryIssue(cleanID, "missing_snapshot", "report snapshot is missing"))
			continue
		}
		if errors.Is(err, ErrInvalidReportSnapshot) {
			issues = append(issues, s.reportRecoveryIssue(cleanID, "corrupt_snapshot", err.Error()))
			continue
		}
		return nil, err
	}
	return issues, nil
}

func reportSnapshotMetadata(report []byte) (ReportSnapshotMetadata, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(report, &fields); err != nil {
		return ReportSnapshotMetadata{}, fmt.Errorf("%w: %v", ErrInvalidReportSnapshot, err)
	}
	rawVersion, ok := fields["schema_version"]
	if !ok {
		return ReportSnapshotMetadata{
			SchemaVersion: 0,
			Legacy:        true,
			Warning:       legacyReportWarning,
		}, nil
	}
	var version int
	if err := json.Unmarshal(rawVersion, &version); err != nil {
		return ReportSnapshotMetadata{}, fmt.Errorf("%w: invalid schema_version", ErrInvalidReportSnapshot)
	}
	return ReportSnapshotMetadata{SchemaVersion: version}, nil
}

func reportEntryRecoverable(entry Entry) bool {
	return entry.Verdict == "canceled" || entry.LifecycleState == "canceled"
}

func (s Store) reportRecoveryIssue(jobID string, kind string, message string) ReportRecoveryIssue {
	path := reportSnapshotPath(jobID)
	return ReportRecoveryIssue{
		JobID: jobID,
		Kind:  kind,
		Path:  path,
		Guidance: fmt.Sprintf(
			"Review %s, repair or remove the partial artifact, then run ceo-packet --workspace %q --continue-job %s",
			path,
			s.root,
			jobID,
		),
		Error: message,
	}
}

func reportSnapshotPath(jobID string) string {
	return filepath.Join(JobReportDir, jobID+".json")
}

func cleanSnapshotJobID(jobID string) (string, error) {
	cleanID := strings.TrimSpace(jobID)
	if cleanID == "" || cleanID == "." || cleanID == ".." {
		return "", ErrInvalidJobID
	}
	if strings.ContainsAny(cleanID, `/\`) {
		return "", ErrInvalidJobID
	}
	return cleanID, nil
}
