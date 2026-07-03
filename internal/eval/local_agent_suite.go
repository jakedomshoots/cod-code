package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func RunLocalAgentSuite(ctx context.Context, req LocalAgentSuiteRequest) (LocalAgentSuiteSummary, error) {
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return LocalAgentSuiteSummary{}, fmt.Errorf("create local agent suite output dir: %w", err)
	}
	task, err := localAgentTask(req.Task)
	if err != nil {
		return LocalAgentSuiteSummary{}, err
	}
	agentIDs := normalizeLocalAgentIDs(req.Agents)
	summary := LocalAgentSuiteSummary{
		SchemaVersion: localAgentSchemaVersion,
		Mode:          localAgentSuiteMode,
		Task:          task.name,
		Prompt:        task.prompt,
		AgentCount:    len(agentIDs),
		Results:       make([]LocalAgentResult, 0, len(agentIDs)),
	}
	for _, agentID := range agentIDs {
		spec, err := buildLocalAgentSpec(agentID, req.CEOHarnessBinary, task)
		if err != nil {
			return LocalAgentSuiteSummary{}, err
		}
		result := runLocalAgent(ctx, req, spec)
		summary.Results = append(summary.Results, result)
		accumulateLocalAgentStatus(&summary, result.Status)
	}
	summary.IterationBacklog = buildLocalAgentIterations(summary.Task, summary.Results)
	if err := writeJSONFile(filepath.Join(req.OutputDir, "summary.json"), summary); err != nil {
		return LocalAgentSuiteSummary{}, err
	}
	if err := writeLocalAgentMarkdown(filepath.Join(req.OutputDir, "summary.md"), summary); err != nil {
		return LocalAgentSuiteSummary{}, err
	}
	if err := writeLocalAgentBacklog(filepath.Join(req.OutputDir, "iteration-backlog.md"), summary.IterationBacklog); err != nil {
		return LocalAgentSuiteSummary{}, err
	}
	return summary, nil
}

func runLocalAgent(ctx context.Context, req LocalAgentSuiteRequest, spec localAgentSpec) LocalAgentResult {
	resultDir := filepath.Join(req.OutputDir, spec.id)
	workspaceDir := filepath.Join(resultDir, "workspace")
	result := LocalAgentResult{
		ID:           spec.id,
		Name:         spec.name,
		Status:       localAgentStatusSkipped,
		Binary:       spec.binary,
		WorkspaceDir: workspaceDir,
		ExitCode:     -1,
		SetupHint:    spec.setupHint,
		Note:         "binary not found on PATH; skipped instead of failed",
	}
	if err := prepareLocalAgentWorkspace(workspaceDir); err != nil {
		result.Status = localAgentStatusFail
		result.Error = err.Error()
		result.Note = "failed to prepare isolated workspace"
		return result
	}
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		result.Status = localAgentStatusFail
		result.Error = fmt.Errorf("resolve local agent workspace: %w", err).Error()
		result.Note = "failed to resolve isolated workspace"
		return result
	}
	result.WorkspaceDir = absWorkspace
	resolved, err := resolveLocalAgentBinary(spec.binary)
	if err != nil {
		return result
	}
	result.ResolvedPath = resolved
	result.SetupHint = ""
	command := localAgentCommand(resolved, spec.args, absWorkspace)
	result.Command = command
	result.CommandPath = filepath.Join(resultDir, "command.json")
	result.StdoutPath = filepath.Join(resultDir, "stdout.log")
	result.StderrPath = filepath.Join(resultDir, "stderr.log")
	result.AppAfterPath = filepath.Join(resultDir, "app-after.txt")
	run := runLocalAgentCommand(ctx, command, absWorkspace, spec.env, localAgentTimeout(req.TimeoutSeconds))
	result.ExitCode = run.exitCode
	result.DurationMS = run.duration.Milliseconds()
	result.Error = run.errText
	result.OutputMatched = outputMatches(run.stdout+run.stderr, spec.expectedOutput)
	result.ObservedFile, result.FileMatched = observedFileMatch(absWorkspace, spec.expectedFile)
	result.Status = localAgentStatus(run, result.OutputMatched, result.FileMatched)
	result.Note = localAgentNote(result.Status, spec.expectedOutput)
	if err := writeLocalAgentEvidence(result, run); err != nil {
		result.Status = localAgentStatusFail
		result.Error = err.Error()
		result.Note = "failed to save local agent evidence"
	}
	return result
}
