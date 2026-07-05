package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ceoharness/internal/config"
)

type demoRepoReport struct {
	Status     string   `json:"status"`
	Path       string   `json:"path"`
	ConfigPath string   `json:"config_path"`
	Files      []string `json:"files"`
	NextSteps  []string `json:"next_steps"`
}

func runInitDemoRepo(ctx context.Context, out io.Writer, opts options) error {
	report, err := buildDemoRepo(ctx, opts.initDemoRepoDir)
	if err != nil {
		return err
	}
	return writeDemoRepoReport(out, report, opts.reportFormat)
}

func buildDemoRepo(ctx context.Context, root string) (demoRepoReport, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return demoRepoReport{}, fmt.Errorf("create demo repo: %w", err)
	}
	files := []string{"README.md", "app.txt", "Makefile"}
	if err := writeNewFile(filepath.Join(root, "README.md"), demoReadme()); err != nil {
		return demoRepoReport{}, err
	}
	if err := writeNewFile(filepath.Join(root, "app.txt"), "hello old\n"); err != nil {
		return demoRepoReport{}, err
	}
	if err := writeNewFile(filepath.Join(root, "Makefile"), "test:\n\tgrep -q 'hello new' app.txt\n"); err != nil {
		return demoRepoReport{}, err
	}
	adapters, err := exampleAdapterCommands()
	if err != nil {
		return demoRepoReport{}, err
	}
	configPath, err := config.CreateWorkspace(ctx, root, config.Config{
		ModelCommand:               adapters.ModelCommand,
		CEOModelCommand:            adapters.CEOModelCommand,
		ResearchCommand:            adapters.ResearchCommand,
		CheckCommand:               []string{"make", "test"},
		RequireChecks:              true,
		MaxSubagents:               3,
		MaxContextBytes:            2048,
		MaxSubagentOutputBytes:     800,
		SubagentConcurrency:        1,
		SubagentAttempts:           2,
		NoProgressStop:             2,
		MaxCEOIterations:           4,
		CEORevisionAttempts:        1,
		CheckAttempts:              1,
		CheckBackoffMS:             0,
		MinSubagentConfidence:      0.6,
		WorkspaceBriefMaxFiles:     20,
		ToolCommandTimeoutMS:       30000,
		ModelCommandTimeoutMS:      30000,
		JobTimeoutMS:               120000,
		ProviderCostBudgetMicroUSD: 0,
	})
	if err != nil {
		return demoRepoReport{}, err
	}
	return demoRepoReport{
		Status:     "created",
		Path:       root,
		ConfigPath: configPath,
		Files:      append(files, config.WorkspaceConfigName),
		NextSteps: []string{
			"cod --workspace " + workspaceArg(root) + " --dry-run --replace app.txt old new -- " + strconvQuote("Patch demo app"),
			"cod --workspace " + workspaceArg(root) + " --approve-preview <digest> --replace app.txt old new -- " + strconvQuote("Patch demo app"),
		},
	}, nil
}

func writeNewFile(path string, content string) (err error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", path, closeErr)
		}
	}()
	if _, err := io.WriteString(file, content); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func demoReadme() string {
	return "# Cod Code Golden Demo\n\nThis repo is intentionally tiny. Patch `app.txt` from `hello old` to `hello new`, then run `make test`.\n"
}

func writeDemoRepoReport(out io.Writer, report demoRepoReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := io.WriteString(out, renderDemoRepoText(report))
		return err
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderDemoRepoText(report demoRepoReport) string {
	var builder strings.Builder
	builder.WriteString("Golden demo repo: ")
	builder.WriteString(report.Status)
	builder.WriteString("\nPath: ")
	builder.WriteString(report.Path)
	builder.WriteString("\nConfig: ")
	builder.WriteString(report.ConfigPath)
	builder.WriteString("\nFiles: ")
	builder.WriteString(strings.Join(report.Files, ", "))
	builder.WriteString("\nNext:\n")
	for _, step := range report.NextSteps {
		builder.WriteString("- ")
		builder.WriteString(step)
		builder.WriteString("\n")
	}
	return builder.String()
}

func strconvQuote(value string) string {
	return fmt.Sprintf("%q", value)
}
