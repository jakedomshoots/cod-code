package ceo

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ceoharness/internal/jobpacket"
	"ceoharness/internal/model"
	"ceoharness/internal/subagent"
)

type sequenceCEOModelClient struct {
	responses []string
	prompts   []string
}

func (c *sequenceCEOModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.prompts = append(c.prompts, req.Prompt)
	if len(c.responses) == 0 {
		return model.Response{}, model.ErrPromptRequired
	}
	text := c.responses[0]
	c.responses = c.responses[1:]
	return model.Response{
		Text:        text,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runtime_RunJob_lets_model_ceo_select_subagents_when_custom_delegation_is_available(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["security"],"summary":"Security review is the only needed lane."}`,
			`{"recommended_verdict":"pass","summary":"Selected lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix auth flow",
		Subagents: []jobpacket.Subagent{
			{Name: "planner", Role: "break down work"},
			{Name: "security", Role: "review auth risks"},
		},
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil {
		t.Fatal("expected CEO delegation")
	}
	if report.CEODelegation.SelectedSubagents[0] != "security" {
		t.Fatalf("selected subagents = %#v, want security", report.CEODelegation.SelectedSubagents)
	}
	if len(report.SubagentResults) != 1 || report.SubagentResults[0].AgentName != "security" {
		t.Fatalf("subagent results = %#v, want only security", report.SubagentResults)
	}
	if !strings.Contains(client.prompts[0], "candidate_subagents") {
		t.Fatalf("delegation prompt = %q, want candidate subagents", client.prompts[0])
	}
}

func Test_Runtime_RunJob_lets_model_ceo_select_default_subagents(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["coder"],"summary":"Default coding task only needs coder."}`,
			`{"recommended_verdict":"pass","summary":"Coder lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil {
		t.Fatal("expected CEO delegation")
	}
	if len(report.CEODelegation.SelectedSubagents) != 1 || report.CEODelegation.SelectedSubagents[0] != "coder" {
		t.Fatalf("selected subagents = %#v, want coder", report.CEODelegation.SelectedSubagents)
	}
	if len(report.SubagentResults) != 1 || report.SubagentResults[0].AgentName != "coder" {
		t.Fatalf("subagent results = %#v, want only coder", report.SubagentResults)
	}
	if !strings.Contains(client.prompts[0], "scanner") || !strings.Contains(client.prompts[0], "coder") || !strings.Contains(client.prompts[0], "reviewer") {
		t.Fatalf("delegation prompt = %q, want default candidates", client.prompts[0])
	}
	if !strings.Contains(client.prompts[0], "smallest useful set") {
		t.Fatalf("delegation prompt = %q, want lean selection rule", client.prompts[0])
	}
}

func Test_Runtime_RunJob_ignores_duplicate_candidate_in_new_subagents(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["coder"],"new_subagents":[{"name":"coder","role":"apply patches"}],"summary":"Use coder."}`,
			`{"recommended_verdict":"pass","summary":"Coder lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.CEODelegation.NewSubagents) != 0 {
		t.Fatalf("new subagents = %#v, want duplicate candidate ignored", report.CEODelegation.NewSubagents)
	}
	if len(report.JobPacket.Subagents) != 1 || report.JobPacket.Subagents[0].Name != "coder" {
		t.Fatalf("job packet subagents = %#v, want existing coder selected", report.JobPacket.Subagents)
	}
}

func Test_Runtime_RunJob_records_ceo_delegation_route_metadata(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["coder"],"summary":"Default coding task only needs coder."}`,
			`{"recommended_verdict":"pass","summary":"Coder lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewerAndRoute(client, subagent.RouteMetadata{
		Source:       "http",
		ProviderName: "main",
	})

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Fix a failing test",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil {
		t.Fatal("expected CEO delegation")
	}
	if report.CEODelegation.ModelSource != "http" || report.CEODelegation.ProviderName != "main" {
		t.Fatalf("CEO delegation route = source %q provider %q, want http main", report.CEODelegation.ModelSource, report.CEODelegation.ProviderName)
	}
}

func Test_Runtime_RunJob_rejects_model_ceo_selection_above_subagent_budget(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["scanner","coder","reviewer"],"summary":"Too many lanes."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task:         "Fix a failing test",
		MaxSubagents: 2,
	})

	// Then
	if !errors.Is(err, ErrInvalidCEODelegation) {
		t.Fatalf("error = %v, want ErrInvalidCEODelegation", err)
	}
}

func Test_Runtime_RunJob_lets_model_ceo_create_and_select_specialist_subagent(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["planner","db_reviewer"],"new_subagents":[{"name":"db_reviewer","role":"review migrations","stage":3,"allowed_actions":["read_workspace","run_checks"]}],"assignments":{"db_reviewer":"Check migration risk only."},"summary":"Database work needs a specialist."}`,
			`{"recommended_verdict":"pass","summary":"Specialist lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Plan a database migration",
	})
	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if report.CEODelegation == nil || len(report.CEODelegation.NewSubagents) != 1 {
		t.Fatalf("CEODelegation = %#v, want one new subagent", report.CEODelegation)
	}
	if report.CEODelegation.NewSubagents[0].Name != "db_reviewer" {
		t.Fatalf("new subagents = %#v, want db_reviewer", report.CEODelegation.NewSubagents)
	}
	if report.CEODelegation.NewSubagents[0].Stage != 3 {
		t.Fatalf("new subagent stage = %d, want 3", report.CEODelegation.NewSubagents[0].Stage)
	}
	if len(report.JobPacket.Subagents) != 2 || report.JobPacket.Subagents[1].Name != "db_reviewer" {
		t.Fatalf("job packet subagents = %#v, want planner and db_reviewer", report.JobPacket.Subagents)
	}
	if report.JobPacket.Subagents[1].Stage != 3 {
		t.Fatalf("job packet stage = %d, want 3", report.JobPacket.Subagents[1].Stage)
	}
	if got := jobpacket.ActionStrings(report.JobPacket.Subagents[1].AllowedActions); strings.Join(got, ",") != "read_workspace,run_checks" {
		t.Fatalf("allowed actions = %#v, want read_workspace and run_checks", got)
	}
	if len(report.SubagentResults) != 2 || report.SubagentResults[1].AgentName != "db_reviewer" {
		t.Fatalf("subagent results = %#v, want db_reviewer to run", report.SubagentResults)
	}
	if report.SubagentResults[1].Stage != 3 {
		t.Fatalf("result stage = %d, want 3", report.SubagentResults[1].Stage)
	}
	if report.SubagentResults[1].Assignment != "Check migration risk only." {
		t.Fatalf("assignment = %q, want specialist assignment", report.SubagentResults[1].Assignment)
	}
}

func Test_Runtime_RunJob_ignores_ceo_created_subagent_when_it_is_not_selected(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["planner"],"new_subagents":[{"name":"db_reviewer","role":"review migrations"}],"summary":"Created but not selected."}`,
			`{"recommended_verdict":"pass","summary":"Planner lane passed."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	report, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Plan a database migration",
	})

	// Then
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if len(report.CEODelegation.NewSubagents) != 0 {
		t.Fatalf("new subagents = %#v, want ignored unselected subagent", report.CEODelegation.NewSubagents)
	}
	if len(report.JobPacket.Subagents) != 1 || report.JobPacket.Subagents[0].Name != "planner" {
		t.Fatalf("job packet subagents = %#v, want planner only", report.JobPacket.Subagents)
	}
}

func Test_Runtime_RunJob_rejects_invalid_ceo_created_subagent_when_selected(t *testing.T) {
	// Given
	client := &sequenceCEOModelClient{
		responses: []string{
			`{"selected_subagents":["planner","specialist"],"new_subagents":[{"name":"specialist"}],"summary":"Selected invalid specialist."}`,
		},
	}
	runtime := NewRuntimeWithCEOReviewer(client)

	// When
	_, err := runtime.RunJob(context.Background(), JobRequest{
		Task: "Plan a database migration",
	})

	// Then
	if err == nil || !strings.Contains(err.Error(), "new_subagents") {
		t.Fatalf("RunJob error = %v, want selected invalid new_subagents rejection", err)
	}
}
