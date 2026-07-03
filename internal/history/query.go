package history

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type TimeRange struct {
	Since time.Time
	Until time.Time
}

func (s Store) ReadByVerdict(ctx context.Context, verdict string) ([]Entry, error) {
	cleanVerdict := strings.TrimSpace(verdict)
	if cleanVerdict == "" {
		return s.ReadAll(ctx)
	}
	entries, err := s.ReadAll(ctx)
	if err != nil {
		return nil, err
	}
	filtered := []Entry{}
	for _, entry := range entries {
		if entry.Verdict == cleanVerdict {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func (s Store) ReadRecent(ctx context.Context, limit int) ([]Entry, error) {
	entries, err := s.ReadAll(ctx)
	if err != nil {
		return nil, err
	}
	return limitEntries(entries, limit), nil
}

func LimitEntries(entries []Entry, limit int) []Entry {
	return limitEntries(entries, limit)
}

func FilterByCreatedAtRange(entries []Entry, bounds TimeRange) ([]Entry, error) {
	if bounds.Since.IsZero() && bounds.Until.IsZero() {
		return append([]Entry(nil), entries...), nil
	}
	if !bounds.Since.IsZero() && !bounds.Until.IsZero() && bounds.Until.Before(bounds.Since) {
		return nil, ErrInvalidTimeRange
	}
	filtered := []Entry{}
	for _, entry := range entries {
		cleanCreatedAt := strings.TrimSpace(entry.CreatedAt)
		if cleanCreatedAt == "" {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339Nano, cleanCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.ID, ErrInvalidCreatedAt)
		}
		if !bounds.Since.IsZero() && createdAt.Before(bounds.Since) {
			continue
		}
		if !bounds.Until.IsZero() && createdAt.After(bounds.Until) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered, nil
}

func FilterByTaskSubstring(entries []Entry, query string) []Entry {
	cleanQuery := strings.ToLower(strings.TrimSpace(query))
	if cleanQuery == "" {
		return append([]Entry(nil), entries...)
	}
	filtered := []Entry{}
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Task), cleanQuery) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func limitEntries(entries []Entry, limit int) []Entry {
	if limit < 1 || limit >= len(entries) {
		return append([]Entry(nil), entries...)
	}
	start := len(entries) - limit
	return append([]Entry(nil), entries[start:]...)
}

func (s Store) FindByID(ctx context.Context, id string) (Entry, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return Entry{}, ErrEntryNotFound
	}
	entries, err := s.ReadAll(ctx)
	if err != nil {
		return Entry{}, err
	}
	for _, entry := range entries {
		if entry.ID == cleanID {
			return entry, nil
		}
	}
	return Entry{}, ErrEntryNotFound
}
