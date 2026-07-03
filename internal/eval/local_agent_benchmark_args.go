package eval

import (
	"fmt"
	"path/filepath"
)

func localAgentBenchmarkArgs(spec localAgentSpec, resultDir string) ([]string, error) {
	args := append([]string(nil), spec.args...)
	if spec.id != "ceo_harness" {
		return args, nil
	}
	artifactRoot, err := filepath.Abs(filepath.Join(resultDir, "runtime-artifacts"))
	if err != nil {
		return nil, fmt.Errorf("resolve CEO artifact root: %w", err)
	}
	return append([]string{"--artifact-root", artifactRoot}, args...), nil
}
