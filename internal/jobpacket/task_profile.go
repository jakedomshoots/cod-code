package jobpacket

import "strings"

func classifyTaskProfile(task string) TaskProfile {
	lowerTask := strings.ToLower(task)
	kind := classifyTaskKind(lowerTask)
	riskAreas := classifyRiskAreas(lowerTask)
	return TaskProfile{
		Kind:      kind,
		RiskLevel: classifyRiskLevel(lowerTask, kind, riskAreas),
		RiskAreas: riskAreas,
	}
}

func classifyTaskKind(task string) string {
	coding := containsAnySignal(task, []string{
		"fix", "implement", "build", "code", "refactor", "test", "bug", "api",
	})
	planning := containsAnySignal(task, []string{
		"plan", "roadmap", "strategy", "architect", "design", "scope", "spec",
	})
	research := containsAnySignal(task, []string{
		"research", "look up", "compare", "find", "latest", "docs", "source",
	})
	count := 0
	for _, matched := range []bool{coding, planning, research} {
		if matched {
			count++
		}
	}
	if count > 1 {
		return "mixed"
	}
	if research {
		return "research"
	}
	if planning {
		return "planning"
	}
	return "coding"
}

func classifyRiskLevel(task string, kind string, riskAreas []string) string {
	if len(riskAreas) > 0 {
		return "high"
	}
	if kind == "mixed" || containsAnySignal(task, []string{
		"refactor", "api", "config", "release", "production", "prod",
	}) {
		return "medium"
	}
	return "low"
}

func classifyRiskAreas(task string) []string {
	riskAreas := []string{}
	if containsAnySignal(task, []string{"payment", "billing", "invoice", "subscription"}) {
		riskAreas = append(riskAreas, "billing")
	}
	if containsAnySignal(task, []string{"database", "migration", "schema", "delete", "data loss"}) {
		riskAreas = append(riskAreas, "database")
	}
	if containsAnySignal(task, []string{"deploy", "deployment", "release", "production", "prod"}) {
		riskAreas = append(riskAreas, "release")
	}
	if containsAnySignal(task, []string{"auth", "security", "secret", "token", "permission"}) {
		riskAreas = append(riskAreas, "security")
	}
	return riskAreas
}

func containsAnySignal(text string, signals []string) bool {
	for _, signal := range signals {
		if strings.Contains(text, signal) {
			return true
		}
	}
	return false
}
