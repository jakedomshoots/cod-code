package cli

import (
	"fmt"
	"strings"

	"ceoharness/internal/history"
)

type tuiQueueGroup struct {
	title string
	tag   string
	items []tuiQueueItem
}

type tuiQueueItem struct {
	index int
	job   tuiJob
}

func (m tuiModel) render() string {
	var builder strings.Builder
	writeTUIHeader(&builder, m)
	if len(m.jobs) == 0 {
		writeTUIEmpty(&builder, m.workspace)
		writeTUIProviderHealth(&builder, m.providerSummary)
		writeTUICommandHints(&builder, m.workspace)
		return builder.String()
	}
	writeTUIQueue(&builder, m)
	writeTUISelectedJob(&builder, m)
	writeTUIProviderHealth(&builder, m.providerSummary)
	writeTUICommandHints(&builder, m.workspace)
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

func writeTUIHeader(builder *strings.Builder, model tuiModel) {
	builder.WriteString("Cod Code Chat\n")
	writeTUILine(builder, "Workspace", model.workspace)
	writeTUILine(builder, "Queue", fmt.Sprintf("%d %s", len(model.jobs), pluralize("job", len(model.jobs))))
	writeTUILine(builder, "Needs", fmt.Sprintf("%d action%s", model.inboxCount(), pluralSuffix(model.inboxCount())))
	builder.WriteString("------------------------------------------------------------\n")
}

func writeTUIQueue(builder *strings.Builder, model tuiModel) {
	builder.WriteString("Queue\n")
	for _, group := range model.queueGroups() {
		if len(group.items) == 0 {
			continue
		}
		builder.WriteString("[")
		builder.WriteString(group.tag)
		builder.WriteString("] ")
		builder.WriteString(group.title)
		builder.WriteString(fmt.Sprintf(" (%d)\n", len(group.items)))
		for _, item := range group.items {
			prefix := "  "
			if item.index == model.selected {
				prefix = "> "
			}
			builder.WriteString(prefix)
			builder.WriteString(padRight(item.job.id, 12))
			builder.WriteString(padRight(tuiJobState(item.job), 24))
			builder.WriteString(trimText(oneLine(item.job.task), 64))
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n")
}

func (m tuiModel) queueGroups() []tuiQueueGroup {
	groups := []tuiQueueGroup{
		{title: "Needs input", tag: "INPUT"},
		{title: "Needs decision", tag: "REVIEW"},
		{title: "Failed", tag: "FAIL"},
		{title: "Passed", tag: "PASS"},
		{title: "Other", tag: "OTHER"},
	}
	indexByLane := map[string]int{"needs_input": 0, "needs_decision": 1, "failed": 2, "passed": 3, "other": 4}
	for index, job := range m.jobs {
		lane := tuiJobLane(job)
		groupIndex, ok := indexByLane[lane]
		if !ok {
			groupIndex = indexByLane["other"]
		}
		groups[groupIndex].items = append(groups[groupIndex].items, tuiQueueItem{index: index, job: job})
	}
	return groups
}

func tuiJobLane(job tuiJob) string {
	if job.inboxReason == "needs_input" || job.verdict == "needs_input" {
		return "needs_input"
	}
	if job.inboxReason != "" {
		return "needs_decision"
	}
	switch strings.ToLower(strings.TrimSpace(job.verdict)) {
	case "fail", "failed", "failed_or_unresolved":
		return "failed"
	case "pass":
		return "passed"
	default:
		return "other"
	}
}

func tuiJobState(job tuiJob) string {
	if job.inboxReason != "" {
		return job.inboxReason
	}
	if strings.TrimSpace(job.verdict) != "" {
		return job.verdict
	}
	return "unknown"
}

func writeTUISelectedJob(builder *strings.Builder, model tuiModel) {
	job := model.selectedJob()
	builder.WriteString("Selected\n")
	writeTUILine(builder, "Job", job.id)
	writeTUILine(builder, "Task", job.task)
	writeTUILine(builder, "Verdict", job.verdict)
	if job.inboxReason == "" {
		writeTUILine(builder, "Inbox", "clear")
	} else {
		writeTUILine(builder, "Inbox", job.inboxReason)
	}
	builder.WriteString("\n")

	builder.WriteString("Evidence\n")
	writeTUILine(builder, "Patch", job.patchPreview)
	writeTUILine(builder, "Check", job.checkOutput)
	writeTUILine(builder, "Snapshot", job.snapshotNote)
	if strings.TrimSpace(job.patchPreview) == "" && strings.TrimSpace(job.checkOutput) == "" && strings.TrimSpace(job.snapshotNote) == "" {
		writeTUILine(builder, "Status", "no saved report details yet")
	}
	builder.WriteString("\n")

	builder.WriteString("Actions\n")
	if job.actionCommand != "" {
		writeTUILine(builder, "Primary", job.action)
		writeTUILine(builder, "Command", job.actionCommand)
	} else {
		writeTUILine(builder, "Primary", "none")
	}
	writeTUILine(builder, "Rerun", tuiRetryCommand(model.workspace, job.id))
	builder.WriteString("\n")
}

func writeTUIEmpty(builder *strings.Builder, workspace string) {
	builder.WriteString("Queue\n")
	builder.WriteString("No saved jobs yet.\n")
	builder.WriteString("Start a real task, then come back here to review chat turns, verdicts, and proof.\n\n")
	builder.WriteString("Start\n")
	writeTUICommandAction(builder, "Open chat", "cod")
	writeTUICommandAction(builder, "Quickstart", "cod start "+workspaceArg(workspace))
	writeTUICommandAction(builder, "Check setup", "cod doctor --workspace "+workspaceArg(workspace)+" --format text")
	builder.WriteString("\n")
}

func writeTUIProviderHealth(builder *strings.Builder, summary history.ProviderHealthSummary) {
	builder.WriteString("Systems\n")
	if summary.ProviderCount == 0 && summary.AttemptCount == 0 {
		writeTUILine(builder, "Providers", "no evidence yet; run provider proof or config-check.")
		builder.WriteString("\n")
		return
	}
	writeTUILine(builder, "Providers", fmt.Sprintf(
		"%d %s | %d %s | %d pass | %d fail",
		summary.ProviderCount,
		pluralize("provider", summary.ProviderCount),
		summary.AttemptCount,
		pluralize("attempt", summary.AttemptCount),
		summary.PassCount,
		summary.FailCount,
	))
	builder.WriteString("\n")
}

func writeTUICommandHints(builder *strings.Builder, workspace string) {
	builder.WriteString("Shortcuts\n")
	writeTUILine(builder, "Navigate", "j/down next | k/up previous")
	writeTUILine(builder, "Primary", "enter/a dispatch selected action")
	writeTUILine(builder, "Rerun", "r print rerun command")
	writeTUILine(builder, "Quit", "q")
	builder.WriteString("Next\n")
	writeTUICommandAction(builder, "Review inbox", "cod inbox --workspace "+workspaceArg(workspace))
	writeTUICommandAction(builder, "Check setup", "cod doctor --workspace "+workspaceArg(workspace)+" --format text")
	writeTUICommandAction(builder, "Tool map", "cod tools manifest --format json")
}

func writeTUILine(builder *strings.Builder, label string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	builder.WriteString(padRight(label, 12))
	builder.WriteString(trimText(oneLine(value), 96))
	builder.WriteString("\n")
}

func writeTUICommandAction(builder *strings.Builder, label string, command string) {
	builder.WriteString(padRight(label, 14))
	builder.WriteString(command)
	builder.WriteString("\n")
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func padRight(text string, width int) string {
	if len(text) >= width {
		return text + " "
	}
	return text + strings.Repeat(" ", width-len(text))
}
