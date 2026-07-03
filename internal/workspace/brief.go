package workspace

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const DefaultBriefMaxFiles = 40

type BriefRequest struct {
	MaxFiles     int
	ExcludePaths []string
}

type BriefFile struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes"`
}

type Brief struct {
	FileCount         int         `json:"file_count"`
	Files             []BriefFile `json:"files"`
	Truncated         bool        `json:"truncated"`
	ExcludePaths      []string    `json:"exclude_paths,omitempty"`
	ExcludedPathCount int         `json:"excluded_path_count,omitempty"`
}

func (w Workspace) Brief(ctx context.Context, req BriefRequest) (Brief, error) {
	if err := ctx.Err(); err != nil {
		return Brief{}, err
	}
	maxFiles := req.MaxFiles
	if maxFiles < 1 {
		maxFiles = DefaultBriefMaxFiles
	}
	excludes, err := cleanBriefExcludes(req.ExcludePaths)
	if err != nil {
		return Brief{}, err
	}
	brief := Brief{
		Files:        make([]BriefFile, 0, maxFiles),
		ExcludePaths: append([]string(nil), excludes...),
	}
	err = filepath.WalkDir(w.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == w.root {
			return nil
		}
		rel, err := filepath.Rel(w.root, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		cleanPath, err := cleanRelativePath(rel)
		if err != nil {
			return err
		}
		if briefPathExcluded(cleanPath, excludes) {
			brief.ExcludedPathCount++
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if shouldSkipSearchDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipBriefFile(entry.Name()) {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return addBriefFile(w.root, path, maxFiles, &brief)
	})
	if err != nil {
		return Brief{}, fmt.Errorf("build workspace brief: %w", err)
	}
	return brief, nil
}

func cleanBriefExcludes(patterns []string) ([]string, error) {
	excludes := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		cleanPattern, err := cleanRelativePath(pattern)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, filepath.ToSlash(cleanPattern))
	}
	return excludes, nil
}

func briefPathExcluded(path string, patterns []string) bool {
	cleanPath := filepath.ToSlash(path)
	for _, pattern := range patterns {
		if cleanPath == pattern || strings.HasPrefix(cleanPath, pattern+"/") {
			return true
		}
		if strings.HasSuffix(pattern, "/**") && strings.HasPrefix(cleanPath, strings.TrimSuffix(pattern, "/**")+"/") {
			return true
		}
		if matched, err := filepath.Match(pattern, cleanPath); err == nil && matched {
			return true
		}
	}
	return false
}

func shouldSkipBriefFile(name string) bool {
	return name == ".ceo-harness.json"
}

func addBriefFile(root string, path string, maxFiles int, brief *Brief) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("relative path: %w", err)
	}
	cleanPath, err := cleanRelativePath(rel)
	if err != nil {
		return err
	}
	if err := rejectSymlinkPath(root, cleanPath); err != nil {
		return err
	}
	brief.FileCount++
	if len(brief.Files) >= maxFiles {
		brief.Truncated = true
		return nil
	}
	brief.Files = append(brief.Files, BriefFile{
		Path:  cleanPath,
		Bytes: info.Size(),
	})
	return nil
}
