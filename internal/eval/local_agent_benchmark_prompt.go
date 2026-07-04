package eval

import (
	"fmt"
	"strings"
)

func appendCEORequiredCheckArgs(args []string, task Task, tail ...string) []string {
	if len(task.RequiredCommands) == 0 {
		return append(args, tail...)
	}
	checkCommand := strings.Join(task.RequiredCommands, " && ")
	args = append(args, "--check", "sh", "-c", checkCommand)
	args = append(args, "--")
	return append(args, tail...)
}

func localAgentBenchmarkPrompt(task Task) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Complete benchmark task %s: %s.\n", task.ID, task.Objective)
	fmt.Fprintf(&builder, "Edit only the required files unless an evidence artifact is required.\n")
	fmt.Fprintf(&builder, "Required changed files: %s.\n", strings.Join(task.RequiredChangedFiles, ", "))
	if len(task.RequiredChangedFiles) > 1 {
		fmt.Fprintf(&builder, "This is a multi-file source change; update every required changed file and keep the edits consistent across files.\n")
	}
	if len(task.RequiredArtifacts) > 0 {
		fmt.Fprintf(&builder, "Required evidence artifacts: %s.\n", strings.Join(task.RequiredArtifacts, ", "))
		fmt.Fprintf(&builder, "Create every required evidence artifact as a non-empty markdown file inside the workspace.\n")
		fmt.Fprintf(&builder, "Each evidence artifact must summarize the change, commands run, and verification result.\n")
	}
	fmt.Fprintf(&builder, "Required diff terms: %s.\n", strings.Join(task.RequiredDiffTerms, ", "))
	fmt.Fprintf(&builder, "Required commands to satisfy after the edit: %s.\n", strings.Join(task.RequiredCommands, "; "))
	fmt.Fprintf(&builder, "Do not inspect unrelated files or run broad test suites; run only the required commands for verification.\n")
	fmt.Fprintf(&builder, "Stop as soon as the required files, evidence artifacts, diff terms, and commands are satisfied.\n")
	fmt.Fprintf(&builder, "Keep the change minimal and do not remove the Go fixture files.")
	return builder.String()
}
