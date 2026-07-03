package jobpacket

func defaultSubagents(profile TaskProfile, limit int) []Subagent {
	specialists := riskSpecialists(profile.RiskAreas)
	var subagents []Subagent
	switch profile.Kind {
	case "planning":
		subagents = appendSubagents([]Subagent{
			newSubagent("planner", "break down work"),
		}, specialists, []Subagent{
			newSubagent("reviewer", "verify evidence"),
		})
	case "research":
		subagents = appendSubagents([]Subagent{
			newSubagent("researcher", "gather source-backed context"),
		}, specialists, []Subagent{
			newSubagent("reviewer", "verify evidence"),
		})
	case "mixed":
		subagents = appendSubagents([]Subagent{
			newSubagent("planner", "break down work"),
			newSubagent("researcher", "gather source-backed context"),
			newSubagent("coder", "apply bounded changes"),
		}, specialists, []Subagent{
			newSubagent("reviewer", "verify evidence"),
		})
	default:
		subagents = appendSubagents([]Subagent{
			newSubagent("scanner", "inspect scope"),
			newSubagent("coder", "apply bounded changes"),
		}, specialists, []Subagent{
			newSubagent("reviewer", "verify evidence"),
		})
	}
	return limitSubagents(profile, subagents, limit)
}

func newSubagent(name string, role string) Subagent {
	return Subagent{
		Name:           name,
		Role:           role,
		AllowedActions: DefaultActionsForAgent(name),
	}
}

func appendSubagents(first []Subagent, middle []Subagent, last []Subagent) []Subagent {
	out := make([]Subagent, 0, len(first)+len(middle)+len(last))
	out = append(out, first...)
	out = append(out, middle...)
	out = append(out, last...)
	return out
}

func limitSubagents(profile TaskProfile, subagents []Subagent, limit int) []Subagent {
	if limit <= 0 || len(subagents) <= limit {
		return subagents
	}
	selected := map[string]struct{}{}
	add := func(name string) {
		if len(selected) >= limit || !hasSubagentNamed(subagents, name) {
			return
		}
		selected[name] = struct{}{}
	}
	for _, name := range ownerPreference(profile.Kind) {
		add(name)
		if len(selected) > 0 {
			break
		}
	}
	add("reviewer")
	for _, specialist := range riskSpecialists(profile.RiskAreas) {
		add(specialist.Name)
		if len(selected) >= limit {
			break
		}
	}
	for _, subagent := range subagents {
		add(subagent.Name)
	}
	limited := make([]Subagent, 0, limit)
	for _, subagent := range subagents {
		if _, ok := selected[subagent.Name]; ok {
			limited = append(limited, subagent)
		}
	}
	return limited
}

func hasSubagentNamed(subagents []Subagent, name string) bool {
	for _, subagent := range subagents {
		if subagent.Name == name {
			return true
		}
	}
	return false
}
