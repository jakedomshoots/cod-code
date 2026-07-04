package eval

import "errors"

var (
	ErrInvalidTask       = errors.New("invalid eval task")
	ErrInvalidRubric     = errors.New("invalid eval rubric")
	ErrInvalidReport     = errors.New("invalid eval report")
	ErrInvalidCompetitor = errors.New("invalid competitor config")
)

type Task struct {
	ID                     string   `json:"id"`
	Category               string   `json:"category"`
	Title                  string   `json:"title"`
	Objective              string   `json:"objective"`
	RequiredChangedFiles   []string `json:"required_changed_files,omitempty"`
	ForbiddenChangedFiles  []string `json:"forbidden_changed_files,omitempty"`
	RequiredCommands       []string `json:"required_commands,omitempty"`
	RequiredArtifacts      []string `json:"required_artifacts,omitempty"`
	RequiredDiffTerms      []string `json:"required_diff_terms,omitempty"`
	RequiredReportFields   []string `json:"required_report_fields,omitempty"`
	DirtyWorktreeSensitive bool     `json:"dirty_worktree_sensitive,omitempty"`
}

type Rubric struct {
	Path             string
	RequiredSections []string
}

type ScoreRequest struct {
	Task         Task
	ReportPath   string
	WorkspaceDir string
}

type ScoreResult struct {
	TaskID        string        `json:"task_id"`
	Verdict       string        `json:"verdict"`
	Passed        int           `json:"passed"`
	Total         int           `json:"total"`
	Checks        []CheckResult `json:"checks"`
	EvidencePaths []string      `json:"evidence_paths"`
}

type CheckResult struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Evidence string `json:"evidence,omitempty"`
	Message  string `json:"message,omitempty"`
}

type CompetitorConfig struct {
	SchemaVersion int          `json:"schema_version"`
	Competitors   []Competitor `json:"competitors"`
}

type Competitor struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Binary               string   `json:"binary"`
	Homepage             string   `json:"homepage"`
	SetupHint            string   `json:"setup_hint"`
	VersionArgs          []string `json:"version_args"`
	DryRunArgs           []string `json:"dry_run_args"`
	TimeoutSeconds       int      `json:"timeout_seconds"`
	ComparisonDimensions []string `json:"comparison_dimensions"`
}

type ComparisonPlan struct {
	SchemaVersion int                `json:"schema_version"`
	Mode          string             `json:"mode"`
	Results       []ComparisonResult `json:"results"`
}

type ComparisonResult struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Status               string   `json:"status"`
	Binary               string   `json:"binary"`
	Command              []string `json:"command"`
	TimeoutSeconds       int      `json:"timeout_seconds"`
	ComparisonDimensions []string `json:"comparison_dimensions"`
	EvidencePaths        []string `json:"evidence_paths"`
	Note                 string   `json:"note"`
}

type BenchmarkFixtureRequest struct {
	TasksDir   string
	OutputDir  string
	ReportMode string
}

type BenchmarkSummary struct {
	SchemaVersion int                   `json:"schema_version"`
	Mode          string                `json:"mode"`
	TaskCount     int                   `json:"task_count"`
	Passed        int                   `json:"passed"`
	Partial       int                   `json:"partial"`
	Failed        int                   `json:"failed"`
	Skipped       int                   `json:"skipped"`
	Results       []BenchmarkTaskResult `json:"results"`
}

type BenchmarkTaskResult struct {
	TaskID     string `json:"task_id"`
	Verdict    string `json:"verdict"`
	Passed     int    `json:"passed"`
	Total      int    `json:"total"`
	ReportPath string `json:"report_path"`
	ScorePath  string `json:"score_path"`
	LogPath    string `json:"log_path"`
	Reason     string `json:"reason,omitempty"`
}

type CompetitorSmokeRequest struct {
	CompetitorsPath string
	OutputDir       string
	TimeoutSeconds  int
}

type CompetitorSmokeSummary struct {
	SchemaVersion int                     `json:"schema_version"`
	Mode          string                  `json:"mode"`
	Competitors   int                     `json:"competitors"`
	SmokePassed   int                     `json:"smoke_passed"`
	SmokeFailed   int                     `json:"smoke_failed"`
	SetupBlocked  int                     `json:"setup_blocked,omitempty"`
	Skipped       int                     `json:"skipped"`
	SetupActions  string                  `json:"setup_actions,omitempty"`
	Results       []CompetitorSmokeResult `json:"results"`
}

type CompetitorSmokeResult struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Status        string             `json:"status"`
	Binary        string             `json:"binary"`
	ResolvedPath  string             `json:"resolved_path,omitempty"`
	Version       SmokeCommandResult `json:"version"`
	DryRun        SmokeCommandResult `json:"dry_run"`
	EvidencePaths []string           `json:"evidence_paths"`
	SetupHint     string             `json:"setup_hint,omitempty"`
	Note          string             `json:"note"`
}

type SmokeCommandResult struct {
	Command  []string `json:"command"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Error    string   `json:"error,omitempty"`
}

type LocalAgentSuiteRequest struct {
	OutputDir        string
	TimeoutSeconds   int
	Agents           []string
	CEOHarnessBinary string
	Task             string
}

type LocalAgentSuiteSummary struct {
	SchemaVersion    int                   `json:"schema_version"`
	Mode             string                `json:"mode"`
	Task             string                `json:"task"`
	Prompt           string                `json:"prompt"`
	AgentCount       int                   `json:"agent_count"`
	Passed           int                   `json:"passed"`
	Failed           int                   `json:"failed"`
	TimedOut         int                   `json:"timed_out"`
	Skipped          int                   `json:"skipped"`
	Results          []LocalAgentResult    `json:"results"`
	IterationBacklog []LocalAgentIteration `json:"iteration_backlog"`
}

type LocalAgentResult struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	Binary        string   `json:"binary"`
	ResolvedPath  string   `json:"resolved_path,omitempty"`
	Command       []string `json:"command"`
	WorkspaceDir  string   `json:"workspace_dir"`
	ExitCode      int      `json:"exit_code"`
	DurationMS    int64    `json:"duration_ms"`
	OutputMatched bool     `json:"output_matched"`
	FileMatched   bool     `json:"file_matched"`
	ObservedFile  string   `json:"observed_file,omitempty"`
	StdoutPath    string   `json:"stdout_path"`
	StderrPath    string   `json:"stderr_path"`
	AppAfterPath  string   `json:"app_after_path,omitempty"`
	CommandPath   string   `json:"command_path"`
	SetupHint     string   `json:"setup_hint,omitempty"`
	Error         string   `json:"error,omitempty"`
	Note          string   `json:"note"`
}

type LocalAgentIteration struct {
	Priority int    `json:"priority"`
	Area     string `json:"area"`
	Finding  string `json:"finding"`
	NextStep string `json:"next_step"`
	Evidence string `json:"evidence"`
}
