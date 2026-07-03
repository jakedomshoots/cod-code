package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/subagent"
)

const demoTaskPrefix = "Golden demo: patch app.txt from hello old to hello new"

type demoSubagentRunner struct{}

type demoRun struct {
	Report       ceo.Report
	WorkspaceDir string
}

func runDemo(ctx context.Context, out io.Writer, opts options) error {
	demo, err := runDemoReport(ctx, opts)
	if err != nil {
		return err
	}
	if err := writeRunReport(out, reportOutputRequest{
		Report:       demo.Report,
		Format:       opts.reportFormat,
		WorkspaceDir: demo.WorkspaceDir,
	}); err != nil {
		return err
	}
	return verdictError(demo.Report)
}

func runDemoReport(ctx context.Context, opts options) (demoRun, error) {
	root, err := os.MkdirTemp("", "ceo-harness-demo-*")
	if err != nil {
		return demoRun{}, fmt.Errorf("create demo workspace: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.txt"), []byte("hello old\n"), 0o644); err != nil {
		return demoRun{}, fmt.Errorf("write demo fixture: %w", err)
	}
	runtime := ceo.NewRuntimeWithSubagentRunner(demoSubagentRunner{})
	report, err := runtime.RunJob(ctx, ceo.JobRequest{
		Task:              fmt.Sprintf("%s (workspace: %s)", demoTaskPrefix, root),
		WorkspaceDir:      root,
		CheckCommand:      []string{"sh", "-c", `grep -q "hello new" "$CEO_DEMO_WORKSPACE/app.txt"`},
		CheckEnv:          []string{"CEO_DEMO_WORKSPACE=" + root},
		ApplyModelPatches: true,
		MaxModelPatches:   1,
		MaxContextBytes:   opts.maxContextBytes,
	})
	if err != nil {
		return demoRun{}, err
	}
	return demoRun{Report: report, WorkspaceDir: root}, nil
}

func (r demoSubagentRunner) Run(ctx context.Context, packet subagent.TaskPacket) (subagent.Result, error) {
	if err := ctx.Err(); err != nil {
		return subagent.Result{}, err
	}
	result := subagent.Result{
		AgentName:       packet.AgentName,
		Role:            packet.Role,
		Assignment:      strings.TrimSpace(packet.Assignment),
		Status:          "pass",
		Attempts:        1,
		ModelSource:     "demo",
		ContextReceived: packet.ContextMode,
		AllowedActions:  append([]string(nil), packet.AllowedActions...),
		ContextBytes:    len(packet.Task) + len(packet.WorkspaceBrief) + len(packet.PriorFindings),
		PromptBytes:     len(packet.Task),
		PriorFindings:   strings.TrimSpace(packet.PriorFindings),
		Summary:         demoSummary(packet.AgentName),
		Evidence:        []string{"demo runner executed"},
	}
	if packet.AgentName == "coder" {
		result.PatchProposals = []subagent.PatchProposal{
			{Path: "app.txt", Old: "old", New: "new"},
		}
		result.Evidence = append(result.Evidence, "proposed app.txt patch")
	}
	return result, nil
}

func demoSummary(agentName string) string {
	switch agentName {
	case "scanner":
		return "found demo app.txt fixture"
	case "coder":
		return "proposed bounded app.txt patch"
	case "reviewer":
		return "verified demo evidence path"
	default:
		return "completed demo task"
	}
}
