package eval

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	localAgentSuiteMode      = "local_agent_comparison_suite"
	localAgentMarker         = "CEO_HARNESS_EVAL_OK"
	localAgentStatusPass     = "pass"
	localAgentStatusFail     = "fail"
	localAgentStatusTimeout  = "timeout"
	localAgentStatusSkipped  = "skipped_missing_binary"
	defaultLocalAgentTimeout = 45 * time.Second
	localAgentSchemaVersion  = 1
	localAgentTaskReadiness  = "readiness"
	localAgentTaskEditFile   = "edit-file"
)

var defaultLocalAgentIDs = []string{"ceo_harness", "codex_cli", "claude_code", "aider", "opencode", "goose", "pi", "oh_my_pi"}

type localAgentSpec struct {
	id                       string
	name                     string
	binary                   string
	args                     []string
	env                      []string
	workspaceConfig          []byte
	benchmarkWritesArtifacts bool
	expectedOutput           string
	expectedFile             string
	setupHint                string
}

type localAgentTaskSpec struct {
	name         string
	prompt       string
	expectedFile string
}

func localAgentTask(raw string) (localAgentTaskSpec, error) {
	switch strings.TrimSpace(raw) {
	case "", localAgentTaskReadiness:
		return localAgentTaskSpec{
			name:   localAgentTaskReadiness,
			prompt: "Reply exactly " + localAgentMarker + ". Do not edit files.",
		}, nil
	case localAgentTaskEditFile:
		return localAgentTaskSpec{
			name:         localAgentTaskEditFile,
			prompt:       "In this directory, edit app.txt so it contains exactly `hello new` followed by a newline. Do not modify any other file.",
			expectedFile: "hello new\n",
		}, nil
	default:
		return localAgentTaskSpec{}, fmt.Errorf("%w: unknown local agent task %q", ErrInvalidTask, raw)
	}
}

func normalizeLocalAgentIDs(agents []string) []string {
	if len(agents) == 0 {
		return append([]string(nil), defaultLocalAgentIDs...)
	}
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		id := strings.TrimSpace(agent)
		if id != "" && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return append([]string(nil), defaultLocalAgentIDs...)
	}
	return ids
}
