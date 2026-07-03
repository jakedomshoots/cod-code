package cli

import (
	"fmt"
	"strings"
)

func parseProviderRecommendation(raw string) (string, error) {
	clean := strings.TrimSpace(raw)
	switch clean {
	case "avoid", "watch", "healthy", "unknown":
		return clean, nil
	default:
		return "", fmt.Errorf("--recommendation must be one of avoid, watch, healthy, unknown")
	}
}
