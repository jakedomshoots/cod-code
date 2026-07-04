package eval

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
)

type cliOptions struct {
	tasksDir                          string
	rubricPath                        string
	competitorsPath                   string
	outputDir                         string
	timeoutSeconds                    int
	listTasks                         bool
	rubric                            bool
	validateCompetitors               bool
	comparisonPlan                    bool
	comparisonSmoke                   bool
	benchmarkFixtures                 bool
	localAgentSuite                   bool
	localAgentBenchmark               bool
	reportPath                        string
	taskID                            string
	workspace                         string
	localAgents                       string
	localAgentTask                    string
	localAgentBenchmarkTask           string
	localAgentBenchmarkRepeat         int
	localAgentBenchmarkConcurrency    int
	localAgentBenchmarkTimeoutRetries int
	localAgentBenchmarkAgentTimeouts  string
	localAgentBenchmarkAgentModels    string
	ceoHarnessBinary                  string
	ceoBenchmarkMode                  string
	ceoBenchmarkModelCommand          string
	ceoBenchmarkProviderName          string
	ceoBenchmarkProviderPreset        string
	ceoBenchmarkProviderModel         string
	ceoBenchmarkProviderAPIKeyEnv     string
	ceoBenchmarkProviderMaxOutputToks int
}

func RunCLI(ctx context.Context, out io.Writer, errOut io.Writer, args []string) error {
	opts, err := parseCLI(args, errOut)
	if err != nil {
		return err
	}
	if opts.listTasks {
		return runListTasks(ctx, out, opts.tasksDir)
	}
	if opts.rubric {
		return runRubric(out, opts.rubricPath)
	}
	if opts.validateCompetitors {
		return runValidateCompetitors(out, opts.competitorsPath)
	}
	if opts.comparisonPlan {
		return runComparisonPlan(ctx, out, opts.competitorsPath)
	}
	if opts.comparisonSmoke {
		return runComparisonSmoke(ctx, out, opts)
	}
	if opts.benchmarkFixtures {
		return runBenchmarkFixtures(ctx, out, opts)
	}
	if opts.localAgentSuite {
		return runLocalAgentSuite(ctx, out, opts)
	}
	if opts.localAgentBenchmark {
		return runLocalAgentBenchmarkCLI(ctx, out, opts)
	}
	if opts.reportPath != "" {
		return runScore(ctx, out, opts)
	}
	return fmt.Errorf("usage: ceo-eval --list | --rubric | --validate-competitors | --comparison-plan | --comparison-smoke | --benchmark-fixtures | --local-agent-suite | --local-agent-benchmark | --task <id> --report <path>")
}

func parseCLI(args []string, errOut io.Writer) (cliOptions, error) {
	opts := cliOptions{}
	flags := flag.NewFlagSet("ceo-eval", flag.ContinueOnError)
	flags.SetOutput(errOut)
	flags.StringVar(&opts.tasksDir, "tasks", "evals/tasks", "task spec directory")
	flags.StringVar(&opts.rubricPath, "rubric-path", "evals/rubric.md", "rubric markdown path")
	flags.StringVar(&opts.competitorsPath, "competitors", "evals/competitors.json", "competitor config path")
	flags.StringVar(&opts.outputDir, "output-dir", ".omo/evidence/eval", "directory for generated eval evidence")
	flags.IntVar(&opts.timeoutSeconds, "timeout-seconds", 15, "timeout for local smoke commands")
	flags.BoolVar(&opts.listTasks, "list", false, "list task ids")
	flags.BoolVar(&opts.rubric, "rubric", false, "validate rubric")
	flags.BoolVar(&opts.validateCompetitors, "validate-competitors", false, "validate competitor config")
	flags.BoolVar(&opts.comparisonPlan, "comparison-plan", false, "print plan-only competitor result placeholders")
	flags.BoolVar(&opts.comparisonSmoke, "comparison-smoke", false, "run local competitor version and dry-run smoke commands")
	flags.BoolVar(&opts.benchmarkFixtures, "benchmark-fixtures", false, "score all tasks against deterministic fixture reports")
	registerLocalAgentFlags(flags, &opts)
	flags.StringVar(&opts.reportPath, "report", "", "saved report JSON path")
	flags.StringVar(&opts.taskID, "task", "", "task id to score")
	flags.StringVar(&opts.workspace, "workspace", ".", "workspace root for evidence paths and git status")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}
	return opts, nil
}

func runListTasks(ctx context.Context, out io.Writer, tasksDir string) error {
	tasks, err := LoadTasks(ctx, tasksDir)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if _, err := fmt.Fprintln(out, task.ID); err != nil {
			return fmt.Errorf("write task id: %w", err)
		}
	}
	return nil
}

func runValidateCompetitors(out io.Writer, competitorsPath string) error {
	config, err := LoadCompetitors(competitorsPath)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "competitors_valid=true count=%d path=%s\n", len(config.Competitors), competitorsPath); err != nil {
		return fmt.Errorf("write competitor validation result: %w", err)
	}
	return nil
}

func runComparisonPlan(ctx context.Context, out io.Writer, competitorsPath string) error {
	config, err := LoadCompetitors(competitorsPath)
	if err != nil {
		return err
	}
	plan, err := BuildComparisonPlan(ctx, config)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(plan); err != nil {
		return fmt.Errorf("write comparison plan: %w", err)
	}
	return nil
}

func runComparisonSmoke(ctx context.Context, out io.Writer, opts cliOptions) error {
	summary, err := RunCompetitorSmoke(ctx, CompetitorSmokeRequest{
		CompetitorsPath: opts.competitorsPath,
		OutputDir:       opts.outputDir,
		TimeoutSeconds:  opts.timeoutSeconds,
	})
	if err != nil {
		return err
	}
	return writeIndentedJSON(out, summary)
}

func runBenchmarkFixtures(ctx context.Context, out io.Writer, opts cliOptions) error {
	summary, err := RunBenchmarkFixtures(ctx, BenchmarkFixtureRequest{
		TasksDir:   opts.tasksDir,
		OutputDir:  opts.outputDir,
		ReportMode: "deterministic_fixture_scoring",
	})
	if err != nil {
		return err
	}
	return writeIndentedJSON(out, summary)
}

func runLocalAgentSuite(ctx context.Context, out io.Writer, opts cliOptions) error {
	summary, err := RunLocalAgentSuite(ctx, LocalAgentSuiteRequest{
		OutputDir:        opts.outputDir,
		TimeoutSeconds:   opts.timeoutSeconds,
		Agents:           splitCSV(opts.localAgents),
		CEOHarnessBinary: opts.ceoHarnessBinary,
		Task:             opts.localAgentTask,
	})
	if err != nil {
		return err
	}
	return writeIndentedJSON(out, summary)
}

func runLocalAgentBenchmarkCLI(ctx context.Context, out io.Writer, opts cliOptions) error {
	modelCommand, err := parseCEOBenchmarkModelCommand(opts.ceoBenchmarkModelCommand)
	if err != nil {
		return err
	}
	agentTimeouts, err := parseLocalAgentBenchmarkAgentTimeouts(opts.localAgentBenchmarkAgentTimeouts)
	if err != nil {
		return err
	}
	agentModels, err := parseLocalAgentBenchmarkAgentModels(opts.localAgentBenchmarkAgentModels)
	if err != nil {
		return err
	}
	summary, err := RunLocalAgentBenchmark(ctx, LocalAgentBenchmarkRequest{
		TasksDir:                          opts.tasksDir,
		OutputDir:                         opts.outputDir,
		TimeoutSeconds:                    opts.timeoutSeconds,
		Agents:                            splitCSV(opts.localAgents),
		CEOHarnessBinary:                  opts.ceoHarnessBinary,
		CEOBenchmarkMode:                  opts.ceoBenchmarkMode,
		CEOBenchmarkModelCommand:          modelCommand,
		CEOBenchmarkProviderName:          opts.ceoBenchmarkProviderName,
		CEOBenchmarkProviderPreset:        opts.ceoBenchmarkProviderPreset,
		CEOBenchmarkProviderModel:         opts.ceoBenchmarkProviderModel,
		CEOBenchmarkProviderAPIKeyEnv:     opts.ceoBenchmarkProviderAPIKeyEnv,
		CEOBenchmarkProviderMaxOutputToks: opts.ceoBenchmarkProviderMaxOutputToks,
		BenchmarkTaskID:                   opts.localAgentBenchmarkTask,
		RepeatCount:                       opts.localAgentBenchmarkRepeat,
		Concurrency:                       opts.localAgentBenchmarkConcurrency,
		TimeoutRetries:                    opts.localAgentBenchmarkTimeoutRetries,
		AgentTimeoutSeconds:               agentTimeouts,
		AgentModels:                       agentModels,
	})
	if err != nil {
		return err
	}
	return writeIndentedJSON(out, summary)
}

func runRubric(out io.Writer, rubricPath string) error {
	rubric, err := LoadRubric(rubricPath)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "rubric_valid=true path=%s sections=%d\n", rubric.Path, len(rubric.RequiredSections)); err != nil {
		return fmt.Errorf("write rubric result: %w", err)
	}
	return nil
}

func runScore(ctx context.Context, out io.Writer, opts cliOptions) error {
	tasks, err := LoadTasks(ctx, opts.tasksDir)
	if err != nil {
		return err
	}
	task, err := FindTask(tasks, opts.taskID)
	if err != nil {
		return err
	}
	result, err := ScoreReport(ctx, ScoreRequest{
		Task:         task,
		ReportPath:   opts.reportPath,
		WorkspaceDir: opts.workspace,
	})
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("write score result: %w", err)
	}
	return nil
}

func writeIndentedJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	return nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean != "" {
			values = append(values, clean)
		}
	}
	return values
}
