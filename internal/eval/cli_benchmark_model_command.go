package eval

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseCEOBenchmarkModelCommand(raw string) ([]string, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return nil, nil
	}
	var command []string
	if err := json.Unmarshal([]byte(clean), &command); err != nil {
		return nil, fmt.Errorf("parse --ceo-benchmark-model-command-json: %w", err)
	}
	for _, part := range command {
		if strings.TrimSpace(part) == "" {
			return nil, fmt.Errorf("parse --ceo-benchmark-model-command-json: command entries must be non-empty")
		}
	}
	return command, nil
}
