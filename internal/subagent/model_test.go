package subagent

import (
	"context"
	"strings"
	"testing"

	"ceoharness/internal/model"
)

type captureModelClient struct {
	prompts []string
}

func (c *captureModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	c.prompts = append(c.prompts, req.Prompt)
	return model.Response{
		Text:        "model reviewed compact prompt",
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runner_Run_uses_model_client_when_prompt_is_built(t *testing.T) {
	// Given
	client := &captureModelClient{}
	runner := NewRunnerWithModel(client)

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Fix a failing test",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 256,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "model reviewed compact prompt" {
		t.Fatalf("Summary = %q, want model response", result.Summary)
	}
	if len(client.prompts) != 1 {
		t.Fatalf("prompt count = %d, want 1", len(client.prompts))
	}
	if !strings.Contains(client.prompts[0], "role: inspect scope") {
		t.Fatalf("prompt = %q, want role", client.prompts[0])
	}
}

type toolRequestModelClient struct{}

func (c toolRequestModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text:        `{"summary":"inspect workspace","tool_requests":[{"action":"read_workspace","path":"README.md"},{"action":"search_workspace","query":"TODO"}]}`,
		PromptBytes: len(req.Prompt),
	}, nil
}

func Test_Runner_Run_parses_tool_requests_when_model_returns_json_envelope(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(toolRequestModelClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect docs",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		AllowedActions:  []string{"read_workspace", "search_workspace"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.ToolRequests) != 2 {
		t.Fatalf("ToolRequests length = %d, want 2", len(result.ToolRequests))
	}
	if result.ToolRequests[0].Action != "read_workspace" || result.ToolRequests[0].Path != "README.md" {
		t.Fatalf("ToolRequests[0] = %+v, want README read request", result.ToolRequests[0])
	}
	if result.Summary != "inspect workspace" {
		t.Fatalf("Summary = %q, want parsed summary", result.Summary)
	}
}

type structuredModelClient struct{}

func (c structuredModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"summary":"patch ready","evidence":["found old text"],"tool_requests":[{"action":"read_workspace","path":"app.txt"}],"patches":[{"path":"app.txt","old":"old","new":"new"}]}`,
	}, nil
}

func Test_Runner_Run_parses_structured_model_output_when_model_returns_full_json(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(structuredModelClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Patch app",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		AllowedActions:  []string{"read_workspace", "propose_patch"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary != "patch ready" {
		t.Fatalf("Summary = %q, want parsed summary", result.Summary)
	}
	if len(result.Evidence) != 1 || result.Evidence[0] != "found old text" {
		t.Fatalf("Evidence = %+v, want parsed evidence", result.Evidence)
	}
	if len(result.ToolRequests) != 1 || result.ToolRequests[0].Path != "app.txt" {
		t.Fatalf("ToolRequests = %+v, want parsed read request", result.ToolRequests)
	}
	if len(result.PatchProposals) != 1 || result.PatchProposals[0].Path != "app.txt" {
		t.Fatalf("PatchProposals = %+v, want parsed patch", result.PatchProposals)
	}
}

type emptyToolRequestWithPatchClient struct{}

func (c emptyToolRequestWithPatchClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"summary":"patch ready","tool_requests":[{}],"patches":[{"path":"app.txt","old":"old","new":"new"}]}`,
	}, nil
}

func Test_Runner_Run_ignores_empty_tool_request_entries(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(emptyToolRequestWithPatchClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Patch file",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		AllowedActions:  []string{"propose_patch"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.ToolRequests) != 0 {
		t.Fatalf("ToolRequests = %+v, want empty entries ignored", result.ToolRequests)
	}
	if len(result.PatchProposals) != 1 || result.PatchProposals[0].Path != "app.txt" {
		t.Fatalf("PatchProposals = %+v, want parsed patch", result.PatchProposals)
	}
}

type shorthandToolRequestClient struct{}

func (c shorthandToolRequestClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"summary":"need context","tool_requests":[{"path":"app.txt"},{"query":"TODO"}]}`,
	}, nil
}

func Test_Runner_Run_normalizes_tool_request_shorthand(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(shorthandToolRequestClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect file",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		AllowedActions:  []string{"read_workspace", "search_workspace"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.ToolRequests) != 2 {
		t.Fatalf("ToolRequests = %+v, want two normalized requests", result.ToolRequests)
	}
	if result.ToolRequests[0].Action != "read_workspace" || result.ToolRequests[0].Path != "app.txt" {
		t.Fatalf("ToolRequests[0] = %+v, want read_workspace app.txt", result.ToolRequests[0])
	}
	if result.ToolRequests[1].Action != "search_workspace" || result.ToolRequests[1].Query != "TODO" {
		t.Fatalf("ToolRequests[1] = %+v, want search_workspace TODO", result.ToolRequests[1])
	}
}

type stringToolRequestClient struct{}

func (c stringToolRequestClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"summary":"need file","tool_requests":["app.txt"]}`,
	}, nil
}

func Test_Runner_Run_normalizes_string_tool_requests(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(stringToolRequestClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Inspect file",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		AllowedActions:  []string{"read_workspace"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.ToolRequests) != 1 || result.ToolRequests[0].Action != "read_workspace" || result.ToolRequests[0].Path != "app.txt" {
		t.Fatalf("ToolRequests = %+v, want string request normalized to read_workspace app.txt", result.ToolRequests)
	}
}

type createFileModelClient struct{}

func (c createFileModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"summary":"new file ready","patches":[{"path":"new.txt","content":"hello new file\n"}]}`,
	}, nil
}

func Test_Runner_Run_parses_create_file_patch_when_model_returns_content(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(createFileModelClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Create file",
		AgentName:       "coder",
		Role:            "apply bounded changes",
		ContextMode:     "lean",
		AllowedActions:  []string{"propose_patch"},
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.PatchProposals) != 1 {
		t.Fatalf("PatchProposals length = %d, want 1", len(result.PatchProposals))
	}
	if result.PatchProposals[0].Path != "new.txt" || result.PatchProposals[0].Content != "hello new file\n" {
		t.Fatalf("PatchProposals[0] = %+v, want create file content", result.PatchProposals[0])
	}
}

type needsInputModelClient struct{}

func (c needsInputModelClient) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	return model.Response{
		Text: `{"status":"needs_input","summary":"missing target repo","questions":["Which package should I change?"]}`,
	}, nil
}

func Test_Runner_Run_marks_needs_input_when_model_requests_user_input(t *testing.T) {
	// Given
	runner := NewRunnerWithModel(needsInputModelClient{})

	// When
	result, err := runner.Run(context.Background(), TaskPacket{
		Task:            "Fix ambiguous package",
		AgentName:       "scanner",
		Role:            "inspect scope",
		ContextMode:     "lean",
		MaxContextBytes: 512,
	})
	// Then
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Status != "needs_input" {
		t.Fatalf("Status = %q, want needs_input", result.Status)
	}
	if len(result.Questions) != 1 || result.Questions[0] != "Which package should I change?" {
		t.Fatalf("Questions = %+v, want model question", result.Questions)
	}
}
