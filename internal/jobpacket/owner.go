package jobpacket

func OwnerForPacket(packet Packet) string {
	for _, name := range ownerPreference(packet.TaskProfile.Kind) {
		if packetHasSubagent(packet, name) {
			return name
		}
	}
	for _, agent := range packet.Subagents {
		if agent.Name != "reviewer" && agent.Name != "security" {
			return agent.Name
		}
	}
	if len(packet.Subagents) > 0 {
		return packet.Subagents[0].Name
	}
	return "ceo"
}

func ownerPreference(kind string) []string {
	switch kind {
	case "planning":
		return []string{"planner"}
	case "research":
		return []string{"researcher"}
	case "mixed":
		return []string{"coder", "planner", "researcher"}
	default:
		return []string{"coder", "scanner"}
	}
}

func packetHasSubagent(packet Packet, name string) bool {
	for _, agent := range packet.Subagents {
		if agent.Name == name {
			return true
		}
	}
	return false
}
