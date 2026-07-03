package history

type RunLedger struct {
	Owner                string   `json:"owner"`
	Verdict              string   `json:"verdict"`
	NextAction           string   `json:"next_action"`
	VerificationStatus   string   `json:"verification_status"`
	RequiredCheckCount   int      `json:"required_check_count"`
	CheckAttemptCount    int      `json:"check_attempt_count"`
	ChangedFileCount     int      `json:"changed_file_count"`
	ChangedFiles         []string `json:"changed_files,omitempty"`
	ProviderRouteCount   int      `json:"provider_route_count"`
	ProviderRouteReasons []string `json:"provider_route_reasons,omitempty"`
}
