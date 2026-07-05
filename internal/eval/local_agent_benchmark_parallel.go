package eval

import (
	"context"
	"sort"
	"sync"
)

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
	summary.SetupBlocked = 0
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

func normalizeLocalAgentBenchmarkTimeoutRetries(retries int) int {
	if retries < 0 {
		return 0
	}
	return retries
}

func normalizeLocalAgentBenchmarkResultRetries(retries int) int {
	if retries < 0 {
		return 0
	}
	return retries
}

func normalizeLocalAgentBenchmarkAgentTimeouts(raw map[string]int) map[string]int {
	clean := make(map[string]int)
	for agent, seconds := range raw {
		if seconds > 0 {
			clean[agent] = seconds
		}
	}
	if len(clean) == 0 {
		return nil
	}
	return clean
}
