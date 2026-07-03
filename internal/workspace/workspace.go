package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathEscapesWorkspace = errors.New("path escapes workspace")
	ErrTextNotFound         = errors.New("text to replace was not found")
	ErrFileAlreadyExists    = errors.New("file already exists")
)

type Workspace struct {
	root string
}

type WriteTextRequest struct {
	Path    string
	Content string
}

type WriteTextResult struct {
	Path string
}

type CreateTextRequest struct {
	Path    string
	Content string
}

type ReplaceTextRequest struct {
	Path string
	Old  string
	New  string
}

type ReplaceTextResult struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
	Old  string `json:"old,omitempty"`
	New  string `json:"new,omitempty"`
}

func New(root string) (Workspace, error) {
	cleanRoot := strings.TrimSpace(root)
	if cleanRoot == "" {
		return Workspace{}, errors.New("workspace root is required")
	}
	return Workspace{root: cleanRoot}, nil
}

func (w Workspace) WriteText(ctx context.Context, req WriteTextRequest) (WriteTextResult, error) {
	if err := ctx.Err(); err != nil {
		return WriteTextResult{}, err
	}

	cleanPath, err := cleanRelativePath(req.Path)
	if err != nil {
		return WriteTextResult{}, err
	}
	target, err := w.writeTarget(cleanPath)
	if err != nil {
		return WriteTextResult{}, err
	}
	if err := os.WriteFile(target, []byte(req.Content), 0o644); err != nil {
		return WriteTextResult{}, fmt.Errorf("write file: %w", err)
	}

	return WriteTextResult{Path: cleanPath}, nil
}

func (w Workspace) ReplaceText(ctx context.Context, req ReplaceTextRequest) (ReplaceTextResult, error) {
	return w.replaceText(ctx, req, true)
}

func (w Workspace) PreviewReplaceText(ctx context.Context, req ReplaceTextRequest) (ReplaceTextResult, error) {
	return w.replaceText(ctx, req, false)
}

func (w Workspace) CreateText(ctx context.Context, req CreateTextRequest) (ReplaceTextResult, error) {
	return w.createText(ctx, req, true)
}

func (w Workspace) PreviewCreateText(ctx context.Context, req CreateTextRequest) (ReplaceTextResult, error) {
	return w.createText(ctx, req, false)
}

func (w Workspace) createText(ctx context.Context, req CreateTextRequest, write bool) (result ReplaceTextResult, err error) {
	if err := ctx.Err(); err != nil {
		return ReplaceTextResult{}, err
	}
	cleanPath, err := cleanRelativePath(req.Path)
	if err != nil {
		return ReplaceTextResult{}, err
	}
	target, err := w.createTarget(cleanPath, write)
	if err != nil {
		return ReplaceTextResult{}, err
	}
	if _, err := os.Lstat(target); err == nil {
		return ReplaceTextResult{}, ErrFileAlreadyExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return ReplaceTextResult{}, fmt.Errorf("stat file: %w", err)
	}
	if write {
		var file *os.File
		file, err = os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, os.ErrExist) {
			return ReplaceTextResult{}, ErrFileAlreadyExists
		}
		if err != nil {
			return ReplaceTextResult{}, fmt.Errorf("create file: %w", err)
		}
		defer func() {
			if closeErr := file.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("close file: %w", closeErr)
			}
		}()
		if _, err = file.WriteString(req.Content); err != nil {
			return ReplaceTextResult{}, fmt.Errorf("write file: %w", err)
		}
	}
	return ReplaceTextResult{
		Path: cleanPath,
		Diff: renderDiff(cleanPath, "", req.Content),
		New:  req.Content,
	}, nil
}

func (w Workspace) replaceText(ctx context.Context, req ReplaceTextRequest, write bool) (ReplaceTextResult, error) {
	if err := ctx.Err(); err != nil {
		return ReplaceTextResult{}, err
	}

	cleanPath, err := cleanRelativePath(req.Path)
	if err != nil {
		return ReplaceTextResult{}, err
	}
	target, err := w.existingTarget(cleanPath)
	if err != nil {
		return ReplaceTextResult{}, err
	}
	content, err := os.ReadFile(target)
	if err != nil {
		return ReplaceTextResult{}, fmt.Errorf("read file: %w", err)
	}
	current := string(content)
	if !strings.Contains(current, req.Old) {
		return ReplaceTextResult{}, ErrTextNotFound
	}
	next := strings.Replace(current, req.Old, req.New, 1)
	if write {
		if err := os.WriteFile(target, []byte(next), 0o644); err != nil {
			return ReplaceTextResult{}, fmt.Errorf("write file: %w", err)
		}
	}

	return ReplaceTextResult{
		Path: cleanPath,
		Diff: renderDiff(cleanPath, current, next),
		Old:  req.Old,
		New:  req.New,
	}, nil
}

func renderDiff(path string, before string, after string) string {
	return fmt.Sprintf("--- %s\n+++ %s\n-%s\n+%s\n", path, path, before, after)
}

func cleanRelativePath(path string) (string, error) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	if cleanPath == "." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return "", ErrPathEscapesWorkspace
	}
	return cleanPath, nil
}

func (w Workspace) writeTarget(cleanPath string) (string, error) {
	if err := w.ensureSafeParent(cleanPath); err != nil {
		return "", err
	}
	target := filepath.Join(w.root, cleanPath)
	if err := rejectSymlink(target); err != nil {
		return "", err
	}
	return target, nil
}

func (w Workspace) createTarget(cleanPath string, write bool) (string, error) {
	if write {
		return w.writeTarget(cleanPath)
	}
	target := filepath.Join(w.root, cleanPath)
	if err := rejectSymlinkPath(w.root, cleanPath); err != nil {
		return "", err
	}
	return target, nil
}

func (w Workspace) existingTarget(cleanPath string) (string, error) {
	target := filepath.Join(w.root, cleanPath)
	if err := rejectSymlinkPath(w.root, cleanPath); err != nil {
		return "", err
	}
	return target, nil
}

func (w Workspace) ensureSafeParent(cleanPath string) error {
	current := w.root
	parent := filepath.Dir(cleanPath)
	if parent == "." {
		return nil
	}
	for _, segment := range strings.Split(parent, string(filepath.Separator)) {
		if segment == "" || segment == "." {
			continue
		}
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o755); err != nil {
				return fmt.Errorf("create parent dir: %w", err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("stat parent dir: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return ErrPathEscapesWorkspace
		}
		if !info.IsDir() {
			return fmt.Errorf("parent is not a directory: %s", segment)
		}
	}
	return nil
}

func rejectSymlinkPath(root string, cleanPath string) error {
	current := root
	for _, segment := range strings.Split(cleanPath, string(filepath.Separator)) {
		if segment == "" || segment == "." {
			continue
		}
		current = filepath.Join(current, segment)
		if err := rejectSymlink(current); err != nil {
			return err
		}
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrPathEscapesWorkspace
	}
	return nil
}
