package ceo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ceoharness/internal/checkrunner"
	"ceoharness/internal/jobpacket"
	"ceoharness/internal/subagent"
	"ceoharness/internal/workspace"
)

const (
	maxCheckFixOutputBytes = 800
	checkFixRole           = "fix failed verification with minimal patch"
)

type checkFixState struct {
	Packet          jobpacket.Packet
	Request         JobRequest
	Space           workspace.Workspace
	HasWorkspace    bool
	ArtifactStore   runtimeArtifactStore
	Results         []subagent.Result
	ChangedFiles    []string
	CheckResults    []checkrunner.Result
	PatchResults    []workspace.ReplaceTextResult
	PatchAudit      []PatchAuditEntry
	WorkspaceBrief  string
	IterationBudget *ceoIterationBudget
	RetryHistory    []RetryHistoryEntry
}

type checkFixRequest struct {
	Packet         jobpacket.Packet
	Request        JobRequest
	Space          workspace.Workspace
	HasWorkspace   bool
	Checks         []checkrunner.Result
	Attempt        int
	WorkspaceBrief string
	FailedChecks   []RepairFailureDetail
}

func (r Runtime) runCheckFixLoop(ctx context.Context, state checkFixState) (checkFixState, error) {
	if state.Request.CheckFixAttempts < 1 || lastCheckPassed(state.CheckResults) {
		return state, nil
	}
	if !state.HasWorkspace {
		return state, fmt.Errorf("workspace is required for check-fix attempts")
	}
	if !state.Request.ApplyModelPatches {
		return state, fmt.Errorf("check-fix attempts require model patch application")
	}

	noProgress := newNoProgressTracker(state.Request.NoProgressStop)
	for attempt := 1; attempt <= state.Request.CheckFixAttempts && !lastCheckPassed(state.CheckResults); attempt++ {
		if !consumeCEOIterationBudget(state.IterationBudget) {
			return state, nil
		}
		failedChecks := repairFailureDetails(state.CheckResults, state.Request.ScorerFailedChecks)
		prompt := buildCheckFixTask(state.Packet.Task, state.CheckResults, attempt, failedChecks)
		history := RetryHistoryEntry{
			Kind:             "check_fix",
			Attempt:          attempt,
			Status:           "fail",
			FailedChecks:     failedChecks,
			CorrectivePrompt: prompt,
		}
		result, err := r.runCoderCheckFix(ctx, checkFixRequest{
			Packet:         state.Packet,
			Request:        state.Request,
			Space:          state.Space,
			HasWorkspace:   state.HasWorkspace,
			Checks:         state.CheckResults,
			Attempt:        attempt,
			WorkspaceBrief: state.WorkspaceBrief,
			FailedChecks:   failedChecks,
		})
		if err != nil {
			return state, err
		}
		result = compactSubagentOutput(result, state.Request.MaxSubagentOutputBytes)
		state.Results = append(state.Results, result)
		path, err := writeSubagentEvidenceFile(ctx, state.ArtifactStore.Space, fmt.Sprintf("coder-fix-%d", attempt), result)
		if err != nil {
			return state, err
		}
		if changedPath, ok := state.ArtifactStore.changedPath(path); ok {
			state.ChangedFiles = append(state.ChangedFiles, changedPath)
		}

		patches, err := coderPatchProposals(result)
		if err != nil {
			history.ModelPatchStatus = "invalid"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("invalid:" + err.Error()) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		if len(patches) == 0 {
			history.ModelPatchStatus = "empty"
			history.Reason = "model returned no patch proposals"
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("empty") {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		if patchesAreNoOp(patches) {
			history.ModelPatchStatus = "no_op"
			history.Reason = "model patch does not change content"
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe("no_op:" + patchSignature(patches)) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		failedPatchSignature := failedPatchNoProgressSignature(patches)
		if noProgress.hasObserved(failedPatchSignature) && noProgress.observe(failedPatchSignature) {
			history.ModelPatchStatus = "identical"
			history.Reason = "model repeated an identical failed patch"
			state.RetryHistory = append(state.RetryHistory, history)
			markLastRetryNoProgress(state.Results, state.RetryHistory)
			break
		}
		if err := enforceModelPatchLimit(patches, state.Request.MaxModelPatches); err != nil {
			history.ModelPatchStatus = "invalid"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			continue
		}
		applied, err := applyPatchRequests(ctx, state.Space, patches)
		if err != nil {
			history.ModelPatchStatus = "apply_failed"
			history.Reason = err.Error()
			state.RetryHistory = append(state.RetryHistory, history)
			if noProgress.observe(applyFailureNoProgressSignature(patches, err)) {
				markLastRetryNoProgress(state.Results, state.RetryHistory)
				break
			}
			continue
		}
		history.ModelPatchStatus = "applied"
		history.ChangedFiles = changedFilesFromPatchResults(applied)
		state.PatchResults = append(state.PatchResults, applied...)
		state.PatchAudit = append(state.PatchAudit, patchAuditEntries(applied, "model", "coder")...)
		state.ChangedFiles = appendChangedPatchFiles(state.ChangedFiles, applied)

		nextChecks, err := r.runChecks(ctx, state.Request)
		if err != nil {
			return state, err
		}
		state.CheckResults = append(state.CheckResults, nextChecks...)
		if lastCheckPassed(state.CheckResults) {
			history.Status = "pass"
			history.FinalVerdict = "pass"
		} else {
			history.FinalVerdict = "fail"
		}
		state.RetryHistory = append(state.RetryHistory, history)
		if history.FinalVerdict != "pass" && noProgress.observe(failedPatchSignature) {
			markLastRetryNoProgress(state.Results, state.RetryHistory)
			break
		}
	}
	return state, nil
}

func (r Runtime) runCoderCheckFix(ctx context.Context, req checkFixRequest) (subagent.Result, error) {
	packet := req.Packet
	packet.Task = buildCheckFixTask(req.Packet.Task, req.Checks, req.Attempt, req.FailedChecks)
	toolResults := checkFixRequiredFileToolResults(ctx, req)
	allowedActions := jobpacket.DefaultActionsForAgent("coder")
	if len(toolResults) > 0 {
		allowedActions = []jobpacket.Action{jobpacket.ActionProposePatch}
	}
	agent := jobpacket.Subagent{
		Name:           "coder",
		Role:           checkFixRole,
		AllowedActions: allowedActions,
	}
	attempts := req.Request.SubagentAttempts
	if attempts < 1 {
		attempts = 1
	}
	backoff := time.Duration(req.Request.SubagentBackoffMS) * time.Millisecond
	return r.runSubagentWithAttempts(ctx, subagentRunInput{
		Packet:         packet,
		Agent:          agent,
		Attempts:       attempts,
		Backoff:        backoff,
		NoProgressStop: req.Request.NoProgressStop,
		WorkspaceBrief: req.WorkspaceBrief,
		ToolResults:    toolResults,
	})
}

func checkFixRequiredFileToolResults(ctx context.Context, req checkFixRequest) []subagent.ToolResult {
	if !req.HasWorkspace {
		return nil
	}
	paths := checkFixContextPaths(req)
	results := make([]subagent.ToolResult, 0, len(paths))
	for _, path := range paths {
		result := subagent.ToolResult{
			Action: "read_workspace",
			Path:   path,
		}
		read, err := req.Space.ReadText(ctx, workspace.ReadTextRequest{
			Path: path,
		})
		if err != nil {
			result.Status = "fail"
			result.Error = err.Error()
		} else {
			result.Status = "pass"
			result.Path = read.Path
			result.Output = read.Content
			result.Bytes = read.Bytes
			result.Truncated = read.Truncated
		}
		results = append(results, result)
	}
	return results
}

func checkFixContextPaths(req checkFixRequest) []string {
	seen := map[string]struct{}{}
	var paths []string
	add := func(path string) {
		clean := cleanCheckFixPathToken(path)
		if clean == "" {
			return
		}
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		paths = append(paths, clean)
	}
	for _, path := range requiredChangedFilePaths(req.Packet.Task) {
		add(path)
	}
	for _, check := range req.Checks {
		for _, arg := range check.Argv {
			for _, field := range strings.Fields(arg) {
				add(field)
			}
		}
	}
	return paths
}

func cleanCheckFixPathToken(token string) string {
	clean := strings.Trim(token, " \t\n\r\"'`()[]{}.,;")
	if clean == "" || strings.HasPrefix(clean, "-") {
		return ""
	}
	if index := strings.Index(clean, ":"); index > 0 {
		clean = clean[:index]
	}
	if marker := "/workspace/"; strings.Contains(clean, marker) {
		clean = clean[strings.LastIndex(clean, marker)+len(marker):]
	}
	if !strings.Contains(clean, "/") {
		return ""
	}
	switch {
	case strings.HasSuffix(clean, ".go"),
		strings.HasSuffix(clean, ".js"),
		strings.HasSuffix(clean, ".ts"),
		strings.HasSuffix(clean, ".tsx"),
		strings.HasSuffix(clean, ".jsx"),
		strings.HasSuffix(clean, ".py"),
		strings.HasSuffix(clean, ".rs"),
		strings.HasSuffix(clean, ".md"),
		strings.HasSuffix(clean, ".json"),
		strings.HasSuffix(clean, ".yaml"),
		strings.HasSuffix(clean, ".yml"):
		return clean
	default:
		return ""
	}
}

func buildCheckFixTask(task string, checks []checkrunner.Result, attempt int, failureDetails ...[]RepairFailureDetail) string {
	if len(checks) == 0 {
		return task
	}
	last := checks[len(checks)-1]
	details := []RepairFailureDetail(nil)
	if len(failureDetails) > 0 {
		details = failureDetails[0]
	}
	scorerDetails := renderRepairFailureDetails(details)
	if scorerDetails != "" {
		scorerDetails = "\n" + scorerDetails
	}
	return fmt.Sprintf(
		"%s\n\nVerification failed. Fix attempt %d.\n%s%s\nStdout:\n%s\nStderr:\n%s\nUse supplied tool_results as the file context. Do not request tools or more reads. Return only coder patch JSON.",
		task,
		attempt,
		renderCheckFixMetadata(last),
		scorerDetails,
		trimCheckFixOutput(last.Stdout),
		trimCheckFixOutput(last.Stderr),
	)
}

func lastCheckPassed(results []checkrunner.Result) bool {
	return len(results) == 0 || results[len(results)-1].Status == "pass"
}

func trimCheckFixOutput(output string) string {
	cleanOutput := strings.TrimSpace(output)
	if len(cleanOutput) <= maxCheckFixOutputBytes {
		return cleanOutput
	}
	return cleanOutput[:maxCheckFixOutputBytes] + "\n[truncated]"
}
