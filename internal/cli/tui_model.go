package cli

import "ceoharness/internal/history"

type tuiModel struct {
	workspace       string
	selected        int
	jobs            []tuiJob
	providerSummary history.ProviderHealthSummary
}

type tuiJob struct {
	id            string
	task          string
	verdict       string
	inboxReason   string
	action        string
	actionCommand string
	patchPreview  string
	checkOutput   string
	snapshotNote  string
}

func newTUIModel(workspace string, entries []history.Entry, inbox []reviewQueueRow, providers []history.ProviderHealth) tuiModel {
	if entries == nil {
		entries = []history.Entry{}
	}
	reasons := map[string]string{}
	for _, row := range inbox {
		reasons[row.ID] = row.ReviewReason
	}
	jobs := make([]tuiJob, 0, len(entries))
	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		reason := reasons[entry.ID]
		jobs = append(jobs, tuiJob{
			id:            entry.ID,
			task:          entry.Task,
			verdict:       entry.Verdict,
			inboxReason:   reason,
			action:        tuiActionLabel(reason),
			actionCommand: tuiActionCommand(workspace, entry.ID, reason),
		})
	}
	return tuiModel{
		workspace:       workspace,
		jobs:            jobs,
		providerSummary: history.SummarizeProviderHealth(providers),
	}
}

func (m tuiModel) applyKey(key string) (tuiModel, string) {
	switch key {
	case "down", "j":
		if m.selected+1 < len(m.jobs) {
			m.selected++
		}
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "enter", "a":
		return m, m.selectedAction()
	case "r":
		return m, m.selectedRetry()
	}
	return m, ""
}

func (m tuiModel) selectedAction() string {
	if len(m.jobs) == 0 || m.selected < 0 || m.selected >= len(m.jobs) {
		return ""
	}
	return m.jobs[m.selected].actionCommand
}

func (m tuiModel) selectedRetry() string {
	if len(m.jobs) == 0 || m.selected < 0 || m.selected >= len(m.jobs) {
		return ""
	}
	return tuiRetryCommand(m.workspace, m.jobs[m.selected].id)
}

func tuiActionLabel(reason string) string {
	switch reason {
	case "needs_input":
		return "answer"
	case "awaiting_human_judgment":
		return "accept"
	case "":
		return ""
	default:
		return "rerun"
	}
}

func tuiActionCommand(workspace string, jobID string, reason string) string {
	prefix := "cod --workspace " + workspaceArg(workspace)
	switch reason {
	case "needs_input":
		return prefix + " --resume " + jobID + " --answer \"...\""
	case "awaiting_human_judgment":
		return prefix + " --judge-job " + jobID + " --human-verdict accept"
	case "":
		return ""
	default:
		return prefix + " --rerun " + jobID
	}
}

func tuiRetryCommand(workspace string, jobID string) string {
	if jobID == "" {
		return ""
	}
	return "cod --workspace " + workspaceArg(workspace) + " --rerun " + jobID
}
