package eval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const (
	localAgentBenchmarkMode      = "local_agent_benchmark"
	defaultLocalAgentBenchmarkID = "docs-roadmap-cli-first"
	localAgentStatusPartial      = "partial"
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
	multiRun := len(tasks) > 1 || repeatCount > 1
	summary := LocalAgentBenchmarkSummary{
		SchemaVersion: localAgentSchemaVersion,
		Mode:          localAgentBenchmarkMode,
		TaskID:        tasks[0].ID,
		TaskTitle:     tasks[0].Title,
		TaskIDs:       localAgentBenchmarkTaskIDs(tasks),
		TaskCount:     len(tasks),
		RepeatCount:   repeatCount,
		Concurrency:   concurrency,
		RunCount:      len(tasks) * repeatCount * len(agentIDs),
		AgentCount:    len(agentIDs),
		Results:       make([]LocalAgentBenchmarkResult, 0, len(tasks)*repeatCount*len(agentIDs)),
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

type localAgentBenchmarkJob struct {
	index    int
	task     Task
	spec     localAgentSpec
	attempt  int
	multiRun bool
}

type localAgentBenchmarkJobResult struct {
	index  int
	result LocalAgentBenchmarkResult
}

func buildLocalAgentBenchmarkJobs(tasks []Task, agentIDs []string, req LocalAgentBenchmarkRequest, repeatCount int, multiRun bool) ([]localAgentBenchmarkJob, error) {
	jobs := make([]localAgentBenchmarkJob, 0, len(tasks)*repeatCount*len(agentIDs))
	for attempt := 1; attempt <= repeatCount; attempt++ {
		for _, task := range tasks {
			for _, agentID := range agentIDs {
				spec, err := buildLocalAgentBenchmarkSpec(agentID, req, task)
				if err != nil {
					return nil, err
				}
				jobs = append(jobs, localAgentBenchmarkJob{
					index:    len(jobs),
					task:     task,
					spec:     spec,
					attempt:  attempt,
					multiRun: multiRun,
				})
			}
		}
	}
	return jobs, nil
}

func runLocalAgentBenchmarkParallel(ctx context.Context, req LocalAgentBenchmarkRequest, summary LocalAgentBenchmarkSummary, jobs []localAgentBenchmarkJob, concurrency int) (LocalAgentBenchmarkSummary, error) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobCh := make(chan localAgentBenchmarkJob)
	resultCh := make(chan localAgentBenchmarkJobResult)
	var workers sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for job := range jobCh {
				result := runLocalAgentBenchmark(runCtx, req, job.task, job.spec, job.attempt, job.multiRun)
				select {
				case resultCh <- localAgentBenchmarkJobResult{index: job.index, result: result}:
				case <-runCtx.Done():
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobCh)
		for _, job := range jobs {
			select {
			case jobCh <- job:
			case <-runCtx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(resultCh)
	}()

	completed := make([]localAgentBenchmarkJobResult, 0, len(jobs))
	for jobResult := range resultCh {
		completed = append(completed, jobResult)
		summary.Results = orderedLocalAgentBenchmarkResults(completed)
		recountLocalAgentBenchmarkSummary(&summary)
		if err := writeLocalAgentBenchmarkSummaryArtifacts(req.OutputDir, summary); err != nil {
			cancel()
			return LocalAgentBenchmarkSummary{}, err
		}
	}
	if err := runCtx.Err(); err != nil && ctx.Err() != nil {
		return LocalAgentBenchmarkSummary{}, err
	}
	summary.Results = orderedLocalAgentBenchmarkResults(completed)
	recountLocalAgentBenchmarkSummary(&summary)
	if err := writeLocalAgentBenchmarkSummaryArtifacts(req.OutputDir, summary); err != nil {
		return LocalAgentBenchmarkSummary{}, err
	}
	return summary, nil
}

func orderedLocalAgentBenchmarkResults(completed []localAgentBenchmarkJobResult) []LocalAgentBenchmarkResult {
	ordered := append([]localAgentBenchmarkJobResult(nil), completed...)
	sort.SliceStable(ordered, func(left int, right int) bool {
		return ordered[left].index < ordered[right].index
	})
	results := make([]LocalAgentBenchmarkResult, 0, len(ordered))
	for _, item := range ordered {
		results = append(results, item.result)
	}
	return results
}

func recountLocalAgentBenchmarkSummary(summary *LocalAgentBenchmarkSummary) {
	summary.Passed = 0
	summary.Partial = 0
	summary.Failed = 0
	summary.TimedOut = 0
	summary.Skipped = 0
	summary.IncompleteEvidence = 0
	for _, result := range summary.Results {
		accumulateLocalAgentBenchmarkStatus(summary, result.Status)
		accumulateLocalAgentBenchmarkEvidence(summary, result.EvidenceStatus)
	}
	summary.IterationBacklog = buildLocalAgentBenchmarkIterations(summary.Results)
}

func normalizeLocalAgentBenchmarkConcurrency(concurrency int) int {
	if concurrency <= 0 {
		return 1
	}
	return concurrency
}

func runLocalAgentBenchmark(ctx context.Context, req LocalAgentBenchmarkRequest, task Task, spec localAgentSpec, attempt int, multiRun bool) LocalAgentBenchmarkResult {
	resultDir := localAgentBenchmarkResultDir(req.OutputDir, task, spec, attempt, multiRun)
	workspaceDir := filepath.Join(resultDir, "workspace")
	result := newLocalAgentBenchmarkResult(spec, task, attempt, workspaceDir, resultDir)
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
	run := runLocalAgentCommand(ctx, result.Command, absWorkspace, spec.env, localAgentTimeout(req.TimeoutSeconds))
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

func newLocalAgentBenchmarkResult(spec localAgentSpec, task Task, attempt int, workspaceDir string, resultDir string) LocalAgentBenchmarkResult {
	return LocalAgentBenchmarkResult{
		ID:               spec.id,
		Name:             spec.name,
		TaskID:           task.ID,
		TaskTitle:        task.Title,
		Attempt:          attempt,
		Status:           localAgentStatusSkipped,
		Binary:           spec.binary,
		WorkspaceDir:     workspaceDir,
		ExitCode:         -1,
		CommandPath:      filepath.Join(resultDir, "command.json"),
		StdoutPath:       filepath.Join(resultDir, "stdout.log"),
		StderrPath:       filepath.Join(resultDir, "stderr.log"),
		ReportPath:       filepath.Join(resultDir, "report.json"),
		ScorePath:        filepath.Join(resultDir, "score.json"),
		DiffPath:         filepath.Join(resultDir, "diff.patch"),
		ChangedFilesPath: filepath.Join(resultDir, "changed-files.txt"),
		GitBeforePath:    filepath.Join(resultDir, "git-status-before.txt"),
		GitAfterPath:     filepath.Join(resultDir, "git-status-after.txt"),
		TimingPath:       filepath.Join(resultDir, "timing.txt"),
		EvidenceStatus:   localAgentEvidenceNotRun,
		SetupHint:        spec.setupHint,
		Note:             "binary not found on PATH; skipped instead of failed",
	}
}

func benchmarkResultError(task Task, result LocalAgentBenchmarkResult, note string, err error) LocalAgentBenchmarkResult {
	result.Status = localAgentStatusFail
	result.Error = err.Error()
	result.Note = note
	result.EvidenceStatus = localAgentEvidenceIncomplete
	if writeErr := writeBenchmarkTerminalEvidence(task, result, note); writeErr != nil {
		result.Error = fmt.Sprintf("%s; evidence write failed: %v", result.Error, writeErr)
	}
	return result
}

func writeBenchmarkSetupFailureEvidence(result LocalAgentBenchmarkResult) error {
	check := CheckResult{
		Name:     "setup:prepare_workspace",
		Status:   "fail",
		Evidence: result.WorkspaceDir,
		Message:  result.Error,
	}
	score := ScoreResult{
		TaskID:  result.TaskID,
		Verdict: "fail",
		Passed:  0,
		Total:   1,
		Checks:  []CheckResult{check},
	}
	report := map[string]any{
		"task_id":         result.TaskID,
		"agent_id":        result.ID,
		"status":          result.Status,
		"evidence_status": result.EvidenceStatus,
		"error":           result.Error,
		"note":            result.Note,
		"workspace_dir":   result.WorkspaceDir,
	}
	if err := writeJSONFile(result.CommandPath, result.Command); err != nil {
		return err
	}
	if err := writeTextFile(result.StdoutPath, "workspace setup failed before agent command\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.StderrPath, result.Error+"\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.DiffPath, "workspace setup failed before diff\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.ChangedFilesPath, "workspace setup failed before changed-file scan\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitBeforePath, "workspace setup failed before git status\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.GitAfterPath, "workspace setup failed before git status\n"); err != nil {
		return err
	}
	if err := writeTextFile(result.TimingPath, "duration_ms=0\n"); err != nil {
		return err
	}
	if err := writeJSONFile(result.ReportPath, report); err != nil {
		return err
	}
	return writeJSONFile(result.ScorePath, score)
}

func failedScoreChecks(score ScoreResult) []CheckResult {
	failed := make([]CheckResult, 0)
	for _, check := range score.Checks {
		if check.Status != "pass" {
			failed = append(failed, check)
		}
	}
	return failed
}
