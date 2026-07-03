package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Run_uses_ceo_model_command_to_veto_passing_run(t *testing.T) {
	// Given
	var out bytes.Buffer
	args := []string{
		"--ceo-model-command",
		os.Args[0],
		"-test.run=Test_HelperProcess_cli_ceo_model_command",
		"--",
		"Fix",
		"a",
		"failing",
		"test",
	}
	t.Setenv("GO_WANT_CLI_CEO_MODEL_COMMAND", "veto")

	// When
	err := Run(context.Background(), &out, args)

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		CEOReview struct {
			RecommendedVerdict string `json:"recommended_verdict"`
			Summary            string `json:"summary"`
		} `json:"ceo_review"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOReview.RecommendedVerdict != "fail" {
		t.Fatalf("RecommendedVerdict = %q, want fail", body.CEOReview.RecommendedVerdict)
	}
	if body.CEOReview.Summary != "CEO veto from command" {
		t.Fatalf("CEO summary = %q, want command summary", body.CEOReview.Summary)
	}
	if body.Verdict != "fail" {
		t.Fatalf("Verdict = %q, want fail", body.Verdict)
	}
}

func Test_Run_uses_ceo_model_command_from_env_to_veto_passing_run(t *testing.T) {
	// Given
	var out bytes.Buffer
	commandJSON := `[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_model_command"` +
		`]`
	t.Setenv("CEO_REVIEW_MODEL_COMMAND_JSON", commandJSON)
	t.Setenv("GO_WANT_CLI_CEO_MODEL_COMMAND", "veto")

	// When
	err := Run(context.Background(), &out, []string{"Fix", "a", "failing", "test"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		CEOReview struct {
			RecommendedVerdict string `json:"recommended_verdict"`
		} `json:"ceo_review"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOReview.RecommendedVerdict != "fail" || body.Verdict != "fail" {
		t.Fatalf("CEO review = %#v verdict = %q, want fail/fail", body.CEOReview, body.Verdict)
	}
}

func Test_Run_uses_ceo_model_command_from_workspace_config_to_veto_passing_run(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	content := `{"ceo_model_command":[` +
		`"` + os.Args[0] + `",` +
		`"-test.run=Test_HelperProcess_cli_ceo_model_command"` +
		`]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("GO_WANT_CLI_CEO_MODEL_COMMAND", "veto")

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "Fix", "a", "failing", "test"})

	// Then
	if !errors.Is(err, ErrVerdictFailed) {
		t.Fatalf("Run error = %v, want ErrVerdictFailed", err)
	}
	var body struct {
		CEOReview struct {
			Summary string `json:"summary"`
		} `json:"ceo_review"`
		Verdict string `json:"verdict"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.CEOReview.Summary != "CEO veto from command" || body.Verdict != "fail" {
		t.Fatalf("CEO review = %#v verdict = %q, want command veto", body.CEOReview, body.Verdict)
	}
}

func Test_HelperProcess_cli_ceo_model_command(t *testing.T) {
	if os.Getenv("GO_WANT_CLI_CEO_MODEL_COMMAND") != "veto" {
		return
	}
	prompt, err := io.ReadAll(os.Stdin)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	text := string(prompt)
	if strings.Contains(text, "candidate_subagents") {
		os.Stdout.WriteString(`{"selected_subagents":["scanner","coder","reviewer"],"summary":"Use the default coding crew."}`)
		os.Exit(0)
	}
	if !strings.Contains(text, "guard_verdict: pass") {
		os.Stderr.WriteString("missing guard verdict")
		os.Exit(11)
	}
	os.Stdout.WriteString(`{"recommended_verdict":"fail","summary":"CEO veto from command"}`)
	os.Exit(0)
}
