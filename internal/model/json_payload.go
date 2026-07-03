package model

import "strings"

func JSONPayload(text string) (string, bool) {
	clean := strings.TrimSpace(text)
	if strings.HasPrefix(clean, "{") {
		return clean, true
	}
	return fencedJSONPayload(clean)
}

func fencedJSONPayload(text string) (string, bool) {
	start := strings.Index(text, "```")
	if start < 0 {
		return "", false
	}
	afterFence := text[start+3:]
	lineEnd := strings.IndexByte(afterFence, '\n')
	if lineEnd < 0 {
		return "", false
	}
	language := strings.ToLower(strings.TrimSpace(afterFence[:lineEnd]))
	if language != "" && language != "json" {
		return "", false
	}
	bodyStart := start + 3 + lineEnd + 1
	end := strings.Index(text[bodyStart:], "```")
	if end < 0 {
		return "", false
	}
	payload := strings.TrimSpace(text[bodyStart : bodyStart+end])
	return payload, strings.HasPrefix(payload, "{")
}
