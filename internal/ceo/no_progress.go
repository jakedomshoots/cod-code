package ceo

import (
	"strings"

	"ceoharness/internal/subagent"
)

const DefaultNoProgressStop = 2

type noProgressTracker struct {
	threshold int
	last      string
	repeats   int
}

func normalizeNoProgressStop(value int) int {
	if value <= 0 {
		return DefaultNoProgressStop
	}
	return value
}

func newNoProgressTracker(threshold int) noProgressTracker {
	return noProgressTracker{threshold: normalizeNoProgressStop(threshold)}
}

func (t *noProgressTracker) observeResult(result subagent.Result) bool {
	signature, ok := noProgressResultSignature(result)
	if !ok {
		t.last = ""
		t.repeats = 0
		return false
	}
	return t.observe(signature)
}

func (t *noProgressTracker) observeError(message string) bool {
	clean := strings.TrimSpace(message)
	if clean == "" {
		clean = "subagent error"
	}
	return t.observe("error:" + clean)
}

func (t *noProgressTracker) observe(signature string) bool {
	if signature != t.last {
		t.last = signature
		t.repeats = 1
		return false
	}
	t.repeats++
	return t.threshold > 0 && t.repeats >= t.threshold
}

func (t noProgressTracker) hasObserved(signature string) bool {
	return signature != "" && signature == t.last
}

func noProgressResultSignature(result subagent.Result) (string, bool) {
	if result.Status == "pass" {
		return "", false
	}
	if len(result.PatchProposals) > 0 || len(result.ToolRequests) > 0 || len(result.Questions) > 0 {
		return "", false
	}
	status := strings.TrimSpace(result.Status)
	summary := strings.TrimSpace(result.Summary)
	if status == "" && summary == "" {
		return "", false
	}
	return "result:" + status + "\x00" + summary, true
}

func markNoProgressStopped(result subagent.Result) subagent.Result {
	result.NoProgressStopped = true
	result.Evidence = append(result.Evidence, "no progress stop reached")
	return result
}
