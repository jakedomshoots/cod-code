package jobpacket

func riskSpecialists(riskAreas []string) []Subagent {
	if len(riskAreas) == 0 {
		return nil
	}
	specialists := make([]Subagent, 0, len(riskAreas))
	for _, riskArea := range riskAreas {
		switch riskArea {
		case "billing":
			specialists = append(specialists, newSubagent("billing", "review payment risk"))
		case "database":
			specialists = append(specialists, newSubagent("database", "review data migration risk"))
		case "release":
			specialists = append(specialists, newSubagent("release", "review deployment risk"))
		case "security":
			specialists = append(specialists, newSubagent("security", "review auth and secret risk"))
		}
	}
	return specialists
}
