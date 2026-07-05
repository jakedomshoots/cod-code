package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	localAgentBenchmarkMode      = "local_agent_benchmark"
	defaultLocalAgentBenchmarkID = "docs-roadmap-cli-first"
	localAgentStatusPartial      = "partial"
	localAgentStatusSetupBlocked = "setup_blocked"
	localAgentEvidenceComplete   = "complete"
	localAgentEvidenceIncomplete = "incomplete"
	localAgentEvidenceNotRun     = "not_run"
)

func RunLocalAgentBenchmark(ctx context.Context, req LocalAgentBenchmarkRequest) (LocalAgentBenchmarkSummary, error) {
	tasks, err := localAgentBenchmarkTasks(ctx, req)
	if err != nil {
		return LocalAgentBenchmarkSummary{}, err
	}
	if len(tasks) == 0 {
		return LocalAgentBenchmarkSummary{}, fmt.Errorf("%w: no benchmark tasks selected", ErrInvalidTask)
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return LocalAgentBenchmarkSummary{}, fmt.Errorf("create local agent benchmark output dir: %w", err)
	}
	agentIDs := normalizeLocalAgentIDs(req.Agents)
	repeatCount := normalizeLocalAgentBenchmarkRepeat(req.RepeatCount)
	concurrency := normalizeLocalAgentBenchmarkConcurrency(req.Concurrency)
	timeoutRetries := normalizeLocalAgentBenchmarkTimeoutRetries(req.TimeoutRetries)
	resultRetries := normalizeLocalAgentBenchmarkResultRetries(req.ResultRetries)
	agentTimeouts := normalizeLocalAgentBenchmarkAgentTimeouts(req.AgentTimeoutSeconds)
	agentModels := normalizeLocalAgentBenchmarkAgentModels(req.AgentModels)
	multiRun := len(tasks) > 1 || repeatCount > 1
	summary := LocalAgentBenchmarkSummary{
		SchemaVersion:  localAgentSchemaVersion,
		Mode:           localAgentBenchmarkMode,
		TaskID:         tasks[0].ID,
		TaskTitle:      tasks[0].Title,
		TaskIDs:        localAgentBenchmarkTaskIDs(tasks),
		TaskCount:      len(tasks),
		RepeatCount:    repeatCount,
		Concurrency:    concurrency,
		TimeoutRetries: timeoutRetries,
		ResultRetries:  resultRetries,
		AgentTimeouts:  agentTimeouts,
		AgentModels:    agentModels,
		RunCount:       len(tasks) * repeatCount * len(agentIDs),
		AgentCount:     len(agentIDs),
		Results:        make([]LocalAgentBenchmarkResult, 0, len(tasks)*repeatCount*len(agentIDs)),
	}
	jobs, err := buildLocalAgentBenchmarkJobs(tasks, agentIDs, req, repeatCount, multiRun)
	if err != nil {
		return LocalAgentBenchmarkSummary{}, err
	}
	if concurrency > len(jobs) {
		concurrency = len(jobs)
		summary.Concurrency = concurrency
	}
	if concurrency <= 1 {
		for _, job := range jobs {
			result := runLocalAgentBenchmark(ctx, req, job.task, job.spec, job.attempt, job.multiRun)
			summary.Results = append(summary.Results, result)
			accumulateLocalAgentBenchmarkStatus(&summary, result.Status)
			accumulateLocalAgentBenchmarkEvidence(&summary, result.EvidenceStatus)
			summary.IterationBacklog = buildLocalAgentBenchmarkIterations(summary.Results)
			if err := writeLocalAgentBenchmarkSummaryArtifacts(req.OutputDir, summary); err != nil {
				return LocalAgentBenchmarkSummary{}, err
			}
		}
		summary.IterationBacklog = buildLocalAgentBenchmarkIterations(summary.Results)
		if err := writeLocalAgentBenchmarkSummaryArtifacts(req.OutputDir, summary); err != nil {
			return LocalAgentBenchmarkSummary{}, err
		}
		return summary, nil
	}
	return runLocalAgentBenchmarkParallel(ctx, req, summary, jobs, concurrency)
}

func runLocalAgentBenchmark(ctx context.Context, req LocalAgentBenchmarkRequest, task Task, spec localAgentSpec, attempt int, multiRun bool) LocalAgentBenchmarkResult {
	maxRunAttempts := maxLocalAgentBenchmarkAttempts(req)
	prior := make([]RetryAttempt, 0, maxRunAttempts-1)
	for runAttempt := 1; runAttempt <= maxRunAttempts; runAttempt++ {
		result := runLocalAgentBenchmarkAttempt(ctx, req, task, spec, attempt, multiRun, runAttempt, maxRunAttempts)
		result.PriorAttempts = append([]RetryAttempt(nil), prior...)
		if !shouldRetryLocalAgentBenchmarkResult(req, result, runAttempt) {
			if runAttempt > 1 && result.Status == localAgentStatusPass {
				result.Note = fmt.Sprintf("%s after %d prior non-pass attempt(s)", result.Note, len(prior))
			}
			return result
		}
		prior = append(prior, retryAttemptFromResult(result))
	}
	return LocalAgentBenchmarkResult{}
}

func maxLocalAgentBenchmarkAttempts(req LocalAgentBenchmarkRequest) int {
	timeoutAttempts := normalizeLocalAgentBenchmarkTimeoutRetries(req.TimeoutRetries) + 1
	resultAttempts := normalizeLocalAgentBenchmarkResultRetries(req.ResultRetries) + 1
	if resultAttempts > timeoutAttempts {
		return resultAttempts
	}
	return timeoutAttempts
}

func shouldRetryLocalAgentBenchmarkResult(req LocalAgentBenchmarkRequest, result LocalAgentBenchmarkResult, runAttempt int) bool {
	switch result.Status {
	case localAgentStatusTimeout:
		return runAttempt <= normalizeLocalAgentBenchmarkTimeoutRetries(req.TimeoutRetries) ||
			runAttempt <= normalizeLocalAgentBenchmarkResultRetries(req.ResultRetries)
	case localAgentStatusPartial, localAgentStatusFail:
		return runAttempt <= normalizeLocalAgentBenchmarkResultRetries(req.ResultRetries)
	default:
		return false
	}
}

func runLocalAgentBenchmarkAttempt(ctx context.Context, req LocalAgentBenchmarkRequest, task Task, spec localAgentSpec, attempt int, multiRun bool, runAttempt int, maxRunAttempts int) LocalAgentBenchmarkResult {
	resultDir := localAgentBenchmarkAttemptResultDir(req.OutputDir, task, spec, attempt, multiRun, runAttempt, maxRunAttempts)
	workspaceDir := filepath.Join(resultDir, "workspace")
	result := newLocalAgentBenchmarkResult(spec, task, attempt, workspaceDir, resultDir)
	result.RunAttempt = runAttempt
	if err := prepareLocalAgentBenchmarkWorkspace(ctx, workspaceDir, task, spec.workspaceConfig); err != nil {
		result.Status = localAgentStatusFail
		result.Error = err.Error()
		result.Note = "failed to prepare benchmark workspace"
		result.EvidenceStatus = localAgentEvidenceIncomplete
		result.ScoreVerdict = "fail"
		if writeErr := writeBenchmarkSetupFailureEvidence(result); writeErr != nil {
			result.Error = fmt.Sprintf("%s; evidence write failed: %v", result.Error, writeErr)
		}
		return result
	}
	absWorkspace, err := filepath.Abs(workspaceDir)
	if err != nil {
		result.Status = localAgentStatusFail
		result.Error = fmt.Errorf("resolve benchmark workspace: %w", err).Error()
		result.Note = "failed to resolve benchmark workspace"
		return result
	}
	result.WorkspaceDir = absWorkspace
	resolved, err := resolveLocalAgentBinary(spec.binary)
	if err != nil {
		result.Error = err.Error()
		result.EvidenceStatus = localAgentEvidenceIncomplete
		if writeErr := writeBenchmarkTerminalEvidence(task, result, "agent binary was not found"); writeErr != nil {
			result.Error = fmt.Sprintf("%s; evidence write failed: %v", result.Error, writeErr)
		}
		return result
	}
	result.ResolvedPath = resolved
	result.SetupHint = ""
	args, err := localAgentBenchmarkArgs(spec, resultDir)
	if err != nil {
		return benchmarkResultError(task, result, "failed to resolve benchmark agent args", err)
	}
	result.Command = localAgentCommand(resolved, args, absWorkspace)
	if err := writeBenchmarkStatusFile(ctx, result.GitBeforePath, absWorkspace); err != nil {
		return benchmarkResultError(task, result, "failed to capture pre-run git status", err)
	}
	run := runLocalAgentCommand(ctx, result.Command, absWorkspace, spec.env, localAgentBenchmarkTimeout(req, spec))
	result.ExitCode = run.exitCode
	result.DurationMS = run.duration.Milliseconds()
	result.Error = run.errText
	if err := writeBenchmarkRunEvidence(result, run); err != nil {
		return benchmarkResultError(task, result, "failed to save benchmark command evidence", err)
	}
	if spec.benchmarkWritesArtifacts {
		content := benchmarkEvidenceContent(task, result, nil)
		if err := writeBenchmarkArtifactsInWorkspace(absWorkspace, task, content); err != nil {
			return benchmarkResultError(task, result, "failed to write synthetic benchmark evidence", err)
		}
	}
	score, changedFiles, err := scoreLocalAgentBenchmark(ctx, task, result)
	if err != nil {
		return benchmarkResultError(task, result, "failed to score benchmark result", err)
	}
	result.ChangedFiles = changedFiles
	result.ExtraChangedFiles = benchmarkExtraChangedFiles(task, changedFiles)
	result.ScoreVerdict = score.Verdict
	result.PassedChecks = score.Passed
	result.TotalChecks = score.Total
	result.FailedScoreChecks = failedScoreChecks(score)
	result.Status = localAgentBenchmarkStatus(run, score.Verdict)
	result.EvidenceStatus = localAgentBenchmarkEvidenceStatus(result.Status, result.FailedScoreChecks)
	result.Note = localAgentBenchmarkNote(result.Status)
	return result
}

func localAgentBenchmarkTimeout(req LocalAgentBenchmarkRequest, spec localAgentSpec) time.Duration {
	if seconds, ok := req.AgentTimeoutSeconds[spec.id]; ok && seconds > 0 {
		return localAgentTimeout(seconds)
	}
	return localAgentTimeout(req.TimeoutSeconds)
}

func localAgentBenchmarkAttemptResultDir(outputDir string, task Task, spec localAgentSpec, attempt int, multiRun bool, runAttempt int, maxRunAttempts int) string {
	base := localAgentBenchmarkResultDir(outputDir, task, spec, attempt, multiRun)
	if maxRunAttempts <= 1 {
		return base
	}
	return filepath.Join(base, fmt.Sprintf("attempt-%02d", runAttempt))
}
