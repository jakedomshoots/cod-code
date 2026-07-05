package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/ceo"
	"ceoharness/internal/history"
)

type tuiReport struct {
	Workspace string   `json:"workspace"`
	Config    string   `json:"config_path,omitempty"`
	Snapshot  bool     `json:"snapshot"`
	Model     tuiModel `json:"model"`
}

func runTUI(ctx context.Context, in io.Reader, out io.Writer, opts options) error {
	report, err := buildTUIReport(ctx, opts)
	if err != nil {
		return err
	}
	if opts.tuiSnapshot || (opts.reportFormat != "" && opts.reportFormat != reportFormatText) {
		return writeTUIReport(out, report, opts.reportFormat)
	}
	return runInteractiveTUI(in, out, report.Model)
}

func runInteractiveTUI(in io.Reader, out io.Writer, model tuiModel) error {
	if _, err := io.WriteString(out, model.render()); err != nil {
		return err
	}
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}
		if command == "q" || command == "quit" || command == "exit" {
			return nil
		}
		var action string
		model, action = model.applyKey(command)
		if action != "" {
			if _, err := fmt.Fprintf(out, "Action dispatched: %s\n", action); err != nil {
				return err
			}
			continue
		}
		if command == "enter" {
			if _, err := io.WriteString(out, "No action for selected job.\n"); err != nil {
				return err
			}
			continue
		}
		if _, err := io.WriteString(out, model.render()); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read tui input: %w", err)
	}
	return nil
}

func buildTUIReport(ctx context.Context, opts options) (tuiReport, error) {
	workspace := strings.TrimSpace(opts.workspaceDir)
	if workspace == "" {
		workspace = "."
	}
	store, entries, err := readHistoryEntries(ctx, historyQuery{workspaceDir: workspace})
	if err != nil {
		return tuiReport{}, err
	}
	judgments, err := readHumanJudgmentsForHistory(ctx, store, entries)
	if err != nil {
		return tuiReport{}, err
	}
	rows := buildReviewQueueRows(reviewQueueRowsRequest{
		Entries:      entries,
		Judgments:    judgments,
		WorkspaceDir: workspace,
	})
	model := newTUIModel(workspace, entries, rows, history.AggregateProviderHealth(entries))
	addTUISnapshotDetails(ctx, store, &model)
	return tuiReport{
		Workspace: workspace,
		Config:    workspaceConfigPath(workspace),
		Snapshot:  opts.tuiSnapshot,
		Model:     model,
	}, nil
}

func writeTUIReport(out io.Writer, report tuiReport, format reportFormat) error {
	switch format {
	case "", reportFormatText:
		_, err := io.WriteString(out, renderTUIText(report))
		return err
	case reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderTUIText(report tuiReport) string {
	return report.Model.render()
}

func addTUISnapshotDetails(ctx context.Context, store history.Store, model *tuiModel) {
	for index := range model.jobs {
		report, err := readTUISnapshotReport(ctx, store, model.jobs[index].id)
		if err != nil {
			if !errors.Is(err, history.ErrEntryNotFound) {
				model.jobs[index].snapshotNote = "saved report unreadable"
			}
			continue
		}
		if len(report.PatchPreviews) > 0 {
			model.jobs[index].patchPreview = report.PatchPreviews[0].Path
		} else if len(report.PatchResults) > 0 {
			model.jobs[index].patchPreview = report.PatchResults[0].Path
		}
		if len(report.CheckResults) > 0 {
			check := report.CheckResults[len(report.CheckResults)-1]
			model.jobs[index].checkOutput = strings.Join(check.Argv, " ") + " " + check.Status
			if model.jobs[index].checkOutput == " " {
				model.jobs[index].checkOutput = check.Status
			}
		}
	}
}

func readTUISnapshotReport(ctx context.Context, store history.Store, jobID string) (ceo.Report, error) {
	snapshot, err := store.ReadReportSnapshotWithMetadata(ctx, jobID)
	if err != nil {
		return ceo.Report{}, err
	}
	payload := snapshot.Payload
	if snapshot.Metadata.Legacy {
		payload, err = reportPayloadWithCompatibility(snapshot)
		if err != nil {
			return ceo.Report{}, err
		}
	}
	var report ceo.Report
	if err := json.Unmarshal(payload, &report); err != nil {
		return ceo.Report{}, fmt.Errorf("decode tui snapshot report: %w", err)
	}
	return report, nil
}
