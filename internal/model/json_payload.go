package model

import "strings"

func JSONPayload(text string) (string, bool) {
	clean := strings.TrimSpace(text)
	if strings.HasPrefix(clean, "{") {
		if payload, ok := balancedJSONPayload(clean, 0); ok {
			return payload, true
		}
		return clean, true
	}
	if payload, ok := fencedJSONPayload(clean); ok {
		return payload, true
	}
	return embeddedJSONPayload(clean)
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

func embeddedJSONPayload(text string) (string, bool) {
	for start := strings.IndexByte(text, '{'); start >= 0; {
		if payload, ok := balancedJSONPayload(text, start); ok {
			return strings.TrimSpace(payload), true
		}
		next := strings.IndexByte(text[start+1:], '{')
		if next < 0 {
			break
		}
		start += next + 1
	}
	return "", false
}

func balancedJSONPayload(text string, start int) (string, bool) {
	if start < 0 || start >= len(text) || text[start] != '{' {
		return "", false
	}
	depth := 0
	inString := false
	escaped := false
	for index := start; index < len(text); index++ {
		character := text[index]
		if inString {
			switch {
			case escaped:
				escaped = false
			case character == '\\':
				escaped = true
			case character == '"':
				inString = false
			}
			continue
		}
		switch character {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : index+1], true
			}
		}
	}
	return "", false
}
