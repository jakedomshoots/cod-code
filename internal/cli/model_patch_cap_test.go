package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_rejects_model_patch_count_over_limit(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	target := filepath.Join(root, "app.txt")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	args := []string{
		"--workspace",
		root,
		"--apply-model-patches",
		"--max-model-patches",
		"1",
		"--model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_model_patch_cap",
		"--",
		"Patch",
		"too",
		"much",
	}
	t.Setenv("GO_WANT_CLI_MODEL_PATCH_CAP", "1")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if err == nil {
		t.Fatal("expected model patch cap error")
	}
	if !strings.Contains(err.Error(), "max model patches is 1") {
		t.Fatalf("error = %q, want max model patches", err.Error())
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if string(got) != "old" {
		t.Fatalf("content = %q, want unchanged old", string(got))
	}
}

func Test_HelperProcess_cli_model_patch_cap(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_MODEL_PATCH_CAP") != "1" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if strings.Contains(string(prompt), "agent: coder") {
		os.Stdout.WriteString(`{"patches":[{"path":"app.txt","old":"old","new":"new"},{"path":"other.txt","old":"old","new":"new"}]}`)
		os.Exit(0)
	}
	os.Stdout.WriteString("ok")
	os.Exit(0)
}
