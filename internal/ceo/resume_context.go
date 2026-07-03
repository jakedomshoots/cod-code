package ceo

import (
	"fmt"
	"strings"
)

func normalizeResumeContext(input *ResumeContext) *ResumeContext {
	if input == nil {
		return nil
	}
	context := ResumeContext{
		JobID:     strings.TrimSpace(input.JobID),
		Questions: cleanResumeLines(input.Questions),
		Answers:   cleanResumeLines(input.Answers),
	}
	if context.JobID == "" && len(context.Questions) == 0 && len(context.Answers) == 0 {
		return nil
	}
	return &context
}

func taskWithResumeContext(task string, resume *ResumeContext) string {
	cleanTask := strings.TrimSpace(task)
	context := normalizeResumeContext(resume)
	if context == nil {
		return cleanTask
	}
	var builder strings.Builder
	builder.WriteString(cleanTask)
	builder.WriteString("\n\nresume_context:\n")
	if context.JobID != "" {
		builder.WriteString(fmt.Sprintf("previous_job: %s\n", context.JobID))
	}
	writeResumePairs(&builder, context.Questions, context.Answers)
	return strings.TrimSpace(builder.String())
}

func writeResumePairs(builder *strings.Builder, questions []string, answers []string) {
	pairCount := max(len(questions), len(answers))
	for index := 0; index < pairCount; index++ {
		if index < len(questions) {
			builder.WriteString(fmt.Sprintf("question: %s\n", questions[index]))
		}
		if index < len(answers) {
			builder.WriteString(fmt.Sprintf("answer: %s\n", answers[index]))
		}
	}
}

func cleanResumeLines(lines []string) []string {
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			cleaned = append(cleaned, cleanLine)
		}
	}
	return cleaned
}
