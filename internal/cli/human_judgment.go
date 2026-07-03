package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/history"
)

type humanJudgmentReport struct {
	HistoryPath  string                `json:"history_path"`
	JudgmentPath string                `json:"judgment_path"`
	Source       string                `json:"source"`
	Judgment     history.HumanJudgment `json:"judgment"`
}

func runHumanJudgment(ctx context.Context, out io.Writer, opts options) error {
	store, err := history.New(opts.workspaceDir)
	if err != nil {
		return err
	}
	jobID, err := resolveSavedJobID(ctx, opts.workspaceDir, opts.judgeJobID)
	if err != nil {
		return err
	}
	if _, err := store.FindByID(ctx, jobID); err != nil {
		return fmt.Errorf("find judgment job: %w", err)
	}
	path, err := history.HumanJudgmentPath(jobID)
	if err != nil {
		return err
	}
	var judgment history.HumanJudgment
	if strings.TrimSpace(opts.humanVerdict) == "" {
		judgment, err = store.ReadHumanJudgment(ctx, jobID)
		if err != nil {
			return fmt.Errorf("find human judgment: %w", err)
		}
	} else {
		judgment = history.HumanJudgment{
			JobID:   jobID,
			Verdict: opts.humanVerdict,
			Note:    opts.judgmentNote,
		}
		path, err = store.SaveHumanJudgment(ctx, judgment)
		if err != nil {
			return err
		}
		judgment, err = store.ReadHumanJudgment(ctx, jobID)
		if err != nil {
			return fmt.Errorf("read saved human judgment: %w", err)
		}
	}
	return writeHumanJudgmentReport(out, humanJudgmentReport{
		HistoryPath:  store.Path(),
		JudgmentPath: path,
		Source:       "human_judgment",
		Judgment:     judgment,
	})
}

func writeHumanJudgmentReport(out io.Writer, report humanJudgmentReport) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write human judgment report: %w", err)
	}
	return nil
}
