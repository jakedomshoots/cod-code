package cli

import (
	"fmt"
	"strings"

	"ceoharness/internal/history"
)

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
	case "enter":
		return m, m.selectedAction()
	}
	return m, ""
}

func (m tuiModel) selectedAction() string {
	if len(m.jobs) == 0 || m.selected < 0 || m.selected >= len(m.jobs) {
		return ""
	}
	return m.jobs[m.selected].actionCommand
}

func (m tuiModel) render() string {
	var builder strings.Builder
	builder.WriteString("CEO Harness TUI\n")
	builder.WriteString("CEO Harness Dashboard\n")
	builder.WriteString("Workspace: ")
	builder.WriteString(m.workspace)
	builder.WriteString("\n")
	if len(m.jobs) == 0 {
		writeTUIEmpty(&builder, m.workspace)
		return builder.String()
	}
	builder.WriteString(fmt.Sprintf("Inbox: %d\n", m.inboxCount()))
	builder.WriteString(fmt.Sprintf("Jobs (%d)\n", len(m.jobs)))
	for index, job := range m.jobs {
		prefix := "  "
		if index == m.selected {
			prefix = "> "
		}
		builder.WriteString(prefix)
		builder.WriteString(job.id)
		builder.WriteString(" [")
		builder.WriteString(job.verdict)
		builder.WriteString("] ")
		builder.WriteString(trimText(oneLine(job.task), 72))
		builder.WriteString("\n")
	}
	writeTUISelectedJob(&builder, m.selectedJob())
	writeTUIProviderHealth(&builder, m.providerSummary)
	builder.WriteString("Keys: up/down select, enter prints selected action\n")
	builder.WriteString("Next:\n")
	builder.WriteString("- ceo-packet --workspace ")
	builder.WriteString(workspaceArg(m.workspace))
	builder.WriteString(" --inbox\n")
	builder.WriteString("- ceo-packet --workspace ")
	builder.WriteString(workspaceArg(m.workspace))
	builder.WriteString(" --config-check --format text\n")
	return builder.String()
}

func (m tuiModel) inboxCount() int {
	count := 0
	for _, job := range m.jobs {
		if job.inboxReason != "" {
			count++
		}
	}
	return count
}

func (m tuiModel) selectedJob() tuiJob {
	if len(m.jobs) == 0 || m.selected < 0 || m.selected >= len(m.jobs) {
		return tuiJob{}
	}
	return m.jobs[m.selected]
}

func writeTUIEmpty(builder *strings.Builder, workspace string) {
	builder.WriteString("No saved jobs yet.\n")
	builder.WriteString("Setup:\n")
	builder.WriteString("- ceo-packet --quickstart ")
	builder.WriteString(workspaceArg(workspace))
	builder.WriteString("\n")
	builder.WriteString("- ceo-packet --workspace ")
	builder.WriteString(workspaceArg(workspace))
	builder.WriteString(" --config-check --format text\n")
}

func writeTUISelectedJob(builder *strings.Builder, job tuiJob) {
	builder.WriteString("Selected\n")
	builder.WriteString("Job: ")
	builder.WriteString(job.id)
	builder.WriteString("\nInbox: ")
	if job.inboxReason == "" {
		builder.WriteString("clear\n")
	} else {
		builder.WriteString(job.inboxReason)
		builder.WriteString("\n")
	}
	writeTUILine(builder, "Patch preview", job.patchPreview)
	writeTUILine(builder, "Check output", job.checkOutput)
	writeTUILine(builder, "Snapshot", job.snapshotNote)
	if job.actionCommand != "" {
		builder.WriteString("Action: ")
		builder.WriteString(job.action)
		builder.WriteString(" -> ")
		builder.WriteString(job.actionCommand)
		builder.WriteString("\n")
	}
}

func writeTUIProviderHealth(builder *strings.Builder, summary history.ProviderHealthSummary) {
	builder.WriteString(fmt.Sprintf(
		"Provider health: %d %s, %d %s, %d pass, %d fail\n",
		summary.ProviderCount,
		pluralize("provider", summary.ProviderCount),
		summary.AttemptCount,
		pluralize("attempt", summary.AttemptCount),
		summary.PassCount,
		summary.FailCount,
	))
}

func writeTUILine(builder *strings.Builder, label string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(trimText(oneLine(value), 96))
	builder.WriteString("\n")
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
	prefix := "ceo-packet --workspace " + workspaceArg(workspace)
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
