package jobpacket

import "strings"

type Action string

const (
	ActionReadWorkspace    Action = "read_workspace"
	ActionSearchWorkspace  Action = "search_workspace"
	ActionNetworkResearch  Action = "network_research"
	ActionBrowserRead      Action = "browser_read"
	ActionComputerSnapshot Action = "computer_snapshot"
	ActionToolManifest     Action = "tool_manifest"
	ActionProposePatch     Action = "propose_patch"
	ActionRunChecks        Action = "run_checks"
	ActionVerifyEvidence   Action = "verify_evidence"
)

func DefaultActionsForAgent(name string) []Action {
	switch name {
	case "researcher":
		return []Action{ActionReadWorkspace, ActionSearchWorkspace, ActionNetworkResearch}
	case "coder":
		return []Action{ActionReadWorkspace, ActionSearchWorkspace, ActionProposePatch}
	case "billing", "database", "release", "security":
		return []Action{ActionReadWorkspace, ActionSearchWorkspace, ActionRunChecks}
	case "reviewer":
		return []Action{ActionReadWorkspace, ActionRunChecks, ActionVerifyEvidence}
	default:
		return []Action{ActionReadWorkspace, ActionSearchWorkspace}
	}
}

func NormalizeActions(actions []Action) ([]Action, bool) {
	if len(actions) == 0 {
		return nil, true
	}
	normalized := make([]Action, 0, len(actions))
	seen := map[Action]struct{}{}
	for _, action := range actions {
		clean := Action(strings.TrimSpace(string(action)))
		if !IsKnownAction(clean) {
			return nil, false
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	return normalized, true
}

func IsKnownAction(action Action) bool {
	switch action {
	case ActionReadWorkspace,
		ActionSearchWorkspace,
		ActionNetworkResearch,
		ActionBrowserRead,
		ActionComputerSnapshot,
		ActionToolManifest,
		ActionProposePatch,
		ActionRunChecks,
		ActionVerifyEvidence:
		return true
	default:
		return false
	}
}

func ActionStrings(actions []Action) []string {
	out := make([]string, 0, len(actions))
	for _, action := range actions {
		out = append(out, string(action))
	}
	return out
}
