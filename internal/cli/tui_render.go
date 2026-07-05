package cli

import (
	"fmt"
	"strings"

	"ceoharness/internal/history"
)

const tuiRuleWidth = 76

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
		writeTUIEmptyConversation(&builder, m.workspace)
		writeTUIActivity(&builder, m)
		writeTUIStatus(&builder, m.providerSummary)
		writeTUIComposer(&builder, m.workspace)
		return builder.String()
	}
	writeTUIConversation(&builder, m)
	writeTUIActivity(&builder, m)
	writeTUIStatus(&builder, m.providerSummary)
	writeTUIComposer(&builder, m.workspace)
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
	writeTUISection(builder, "Cod Code")
	writeTUIBoxLine(builder, "", "Chat-first coding terminal for tasks, diffs, approvals, and proof.")
	writeTUIBoxLine(builder, "Workspace", model.workspace)
	writeTUIBoxLine(builder, "Session", fmt.Sprintf("%d %s · %d action%s waiting", len(model.jobs), pluralize("job", len(model.jobs)), model.inboxCount(), pluralSuffix(model.inboxCount())))
	writeTUISectionEnd(builder)
}

func writeTUIConversation(builder *strings.Builder, model tuiModel) {
	job := model.selectedJob()
	writeTUISection(builder, "Conversation")
	writeTUIBoxLine(builder, "You", job.task)
	writeTUIBoxLine(builder, "Cod", selectedJobSummary(job))
	if strings.TrimSpace(job.patchPreview) != "" {
		writeTUIBoxLine(builder, "Diff", job.patchPreview)
	}
	if strings.TrimSpace(job.checkOutput) != "" {
		writeTUIBoxLine(builder, "Check", job.checkOutput)
	}
	if strings.TrimSpace(job.snapshotNote) != "" {
		writeTUIBoxLine(builder, "Proof", job.snapshotNote)
	}
	if strings.TrimSpace(job.patchPreview) == "" && strings.TrimSpace(job.checkOutput) == "" && strings.TrimSpace(job.snapshotNote) == "" {
		writeTUIBoxLine(builder, "Proof", "No saved report details yet. Open the job report for the full transcript.")
	}
	writeTUIBoxLine(builder, "Action", selectedJobAction(model, job))
	writeTUIBoxLine(builder, "Rerun", tuiRetryCommand(model.workspace, job.id))
	writeTUISectionEnd(builder)
}

func selectedJobSummary(job tuiJob) string {
	state := tuiJobState(job)
	if job.inboxReason != "" {
		return state + " · waiting on you"
	}
	if strings.TrimSpace(job.verdict) != "" {
		return "verdict " + state
	}
	return "running or awaiting saved verdict"
}

func selectedJobAction(model tuiModel, job tuiJob) string {
	if job.actionCommand == "" {
		return "none"
	}
	return job.action + " · " + job.actionCommand
}

func writeTUIEmptyConversation(builder *strings.Builder, workspace string) {
	writeTUISection(builder, "Conversation")
	writeTUIBoxLine(builder, "Cod", "No chat yet. Start with a real task, then return here for the transcript, diff, and verdict.")
	writeTUIBoxLine(builder, "Run", "cod run --workspace "+workspaceArg(workspace)+" -- \"Fix one real task\"")
	writeTUIBoxLine(builder, "Setup", "cod start "+workspaceArg(workspace)+" · cod doctor --workspace "+workspaceArg(workspace)+" --format text")
	writeTUISectionEnd(builder)
}

func writeTUIActivity(builder *strings.Builder, model tuiModel) {
	writeTUISection(builder, "Activity")
	if len(model.jobs) == 0 {
		writeTUIBoxLine(builder, "Inbox", "No saved jobs yet.")
		writeTUIBoxLine(builder, "Next", "Start a task; approvals, questions, and reruns will land here.")
		writeTUISectionEnd(builder)
		return
	}
	for _, group := range model.queueGroups() {
		if len(group.items) == 0 {
			continue
		}
		writeTUIBoxLine(builder, group.tag, fmt.Sprintf("%s (%d)", group.title, len(group.items)))
		for _, item := range group.items {
			marker := " "
			if item.index == model.selected {
				marker = "›"
			}
			writeTUIBoxLine(builder, "", fmt.Sprintf("%s %-11s %-21s %s", marker, item.job.id, tuiJobState(item.job), trimText(oneLine(item.job.task), 46)))
		}
	}
	writeTUISectionEnd(builder)
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

func writeTUIStatus(builder *strings.Builder, summary history.ProviderHealthSummary) {
	writeTUISection(builder, "Status")
	if summary.ProviderCount == 0 && summary.AttemptCount == 0 {
		writeTUIBoxLine(builder, "Providers", "no evidence yet; run cod doctor or provider proof.")
		writeTUISectionEnd(builder)
		return
	}
	writeTUIBoxLine(builder, "Providers", fmt.Sprintf(
		"%d %s · %d %s · %d pass · %d fail",
		summary.ProviderCount,
		pluralize("provider", summary.ProviderCount),
		summary.AttemptCount,
		pluralize("attempt", summary.AttemptCount),
		summary.PassCount,
		summary.FailCount,
	))
	writeTUISectionEnd(builder)
}

func writeTUIComposer(builder *strings.Builder, workspace string) {
	writeTUISection(builder, "Composer")
	writeTUIBoxLine(builder, "Prompt", "cod run --workspace "+workspaceArg(workspace)+" -- \"Describe the change...\"")
	writeTUIBoxLine(builder, "Slash", "/status · /inbox · /doctor · /tools")
	writeTUIBoxLine(builder, "Context", "@path for files · !cmd for shell output · approvals appear inline above")
	writeTUIBoxLine(builder, "Keys", "j/k move · enter/a act · r rerun · q quit")
	writeTUIBoxLine(builder, "Inbox", "cod inbox --workspace "+workspaceArg(workspace))
	writeTUIBoxLine(builder, "Doctor", "cod doctor --workspace "+workspaceArg(workspace)+" --format text")
	writeTUISectionEnd(builder)
}

func writeTUISection(builder *strings.Builder, title string) {
	builder.WriteString("╭─ ")
	builder.WriteString(title)
	builder.WriteString(" ")
	builder.WriteString(strings.Repeat("─", max(1, tuiRuleWidth-len(title))))
	builder.WriteString("╮\n")
}

func writeTUISectionEnd(builder *strings.Builder) {
	builder.WriteString("╰")
	builder.WriteString(strings.Repeat("─", tuiRuleWidth+4))
	builder.WriteString("╯\n")
}

func writeTUIBoxLine(builder *strings.Builder, label string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	builder.WriteString("│ ")
	if label != "" {
		builder.WriteString(padRight(label, 10))
	}
	builder.WriteString(trimText(oneLine(value), 104))
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
