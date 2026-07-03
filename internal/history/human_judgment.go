package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const HumanJudgmentDir = "ceo-artifacts/human-judgments"

var ErrInvalidHumanJudgment = errors.New("invalid human judgment")

type HumanJudgment struct {
	JobID     string `json:"job_id"`
	CreatedAt string `json:"created_at,omitempty"`
	Verdict   string `json:"verdict"`
	Note      string `json:"note,omitempty"`
}

func (s Store) SaveHumanJudgment(ctx context.Context, judgment HumanJudgment) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	cleanID, err := cleanSnapshotJobID(judgment.JobID)
	if err != nil {
		return "", err
	}
	verdict, err := parseHumanJudgmentVerdict(judgment.Verdict)
	if err != nil {
		return "", err
	}
	saved := HumanJudgment{
		JobID:     cleanID,
		CreatedAt: strings.TrimSpace(judgment.CreatedAt),
		Verdict:   verdict,
		Note:      strings.TrimSpace(judgment.Note),
	}
	if saved.CreatedAt == "" {
		saved.CreatedAt = s.now().UTC().Format(time.RFC3339Nano)
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		return "", fmt.Errorf("encode human judgment: %w", err)
	}
	payload = append(payload, '\n')
	relativePath := humanJudgmentPath(cleanID)
	fullPath := filepath.Join(s.root, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create human judgment dir: %w", err)
	}
	if err := os.WriteFile(fullPath, payload, 0o644); err != nil {
		return "", fmt.Errorf("write human judgment: %w", err)
	}
	return relativePath, nil
}

func (s Store) ReadHumanJudgment(ctx context.Context, jobID string) (HumanJudgment, error) {
	if err := ctx.Err(); err != nil {
		return HumanJudgment{}, err
	}
	cleanID, err := cleanSnapshotJobID(jobID)
	if err != nil {
		return HumanJudgment{}, err
	}
	payload, err := os.ReadFile(filepath.Join(s.root, humanJudgmentPath(cleanID)))
	if errors.Is(err, os.ErrNotExist) {
		return HumanJudgment{}, ErrEntryNotFound
	}
	if err != nil {
		return HumanJudgment{}, fmt.Errorf("read human judgment: %w", err)
	}
	var judgment HumanJudgment
	if err := json.Unmarshal(payload, &judgment); err != nil {
		return HumanJudgment{}, fmt.Errorf("decode human judgment: %w", err)
	}
	return judgment, nil
}

func HumanJudgmentPath(jobID string) (string, error) {
	cleanID, err := cleanSnapshotJobID(jobID)
	if err != nil {
		return "", err
	}
	return humanJudgmentPath(cleanID), nil
}

func humanJudgmentPath(jobID string) string {
	return filepath.Join(HumanJudgmentDir, jobID+".json")
}

func parseHumanJudgmentVerdict(raw string) (string, error) {
	verdict := strings.ToLower(strings.TrimSpace(raw))
	switch verdict {
	case "accept", "reject":
		return verdict, nil
	default:
		return "", fmt.Errorf("verdict %q: %w", raw, ErrInvalidHumanJudgment)
	}
}
