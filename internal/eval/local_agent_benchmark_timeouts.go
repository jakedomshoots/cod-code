package eval

import (
	"fmt"
	"strconv"
	"strings"
)

func parseLocalAgentBenchmarkAgentTimeouts(raw string) (map[string]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	timeouts := make(map[string]int)
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: invalid agent timeout %q; want agent=seconds", ErrInvalidTask, entry)
		}
		agent := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if agent == "" {
			return nil, fmt.Errorf("%w: invalid agent timeout %q; missing agent", ErrInvalidTask, entry)
		}
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds <= 0 {
			return nil, fmt.Errorf("%w: invalid timeout for %s; want positive seconds", ErrInvalidTask, agent)
		}
		timeouts[agent] = seconds
	}
	if len(timeouts) == 0 {
		return nil, nil
	}
	return timeouts, nil
}

func parseLocalAgentBenchmarkAgentModels(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	models := make(map[string]string)
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: invalid agent model %q; want agent=provider/model", ErrInvalidTask, entry)
		}
		agent := strings.TrimSpace(parts[0])
		model := strings.TrimSpace(parts[1])
		if agent == "" || model == "" {
			return nil, fmt.Errorf("%w: invalid agent model %q; missing agent or model", ErrInvalidTask, entry)
		}
		models[agent] = model
	}
	if len(models) == 0 {
		return nil, nil
	}
	return models, nil
}

func normalizeLocalAgentBenchmarkAgentModels(raw map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	models := make(map[string]string)
	for agent, model := range raw {
		agent = strings.TrimSpace(agent)
		model = strings.TrimSpace(model)
		if agent != "" && model != "" {
			models[agent] = model
		}
	}
	if len(models) == 0 {
		return nil
	}
	return models
}
