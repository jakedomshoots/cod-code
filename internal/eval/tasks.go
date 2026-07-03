package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var allowedCategories = map[string]struct{}{
	"bug_fix":         {},
	"cross_language":  {},
	"refactor":        {},
	"test_repair":     {},
	"docs":            {},
	"provider_config": {},
	"recovery":        {},
	"report_quality":  {},
	"rollback":        {},
	"safety_policy":   {},
}

func LoadTasks(ctx context.Context, dir string) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tasks := []Task{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		loaded, err := loadTaskFile(path)
		if err != nil {
			return err
		}
		tasks = append(tasks, loaded...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load eval tasks: %w", err)
	}
	if err := validateTasks(tasks); err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(left, right int) bool {
		return tasks[left].ID < tasks[right].ID
	})
	return tasks, nil
}

func loadTaskFile(path string) ([]Task, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task %s: %w", path, err)
	}
	var tasks []Task
	if err := json.Unmarshal(content, &tasks); err == nil {
		return tasks, nil
	}
	var task Task
	if err := json.Unmarshal(content, &task); err != nil {
		return nil, fmt.Errorf("decode task %s: %w", path, err)
	}
	return []Task{task}, nil
}

func validateTasks(tasks []Task) error {
	if len(tasks) == 0 {
		return fmt.Errorf("%w: no tasks found", ErrInvalidTask)
	}
	seen := map[string]struct{}{}
	for _, task := range tasks {
		if err := validateTask(task); err != nil {
			return err
		}
		if _, ok := seen[task.ID]; ok {
			return fmt.Errorf("%w: duplicate task id %q", ErrInvalidTask, task.ID)
		}
		seen[task.ID] = struct{}{}
	}
	return nil
}

func validateTask(task Task) error {
	if strings.TrimSpace(task.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidTask)
	}
	if _, ok := allowedCategories[task.Category]; !ok {
		return fmt.Errorf("%w: task %s has unknown category %q", ErrInvalidTask, task.ID, task.Category)
	}
	if strings.TrimSpace(task.Title) == "" || strings.TrimSpace(task.Objective) == "" {
		return fmt.Errorf("%w: task %s needs title and objective", ErrInvalidTask, task.ID)
	}
	if len(task.RequiredChangedFiles)+len(task.ForbiddenChangedFiles)+len(task.RequiredCommands)+len(task.RequiredArtifacts)+len(task.RequiredDiffTerms)+len(task.RequiredReportFields) == 0 {
		return fmt.Errorf("%w: task %s needs at least one scoring rule", ErrInvalidTask, task.ID)
	}
	return nil
}

func FindTask(tasks []Task, id string) (Task, error) {
	cleanID := strings.TrimSpace(id)
	for _, task := range tasks {
		if task.ID == cleanID {
			return task, nil
		}
	}
	return Task{}, fmt.Errorf("%w: task %q not found", ErrInvalidTask, cleanID)
}
