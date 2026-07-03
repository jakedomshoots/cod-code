package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func modelCommandFromEnv() ([]string, error) {
	return commandFromJSONEnv(modelCommandEnv)
}

func ceoModelCommandFromEnv() ([]string, error) {
	return commandFromJSONEnv(ceoModelCommandEnv)
}

func researchCommandFromEnv() ([]string, error) {
	return commandFromJSONEnv(researchCommandEnv)
}

func commandFromJSONEnv(name string) ([]string, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return nil, nil
	}
	var argv []string
	if err := json.Unmarshal([]byte(raw), &argv); err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	if len(argv) == 0 {
		return nil, fmt.Errorf("%s must contain at least one argv item", name)
	}
	return argv, nil
}
