package workspace

import (
	"context"
	"fmt"
	"os"
)

const DefaultReadMaxBytes = 8192

type ReadTextRequest struct {
	Path     string
	MaxBytes int
}

type ReadTextResult struct {
	Path      string
	Content   string
	Bytes     int
	Truncated bool
}

func (w Workspace) ReadText(ctx context.Context, req ReadTextRequest) (ReadTextResult, error) {
	if err := ctx.Err(); err != nil {
		return ReadTextResult{}, err
	}
	cleanPath, err := cleanRelativePath(req.Path)
	if err != nil {
		return ReadTextResult{}, err
	}
	target, err := w.existingTarget(cleanPath)
	if err != nil {
		return ReadTextResult{}, err
	}
	content, err := os.ReadFile(target)
	if err != nil {
		return ReadTextResult{}, fmt.Errorf("read file: %w", err)
	}
	maxBytes := req.MaxBytes
	if maxBytes < 1 {
		maxBytes = DefaultReadMaxBytes
	}
	text := string(content)
	truncated := false
	if len(text) > maxBytes {
		text = truncateTextBytes(text, maxBytes)
		truncated = true
	}
	return ReadTextResult{
		Path:      cleanPath,
		Content:   text,
		Bytes:     len(text),
		Truncated: truncated,
	}, nil
}

func truncateTextBytes(text string, maxBytes int) string {
	end := 0
	for index := range text {
		if index > maxBytes {
			break
		}
		end = index
	}
	if end == 0 && len(text) > 0 {
		return ""
	}
	return text[:end]
}
