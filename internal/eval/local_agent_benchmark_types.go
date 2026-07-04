package eval

type LocalAgentBenchmarkRequest struct {
	TasksDir                          string
	OutputDir                         string
	TimeoutSeconds                    int
	Agents                            []string
	CEOHarnessBinary                  string
	CEOBenchmarkMode                  string
	CEOBenchmarkModelCommand          []string
	CEOBenchmarkProviderName          string
	CEOBenchmarkProviderPreset        string
	CEOBenchmarkProviderModel         string
	CEOBenchmarkProviderAPIKeyEnv     string
	CEOBenchmarkProviderMaxOutputToks int
	BenchmarkTaskID                   string
	RepeatCount                       int
	Concurrency                       int
	TimeoutRetries                    int
	ResultRetries                     int
	AgentTimeoutSeconds               map[string]int
	AgentModels                       map[string]string
}

type LocalAgentBenchmarkSummary struct {
	SchemaVersion      int                         `json:"schema_version"`
	Mode               string                      `json:"mode"`
	TaskID             string                      `json:"task_id"`
	TaskTitle          string                      `json:"task_title"`
	TaskIDs            []string                    `json:"task_ids,omitempty"`
	TaskCount          int                         `json:"task_count"`
	RepeatCount        int                         `json:"repeat_count"`
	Concurrency        int                         `json:"concurrency"`
	TimeoutRetries     int                         `json:"timeout_retries"`
	ResultRetries      int                         `json:"result_retries,omitempty"`
	AgentTimeouts      map[string]int              `json:"agent_timeouts,omitempty"`
	AgentModels        map[string]string           `json:"agent_models,omitempty"`
	RunCount           int                         `json:"run_count"`
	AgentCount         int                         `json:"agent_count"`
	Passed             int                         `json:"passed"`
	Partial            int                         `json:"partial"`
	Failed             int                         `json:"failed"`
	TimedOut           int                         `json:"timed_out"`
	SetupBlocked       int                         `json:"setup_blocked,omitempty"`
	Skipped            int                         `json:"skipped"`
	IncompleteEvidence int                         `json:"incomplete_evidence"`
	Results            []LocalAgentBenchmarkResult `json:"results"`
	IterationBacklog   []LocalAgentIteration       `json:"iteration_backlog"`
}

type LocalAgentBenchmarkResult struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	TaskID            string         `json:"task_id"`
	TaskTitle         string         `json:"task_title"`
	Attempt           int            `json:"attempt"`
	RunAttempt        int            `json:"run_attempt,omitempty"`
	PriorAttempts     []RetryAttempt `json:"prior_attempts,omitempty"`
	Status            string         `json:"status"`
	Binary            string         `json:"binary"`
	ResolvedPath      string         `json:"resolved_path,omitempty"`
	Command           []string       `json:"command"`
	WorkspaceDir      string         `json:"workspace_dir"`
	ExitCode          int            `json:"exit_code"`
	DurationMS        int64          `json:"duration_ms"`
	ScoreVerdict      string         `json:"score_verdict,omitempty"`
	PassedChecks      int            `json:"passed_checks"`
	TotalChecks       int            `json:"total_checks"`
	FailedScoreChecks []CheckResult  `json:"failed_score_checks,omitempty"`
	EvidenceStatus    string         `json:"evidence_status"`
	ChangedFiles      []string       `json:"changed_files"`
	ExtraChangedFiles []string       `json:"extra_changed_files,omitempty"`
	ReportPath        string         `json:"report_path"`
	ScorePath         string         `json:"score_path"`
	CommandPath       string         `json:"command_path"`
	StdoutPath        string         `json:"stdout_path"`
	StderrPath        string         `json:"stderr_path"`
	DiffPath          string         `json:"diff_path"`
	ChangedFilesPath  string         `json:"changed_files_path"`
	GitBeforePath     string         `json:"git_status_before_path"`
	GitAfterPath      string         `json:"git_status_after_path"`
	TimingPath        string         `json:"timing_path"`
	SetupHint         string         `json:"setup_hint,omitempty"`
	Error             string         `json:"error,omitempty"`
	Note              string         `json:"note"`
}

type RetryAttempt struct {
	RunAttempt     int    `json:"run_attempt"`
	Status         string `json:"status"`
	EvidenceStatus string `json:"evidence_status"`
	ScorePath      string `json:"score_path"`
	StdoutPath     string `json:"stdout_path"`
	StderrPath     string `json:"stderr_path"`
	TimingPath     string `json:"timing_path"`
	Note           string `json:"note"`
}
