package ceo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ceoharness/internal/subagent"
)

type modelPatchCapRunner struct{}

func (r modelPatchCapRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	summary := "ok"
	if packet.AgentName == "coder" {
		summary = `{"patches":[{"path":"a.txt","old":"old","new":"new"},{"path":"b.txt","old":"old","new":"new"}]}`
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

func Test_Runtime_RunJob_rejects_model_patch_count_over_limit(t *testing.T) {
	// Given
	runtime := NewRuntimeWithSubagentRunner(modelPatchCapRunner{})
	root := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("old"), 0o644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task:              "Patch too much",
		WorkspaceDir:      root,
		ApplyModelPatches: true,
		MaxModelPatches:   1,
	})

	// Then
	if err == nil {
		t.Fatal("expected model patch cap error")
	}
	if !strings.Contains(err.Error(), "max model patches is 1") {
		t.Fatalf("error = %q, want max model patches", err.Error())
	}
	got, err := os.ReadFile(filepath.Join(root, "a.txt"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("a.txt = %q, want unchanged old", string(got))
	}
}
