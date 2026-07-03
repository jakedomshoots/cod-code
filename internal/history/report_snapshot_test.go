package history

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Store_SaveReportSnapshot_reads_snapshot_when_job_id_is_valid(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	report := []byte(`{"job_id":"job-000001","verdict":"pass"}`)

	// When
	path, err := store.SaveReportSnapshot(context.Background(), "job-000001", report)
	if err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}
	got, err := store.ReadReportSnapshot(context.Background(), "job-000001")

	// Then
	if err != nil {
		t.Fatalf("ReadReportSnapshot returned error: %v", err)
	}
	if path != "ceo-artifacts/jobs/job-000001.json" {
		t.Fatalf("path = %q, want snapshot path", path)
	}
	if string(got) != string(report) {
		t.Fatalf("snapshot = %q, want report", string(got))
	}
}

func Test_Store_SaveReportSnapshot_rejects_path_escape_when_job_id_is_invalid(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	// When
	_, err = store.SaveReportSnapshot(context.Background(), "../escape", []byte(`{}`))

	// Then
	if err == nil {
		t.Fatal("expected invalid job id error")
	}
}

func Test_Store_ReadReportSnapshotWithMetadata_marks_legacy_when_schema_version_is_missing(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	report, err := os.ReadFile(filepath.Join("testdata", "legacy-report-no-schema.json"))
	if err != nil {
		t.Fatalf("read legacy fixture: %v", err)
	}
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", report); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	snapshot, err := store.ReadReportSnapshotWithMetadata(context.Background(), "job-000001")

	// Then
	if err != nil {
		t.Fatalf("ReadReportSnapshotWithMetadata returned error: %v", err)
	}
	if string(snapshot.Payload) != string(report) {
		t.Fatalf("Payload = %q, want legacy report", string(snapshot.Payload))
	}
	if !snapshot.Metadata.Legacy {
		t.Fatalf("Metadata = %+v, want legacy marker", snapshot.Metadata)
	}
	if snapshot.Metadata.Warning == "" {
		t.Fatalf("Metadata = %+v, want compatibility warning", snapshot.Metadata)
	}
}

func Test_Store_ReadReportSnapshotWithMetadata_reads_schema_v1_when_present(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	report := []byte(`{"schema_version":1,"job_id":"job-000001","verdict":"pass"}`)
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", report); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	snapshot, err := store.ReadReportSnapshotWithMetadata(context.Background(), "job-000001")

	// Then
	if err != nil {
		t.Fatalf("ReadReportSnapshotWithMetadata returned error: %v", err)
	}
	if snapshot.Metadata.SchemaVersion != ReportSchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", snapshot.Metadata.SchemaVersion, ReportSchemaVersion)
	}
	if snapshot.Metadata.Legacy {
		t.Fatalf("Metadata = %+v, want v1 report", snapshot.Metadata)
	}
}

func Test_Store_ReadReportSnapshotWithMetadata_returns_error_when_snapshot_is_malformed(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	reportPath := filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte(`{"job_id":`), 0o644); err != nil {
		t.Fatalf("write malformed fixture: %v", err)
	}

	// When
	_, err = store.ReadReportSnapshotWithMetadata(context.Background(), "job-000001")

	// Then
	if err == nil {
		t.Fatal("expected malformed snapshot error")
	}
}

func Test_Store_InspectReportRecovery_reports_corrupt_partial_snapshot_guidance(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), Entry{
		Task:           "Interrupted run",
		Verdict:        "canceled",
		LifecycleState: "canceled",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	reportPath := filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json")
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(reportPath, []byte(`{"job_id":`), 0o644); err != nil {
		t.Fatalf("write malformed fixture: %v", err)
	}

	// When
	issues, err := store.InspectReportRecovery(context.Background())

	// Then
	if err != nil {
		t.Fatalf("InspectReportRecovery returned error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(issues))
	}
	issue := issues[0]
	if issue.JobID != "job-000001" || issue.Kind != "corrupt_snapshot" {
		t.Fatalf("issue = %+v, want corrupt snapshot for job-000001", issue)
	}
	if !strings.Contains(issue.Guidance, "--continue-job job-000001") {
		t.Fatalf("Guidance = %q, want continue-job guidance", issue.Guidance)
	}
}

func Test_Store_InspectReportRecovery_reports_canceled_job_as_recoverable(t *testing.T) {
	// Given
	root := t.TempDir()
	store, err := New(root)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if _, err := store.Append(context.Background(), Entry{
		Task:           "Canceled run",
		Verdict:        "canceled",
		LifecycleState: "canceled",
	}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	report := []byte(`{"schema_version":1,"job_id":"job-000001","verdict":"canceled","lifecycle_state":"canceled"}`)
	if _, err := store.SaveReportSnapshot(context.Background(), "job-000001", report); err != nil {
		t.Fatalf("SaveReportSnapshot returned error: %v", err)
	}

	// When
	issues, err := store.InspectReportRecovery(context.Background())

	// Then
	if err != nil {
		t.Fatalf("InspectReportRecovery returned error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("issues length = %d, want 1", len(issues))
	}
	if issues[0].Kind != "interrupted_job" || issues[0].JobID != "job-000001" {
		t.Fatalf("issue = %+v, want interrupted job recovery guidance", issues[0])
	}
}
