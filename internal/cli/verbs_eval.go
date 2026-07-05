package cli

func gauntletEvalArgs(args []string) []string {
	normalized := []string{"--local-agent-benchmark", "--local-agent-benchmark-task", "market-parity-core"}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--agents":
			normalized = append(normalized, "--local-agents")
		case "--task", "--suite":
			normalized = append(normalized, "--local-agent-benchmark-task")
		case "--concurrency":
			normalized = append(normalized, "--local-agent-benchmark-concurrency")
		default:
			normalized = append(normalized, args[index])
		}
	}
	return normalized
}
