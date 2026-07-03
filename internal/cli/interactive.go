package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/ceo"
)

const maxInteractiveTurns = 3

func optionsWithInteractiveFormat(opts options) (options, error) {
	if opts.reportFormatSet && opts.reportFormat != reportFormatText {
		return options{}, fmt.Errorf("--interactive requires --format text")
	}
	opts.reportFormat = reportFormatText
	opts.reportFormatSet = true
	return opts, nil
}

func runInteractive(ctx context.Context, in io.Reader, out io.Writer, opts options) error {
	scanner := bufio.NewScanner(in)
	current := opts
	for turn := 0; turn < maxInteractiveTurns; turn++ {
		report, err := buildRunReport(ctx, current)
		if err != nil {
			return err
		}
		if err := writeInteractiveReport(out, current, report); err != nil {
			return err
		}
		if report.Verdict != "needs_input" {
			return verdictError(report)
		}
		next, err := nextInteractiveOptions(scanner, out, current, report)
		if err != nil {
			return err
		}
		current = next
	}
	return fmt.Errorf("interactive turn limit reached")
}

func writeInteractiveReport(out io.Writer, opts options, report ceo.Report) error {
	return writeRunReport(out, reportOutputRequest{
		Report:       report,
		Format:       opts.reportFormat,
		WorkspaceDir: opts.workspaceDir,
	})
}

func nextInteractiveOptions(scanner *bufio.Scanner, out io.Writer, opts options, report ceo.Report) (options, error) {
	if strings.TrimSpace(opts.workspaceDir) == "" || strings.TrimSpace(report.JobID) == "" {
		return options{}, fmt.Errorf("--interactive requires --workspace when a run needs input")
	}
	questions := reportQuestions(report)
	if len(questions) == 0 {
		return options{}, ErrVerdictNeedsInput
	}
	answers, err := readInteractiveAnswers(scanner, out, questions)
	if err != nil {
		return options{}, err
	}
	next := opts
	next.task = ""
	next.rerunJobID = ""
	next.resumeJobID = report.JobID
	next.resumeAnswers = answers
	next.resumeContext = nil
	return next, nil
}

func readInteractiveAnswers(scanner *bufio.Scanner, out io.Writer, questions []string) ([]string, error) {
	answers := make([]string, 0, len(questions))
	for _, question := range questions {
		if _, err := fmt.Fprintf(out, "\n%s\n> ", question); err != nil {
			return nil, fmt.Errorf("write interactive prompt: %w", err)
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("read interactive answer: %w", err)
			}
			return nil, fmt.Errorf("interactive answer is required")
		}
		answer := strings.TrimSpace(scanner.Text())
		if answer == "" {
			return nil, fmt.Errorf("interactive answer is required")
		}
		answers = append(answers, answer)
	}
	return answers, nil
}
