package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func Test_Run_prints_doctor_report_when_doctor_flag_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--doctor"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status  string `json:"status"`
		Version string `json:"version"`
		Checks  []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Verdict    string `json:"verdict"`
			Workspace  string `json:"workspace"`
			PatchCount int    `json:"patch_count"`
			CheckCount int    `json:"check_count"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	if body.Version != "dev" {
		t.Fatalf("Version = %q, want dev", body.Version)
	}
	for _, check := range body.Checks {
		if check.Name != "golden_demo" {
			continue
		}
		if check.Status != "pass" || check.Verdict != "pass" {
			t.Fatalf("doctor check = %+v, want passing golden demo", check)
		}
		if check.PatchCount != 1 || check.CheckCount != 1 || !strings.Contains(check.Workspace, "ceo-harness-demo") {
			t.Fatalf("doctor check = %+v, want demo patch/check/workspace evidence", check)
		}
		return
	}
	t.Fatalf("Checks = %#v, want golden_demo check", body.Checks)
}

func Test_Run_prints_model_command_doctor_check_when_model_command_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	script := filepath.Join("..", "..", "examples", "command-model.sh")

	// When
	err := Run(context.Background(), &out, []string{"--doctor", "--model-command", "sh", script})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	found := false
	for _, check := range body.Checks {
		if check.Name == "model_command" && check.Status == "pass" && check.Source == "flag" && check.Error == "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Checks = %#v, want passing model_command check", body.Checks)
	}
}

func Test_Run_prints_ceo_model_command_doctor_check_when_ceo_model_command_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer
	script := filepath.Join("..", "..", "examples", "ceo-model.sh")

	// When
	err := Run(context.Background(), &out, []string{"--doctor", "--ceo-model-command", "sh", script})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	found := false
	for _, check := range body.Checks {
		if check.Name == "ceo_model_command" && check.Status == "pass" && check.Source == "flag" && check.Error == "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Checks = %#v, want passing ceo_model_command check", body.Checks)
	}
}

func Test_Run_prints_workspace_source_for_configured_doctor_adapter_checks(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	modelScript, err := filepath.Abs(filepath.Join("..", "..", "examples", "command-model.sh"))
	if err != nil {
		t.Fatalf("model script path: %v", err)
	}
	ceoScript, err := filepath.Abs(filepath.Join("..", "..", "examples", "ceo-model.sh"))
	if err != nil {
		t.Fatalf("ceo script path: %v", err)
	}
	configJSON := `{"model_command":["sh",` + strconv.Quote(modelScript) + `],"ceo_model_command":["sh",` + strconv.Quote(ceoScript) + `]}`
	if err := os.WriteFile(filepath.Join(root, ".ceo-harness.json"), []byte(configJSON), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// When
	err = Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Source string `json:"source"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &body); err != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", err, out.String())
	}
	if body.Status != "pass" {
		t.Fatalf("Status = %q, want pass", body.Status)
	}
	requireDoctorCheckSource(t, body.Checks, "model_command", "workspace")
	requireDoctorCheckSource(t, body.Checks, "ceo_model_command", "workspace")
}

func Test_Run_prints_text_doctor_report_when_text_format_is_supplied(t *testing.T) {
	// Given
	var out bytes.Buffer

	// When
	err := Run(context.Background(), &out, []string{"--doctor", "--format", "text"})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v\n%s", err, out.String())
	}
	body := out.String()
	for _, want := range []string{
		"Doctor: pass",
		"Version: dev",
		"- golden_demo [pass]",
		"source=",
		"verdict=pass",
		"patches=1",
		"checks=1",
		"ceo-harness-demo",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("doctor text missing %q:\n%s", want, body)
		}
	}
}

func Test_Run_doctor_reports_repair_guidance_for_corrupt_partial_job_artifact(t *testing.T) {
	// Given
	var out bytes.Buffer
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ceo-artifacts", "jobs"), 0o755); err != nil {
		t.Fatalf("create jobs dir: %v", err)
	}
	historyLine := `{"id":"job-000001","task":"Interrupted run","verdict":"canceled","lifecycle_state":"canceled"}` + "\n"
	if err := os.WriteFile(filepath.Join(root, "ceo-artifacts", "jobs.jsonl"), []byte(historyLine), 0o644); err != nil {
		t.Fatalf("write history fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ceo-artifacts", "jobs", "job-000001.json"), []byte(`{"job_id":`), 0o644); err != nil {
		t.Fatalf("write corrupt snapshot fixture: %v", err)
	}

	// When
	err := Run(context.Background(), &out, []string{"--workspace", root, "--doctor"})

	// Then
	if err == nil {
		t.Fatal("Run returned nil error, want doctor failure")
	}
	var body struct {
		Status string `json:"status"`
		Checks []struct {
			Name     string `json:"name"`
			Status   string `json:"status"`
			JobID    string `json:"job_id"`
			Path     string `json:"path"`
			Guidance string `json:"guidance"`
		} `json:"checks"`
	}
	if jsonErr := json.Unmarshal(out.Bytes(), &body); jsonErr != nil {
		t.Fatalf("doctor output must be JSON: %v\n%s", jsonErr, out.String())
	}
	if body.Status != "fail" {
		t.Fatalf("Status = %q, want fail", body.Status)
	}
	for _, check := range body.Checks {
		if check.Name == "job_recovery" {
			if check.Status != "fail" || check.JobID != "job-000001" {
				t.Fatalf("job_recovery check = %+v, want corrupt job failure", check)
			}
			if !strings.Contains(check.Path, "ceo-artifacts/jobs/job-000001.json") {
				t.Fatalf("Path = %q, want corrupt snapshot path", check.Path)
			}
			if !strings.Contains(check.Guidance, "--continue-job job-000001") {
				t.Fatalf("Guidance = %q, want continue-job guidance", check.Guidance)
			}
			return
		}
	}
	t.Fatalf("Checks = %#v, want job_recovery check", body.Checks)
}

func requireDoctorCheckSource(t *testing.T, checks []struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Source string `json:"source"`
}, name string, source string,
) {
	t.Helper()
	for _, check := range checks {
		if check.Name == name && check.Status == "pass" && check.Source == source {
			return
		}
	}
	t.Fatalf("Checks = %#v, want %s source %s", checks, name, source)
}
