package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_write_policy_observe_previews_patch_without_writing(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "observe",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello old")
	requirePatchApprovalStatus(t, out.Bytes(), "previewed")
}

func Test_Run_write_policy_preview_previews_patch_without_writing(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "preview",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello old")
	requirePatchApprovalStatus(t, out.Bytes(), "previewed")
}

func Test_Run_write_policy_trusted_local_applies_patch_without_digest(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "trusted-local",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello new")
}

func Test_Run_write_policy_approved_write_applies_patch_with_digest(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	digest := previewDigestForPatch(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "approved-write",
		"--approve-preview", digest,
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello new")
	requirePatchApprovalStatus(t, out.Bytes(), "approved")
}

func Test_Run_write_policy_approved_write_refuses_digest_for_different_patch(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	digest := previewDigestForPatch(t, root)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--write-policy", "approved-write",
		"--approve-preview", digest,
		"--replace", "app.txt", "old", "changed",
		"Patch app text",
	})

	// Then
	if err == nil {
		t.Fatal("expected preview digest mismatch")
	}
	if !strings.Contains(err.Error(), "patch approval digest mismatch") {
		t.Fatalf("error = %q, want digest mismatch", err.Error())
	}
	requireFileContent(t, target, "hello old")
}

func Test_Run_default_write_policy_previews_ordinary_patch_without_writing(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello old")
	requirePatchApprovalStatus(t, out.Bytes(), "previewed")
}

func Test_Run_default_write_policy_previews_separator_patch_without_writing(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--replace", "app.txt", "old", "bad",
		"--",
		"Patch default app",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello old")
	requirePatchApprovalStatus(t, out.Bytes(), "previewed")
}

func Test_Run_high_risk_write_defaults_to_preview_without_digest(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--replace", "app.txt", "old", "new",
		"Fix security token handling",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello old")
	requirePatchApprovalStatus(t, out.Bytes(), "previewed")
}

func Test_Run_high_risk_write_uses_trusted_local_config_opt_in(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(`{"write_policy":"trusted-local"}`), 0o644); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--replace", "app.txt", "old", "new",
		"Fix security token handling",
	})

	// Then
	requireRunSuccess(t, err, out.String())
	requireFileContent(t, target, "hello new")
}

func Test_Run_rollback_report_restores_approved_patch(t *testing.T) {
	// Given
	root, target := writePolicyFixture(t)
	digest := previewDigestForPatch(t, root)
	var applyOut bytes.Buffer
	err := Run(context.Background(), &applyOut, []string{
		"--workspace", root,
		"--approve-preview", digest,
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})
	requireRunSuccess(t, err, applyOut.String())
	requireFileContent(t, target, "hello new")
	reportPath := filepath.Join(root, "apply-report.json")
	if err := os.WriteFile(reportPath, applyOut.Bytes(), 0o644); err != nil {
		t.Fatalf("write report fixture: %v", err)
	}
	var rollbackOut bytes.Buffer

	// When
	err = Run(context.Background(), &rollbackOut, []string{
		"--workspace", root,
		"--rollback-report", reportPath,
	})

	// Then
	requireRunSuccess(t, err, rollbackOut.String())
	requireFileContent(t, target, "hello old")
	requireRollbackCount(t, rollbackOut.Bytes(), 1)
}

func Test_Run_rollback_report_removes_created_model_patch_file(t *testing.T) {
	// Given
	root := t.TempDir()
	target := filepath.Join(root, "docs", "notes.md")
	var applyOut bytes.Buffer
	t.Setenv("GO_WANT_CLI_MODEL_CREATE_FILE_PATCH", "1")
	err := Run(context.Background(), &applyOut, []string{
		"--workspace", root,
		"--write-policy", "trusted-local",
		"--apply-model-patches",
		"--model-command", os.Args[0], "-test.run=Test_HelperProcess_cli_model_create_file_patch",
		"--",
		"Create notes",
	})
	requireRunSuccess(t, err, applyOut.String())
	requireFileContent(t, target, "# Notes\n")
	reportPath := filepath.Join(root, "create-report.json")
	if err := os.WriteFile(reportPath, applyOut.Bytes(), 0o644); err != nil {
		t.Fatalf("write report fixture: %v", err)
	}
	var rollbackOut bytes.Buffer

	// When
	err = Run(context.Background(), &rollbackOut, []string{
		"--workspace", root,
		"--rollback-report", reportPath,
	})

	// Then
	requireRunSuccess(t, err, rollbackOut.String())
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("created model patch file still exists after rollback: %v", err)
	}
	requireRollbackCount(t, rollbackOut.Bytes(), 1)
}

func Test_Run_rollback_report_errors_clearly_when_report_is_missing(t *testing.T) {
	// Given
	root, _ := writePolicyFixture(t)
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{
		"--workspace", root,
		"--rollback-report", filepath.Join(root, "missing-report.json"),
	})

	// Then
	if err == nil {
		t.Fatal("expected missing rollback report error")
	}
	if !strings.Contains(err.Error(), "read rollback report") {
		t.Fatalf("error = %q, want missing rollback report context", err.Error())
	}
}

func writePolicyFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return root, target
}

func previewDigestForPatch(t *testing.T, root string) string {
	t.Helper()
	var previewOut bytes.Buffer
	err := Run(context.Background(), &previewOut, []string{
		"--workspace", root,
		"--write-policy", "preview",
		"--replace", "app.txt", "old", "new",
		"Patch app text",
	})
	requireRunSuccess(t, err, previewOut.String())
	var body struct {
		PatchApproval struct {
			PreviewDigest string `json:"preview_digest"`
		} `json:"patch_approval"`
	}
	if err := json.Unmarshal(previewOut.Bytes(), &body); err != nil {
		t.Fatalf("preview output must be JSON: %v\n%s", err, previewOut.String())
	}
	if strings.TrimSpace(body.PatchApproval.PreviewDigest) == "" {
		t.Fatal("preview digest is required")
	}
	return body.PatchApproval.PreviewDigest
}

func requirePatchApprovalStatus(t *testing.T, raw []byte, want string) {
	t.Helper()
	var body struct {
		PatchApproval struct {
			Status string `json:"status"`
		} `json:"patch_approval"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("output must be JSON: %v\n%s", err, string(raw))
	}
	if body.PatchApproval.Status != want {
		t.Fatalf("PatchApproval.Status = %q, want %q", body.PatchApproval.Status, want)
	}
}

func requireRollbackCount(t *testing.T, raw []byte, want int) {
	t.Helper()
	var body struct {
		RolledBack int `json:"rolled_back"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("rollback output must be JSON: %v\n%s", err, string(raw))
	}
	if body.RolledBack != want {
		t.Fatalf("RolledBack = %d, want %d", body.RolledBack, want)
	}
}

func requireRunSuccess(t *testing.T, err error, output string) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, output)
	}
}

func requireFileContent(t *testing.T, path string, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != want {
		t.Fatalf("content = %q, want %q", string(got), want)
	}
}
