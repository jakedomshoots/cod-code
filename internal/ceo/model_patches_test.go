package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type modelPatchRunner struct{}

func (r modelPatchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"patches":[{"path":"app.txt","old":"old","new":"new"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

type modelCreateFilePatchRunner struct{}

func (r modelCreateFilePatchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"patches":[{"path":"docs/notes.md","content":"# Notes\n"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

type modelFullFileContentPatchRunner struct{}

func (r modelFullFileContentPatchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"patches":[{"path":"app.txt","content":"hello new\n"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

type modelLooseWholeFilePatchRunner struct{}

func (r modelLooseWholeFilePatchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"patches":[{"path":"frontend/state.js","old":"module.exports = { benchmarkFixture: \"TODO: update this benchmark fixture\" }","new":"module.exports = { benchmarkFixture: \"optimistic update keeps rollback evidence\" };"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

type modelToolRequestPatchRunner struct{}

func (r modelToolRequestPatchRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"tool_requests":[{"action":"propose_patch","path":"app.txt","old":"old","new":"new"}]}`
	}
	return subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Status:          "pass",
		Attempts:        1,
		ContextReceived: packet.ContextMode,
		ContextBytes:    len(packet.Task),
		Summary:         summary,
		Evidence:        []string{"ok"},
	}, nil
}

func Test_Runtime_RunJob_applies_coder_model_patch_when_enabled(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Patch app text",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	if len(report.PatchResults) != 1 {
		t.Fatalf("PatchResults length = %d, want 1", len(report.PatchResults))
	}
	if report.PatchResults[0].Path != "app.txt" {
		t.Fatalf("PatchResults[0].Path = %q, want app.txt", report.PatchResults[0].Path)
	}
	if len(report.PatchAudit) != 1 {
		t.Fatalf("PatchAudit length = %d, want 1", len(report.PatchAudit))
	}
	if report.PatchAudit[0].Source != "model" || report.PatchAudit[0].AgentName != "coder" {
		t.Fatalf("PatchAudit[0] = %+v, want coder model source", report.PatchAudit[0])
	}
}

func Test_Runtime_RunJob_applies_coder_tool_request_patch_when_enabled(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelToolRequestPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Patch app text",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new" {
		t.Fatalf("content = %q, want hello new", string(got))
	}
	if len(report.PatchResults) != 1 || report.PatchResults[0].Path != "app.txt" {
		t.Fatalf("PatchResults = %+v, want app.txt patch from tool request", report.PatchResults)
	}
}

func Test_Runtime_RunJob_applies_coder_model_create_file_patch_when_enabled(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelCreateFilePatchRunner{})
	root := t.TempDir()

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Create notes",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "docs", "notes.md"))
	if err != nil {
		t.Fatalf("read created file: %v", err)
	}
	if string(got) != "# Notes\n" {
		t.Fatalf("content = %q, want created notes", string(got))
	}
	if len(report.PatchResults) != 1 || report.PatchResults[0].Path != "docs/notes.md" {
		t.Fatalf("PatchResults = %+v, want docs/notes.md", report.PatchResults)
	}
}

func Test_Runtime_RunJob_applies_coder_model_content_patch_to_existing_file(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelFullFileContentPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Replace app text",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if string(got) != "hello new\n" {
		t.Fatalf("content = %q, want full-file replacement", string(got))
	}
	if len(report.PatchResults) != 1 || report.PatchResults[0].Old != "hello old\n" || report.PatchResults[0].New != "hello new\n" {
		t.Fatalf("PatchResults = %+v, want full-file old/new", report.PatchResults)
	}
}

func Test_Runtime_RunJob_applies_loose_whole_file_model_patch(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelLooseWholeFilePatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "frontend", "state.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(target, []byte("module.exports = { benchmarkFixture: \"TODO: update this benchmark fixture\" };\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Patch benchmark fixture",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read patched file: %v", err)
	}
	if !strings.Contains(string(got), "optimistic update keeps rollback evidence") {
		t.Fatalf("content = %q, want loose whole-file replacement", string(got))
	}
	if len(report.PatchResults) != 1 || report.PatchResults[0].Path != "frontend/state.js" || report.PatchResults[0].Diff == "" {
		t.Fatalf("PatchResults = %+v, want loose patch result", report.PatchResults)
	}
}

func Test_Runtime_RunJob_records_explicit_patch_audit(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "manual.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch app text",
		WorkspaceDir: root,
		Patches: []PatchRequest{
			{Path: "manual.txt", Old: "old", New: "new"},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.PatchAudit) != 1 {
		t.Fatalf("PatchAudit length = %d, want 1", len(report.PatchAudit))
	}
	if report.PatchAudit[0].Path != "manual.txt" || report.PatchAudit[0].Source != "cli" {
		t.Fatalf("PatchAudit[0] = %+v, want CLI patch source", report.PatchAudit[0])
	}
}

func Test_proposedModelPatches_reads_typed_patch_proposals_when_summary_is_plain_text(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "coder",
			Status:         "pass",
			AllowedActions: []string{"propose_patch"},
			Summary:        "patch ready",
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt", Old: "old", New: "new"},
			},
		},
	}

	// When
	patches, err := proposedModelPatches(results)
	// Then
	if err != nil {
		t.Fatalf("proposedModelPatches returned error: %v", err)
	}
	if len(patches) != 1 || patches[0].Path != "app.txt" {
		t.Fatalf("patches = %+v, want typed app patch", patches)
	}
}

func Test_proposedModelPatches_reads_typed_create_file_proposals_when_summary_is_plain_text(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "coder",
			Status:         "pass",
			AllowedActions: []string{"propose_patch"},
			Summary:        "patch ready",
			PatchProposals: []subagent.PatchProposal{
				{Path: "docs/notes.md", Content: "# Notes\n"},
			},
		},
	}

	// When
	patches, err := proposedModelPatches(results)
	// Then
	if err != nil {
		t.Fatalf("proposedModelPatches returned error: %v", err)
	}
	if len(patches) != 1 || patches[0].Content != "# Notes\n" {
		t.Fatalf("patches = %+v, want create file content", patches)
	}
}

func Test_proposedModelPatches_treats_new_without_old_as_full_file_content(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "coder",
			Status:         "pass",
			AllowedActions: []string{"propose_patch"},
			Summary:        "patch ready",
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt", New: "full file\n"},
			},
		},
	}

	// When
	patches, err := proposedModelPatches(results)
	// Then
	if err != nil {
		t.Fatalf("proposedModelPatches returned error: %v", err)
	}
	if len(patches) != 1 || patches[0].Content != "full file\n" || patches[0].New != "" {
		t.Fatalf("patches = %+v, want normalized full-file content", patches)
	}
}

func Test_proposedModelPatches_ignores_empty_path_only_patch(t *testing.T) {
	// Given
	results := []subagent.Result{
		{
			AgentName:      "coder",
			Status:         "pass",
			AllowedActions: []string{"propose_patch"},
			Summary:        "patch ready",
			PatchProposals: []subagent.PatchProposal{
				{Path: "app.txt"},
			},
		},
	}

	// When
	patches, err := proposedModelPatches(results)
	// Then
	if err != nil {
		t.Fatalf("proposedModelPatches returned error: %v", err)
	}
	if len(patches) != 0 {
		t.Fatalf("patches = %+v, want empty path-only proposal ignored", patches)
	}
}

func Test_Runtime_RunJob_ignores_coder_model_patch_by_default(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Patch app text",
		WorkspaceDir: root,
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != "hello old" {
		t.Fatalf("content = %q, want unchanged content", string(got))
	}
	if len(report.PatchResults) != 0 {
		t.Fatalf("PatchResults length = %d, want 0", len(report.PatchResults))
	}
}

func Test_Runtime_RunJob_rejects_model_patch_without_workspace(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchRunner{})

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Patch app text",
		ApplyModelPatches: true,
	})

	// Then
	if err == nil {
		t.Fatal("expected workspace error")
	}
	if !strings.Contains(err.Error(), "workspace is required for model patches") {
		t.Fatalf("error = %q, want workspace required", err.Error())
	}
}

func Test_parseCoderPatchProposal_rejects_malformed_json(t *testing.T) {
	// When
	_, err := parseCoderPatchProposal(`{"patches":[`)

	// Then
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
	if !strings.Contains(err.Error(), "parse coder patch proposal") {
		t.Fatalf("error = %q, want parse coder patch proposal", err.Error())
	}
}
