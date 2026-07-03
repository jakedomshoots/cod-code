package ceo

import "ceoharness/internal/jobpacket"

func contextBudgetForAgent(packet jobpacket.Packet, agent jobpacket.Subagent) int {
	if agent.MaxContextBytes > 0 {
		return agent.MaxContextBytes
	}
	return packet.ContextPolicy.MaxBytes
}
