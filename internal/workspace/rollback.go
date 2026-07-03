package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrUnsupportedRollback = errors.New("unsupported rollback diff")

type rollbackDiff struct {
	Path   string
	Before string
	After  string
}

func (w Workspace) RollbackReplaceText(ctx context.Context, result ReplaceTextResult) (ReplaceTextResult, error) {
	if result.Path != "" && result.Old != "" && result.New != "" {
		return w.ReplaceText(ctx, ReplaceTextRequest{
			Path: result.Path,
			Old:  result.New,
			New:  result.Old,
		})
	}
	parsed, err := parseRollbackDiff(result)
	if err != nil {
		return ReplaceTextResult{}, err
	}
	return w.ReplaceText(ctx, ReplaceTextRequest{
		Path: parsed.Path,
		Old:  parsed.After,
		New:  parsed.Before,
	})
}

func parseRollbackDiff(result ReplaceTextResult) (rollbackDiff, error) {
	lines := strings.Split(result.Diff, "\n")
	if len(lines) != 5 {
		return rollbackDiff{}, ErrUnsupportedRollback
	}
	if !strings.HasPrefix(lines[0], "--- ") || !strings.HasPrefix(lines[1], "+++ ") {
		return rollbackDiff{}, ErrUnsupportedRollback
	}
	if !strings.HasPrefix(lines[2], "-") || !strings.HasPrefix(lines[3], "+") || lines[4] != "" {
		return rollbackDiff{}, ErrUnsupportedRollback
	}
	before := strings.TrimPrefix(lines[2], "-")
	after := strings.TrimPrefix(lines[3], "+")
	path := strings.TrimPrefix(lines[0], "--- ")
	if path == "" || path != strings.TrimPrefix(lines[1], "+++ ") || path != result.Path {
		return rollbackDiff{}, fmt.Errorf("%w: path mismatch", ErrUnsupportedRollback)
	}
	if before == "" || after == "" {
		return rollbackDiff{}, ErrUnsupportedRollback
	}
	return rollbackDiff{Path: path, Before: before, After: after}, nil
}
