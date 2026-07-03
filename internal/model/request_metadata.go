package model

import "strings"

func metadataEnvPairs(metadata RequestMetadata) []string {
	pairs := []string{}
	addMetadataEnvPair(&pairs, "CEO_MODEL_REQUEST_KIND", metadata.Kind)
	addMetadataEnvPair(&pairs, "CEO_AGENT_NAME", metadata.AgentName)
	addMetadataEnvPair(&pairs, "CEO_AGENT_ROLE", metadata.AgentRole)
	addMetadataEnvPair(&pairs, "CEO_CONTEXT_MODE", metadata.ContextMode)
	return pairs
}

func addMetadataEnvPair(pairs *[]string, name string, value string) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return
	}
	*pairs = append(*pairs, name+"="+clean)
}
