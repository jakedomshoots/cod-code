package workspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultSearchMaxMatches = 20
	maxSearchFileBytes      = 1 << 20
)

var errSearchLimitReached = errors.New("search limit reached")

type SearchTextRequest struct {
	Query      string
	MaxMatches int
}

type SearchTextMatch struct {
	Path string
	Line int
	Text string
}

type SearchTextResult struct {
	Query   string
	Matches []SearchTextMatch
}

func (w Workspace) SearchText(ctx context.Context, req SearchTextRequest) (SearchTextResult, error) {
	if err := ctx.Err(); err != nil {
		return SearchTextResult{}, err
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return SearchTextResult{}, errors.New("search query is required")
	}
	maxMatches := req.MaxMatches
	if maxMatches < 1 {
		maxMatches = DefaultSearchMaxMatches
	}
	matches := []SearchTextMatch{}
	err := filepath.WalkDir(w.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == w.root {
			return nil
		}
		if entry.IsDir() {
			if shouldSkipSearchDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		nextMatches, err := w.searchFile(path, query, maxMatches-len(matches))
		if err != nil {
			return err
		}
		matches = append(matches, nextMatches...)
		if len(matches) >= maxMatches {
			return errSearchLimitReached
		}
		return nil
	})
	if errors.Is(err, errSearchLimitReached) {
		err = nil
	}
	if err != nil {
		return SearchTextResult{}, fmt.Errorf("search workspace: %w", err)
	}
	return SearchTextResult{Query: query, Matches: matches}, nil
}

func (w Workspace) searchFile(path string, query string, remaining int) ([]SearchTextMatch, error) {
	if remaining < 1 {
		return nil, errSearchLimitReached
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if info.Size() > maxSearchFileBytes {
		return nil, nil
	}
	rel, err := filepath.Rel(w.root, path)
	if err != nil {
		return nil, fmt.Errorf("relative path: %w", err)
	}
	cleanPath, err := cleanRelativePath(rel)
	if err != nil {
		return nil, err
	}
	if err := rejectSymlinkPath(w.root, cleanPath); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	matches := []SearchTextMatch{}
	for index, line := range strings.Split(string(content), "\n") {
		if !strings.Contains(line, query) {
			continue
		}
		matches = append(matches, SearchTextMatch{
			Path: cleanPath,
			Line: index + 1,
			Text: truncateSearchLine(line),
		})
		if len(matches) >= remaining {
			return matches, nil
		}
	}
	return matches, nil
}

func shouldSkipSearchDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", "build", "ceo-artifacts":
		return true
	default:
		return false
	}
}

func truncateSearchLine(line string) string {
	const maxLineBytes = 240
	if len(line) <= maxLineBytes {
		return line
	}
	return truncateTextBytes(line, maxLineBytes)
}
