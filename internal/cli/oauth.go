package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ceoharness/internal/config"
)

type oauthProviderSpec struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	CLI            string `json:"cli"`
	AuthType       string `json:"auth_type"`
	TokenStorage   bool   `json:"token_storage"`
	LoginCommand   string `json:"login_command"`
	VersionCommand string `json:"version_command"`
	InitReady      bool   `json:"init_ready"`
	ModelScript    string `json:"model_command_script,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

type oauthProviderStatus struct {
	oauthProviderSpec
	Status    string `json:"status"`
	Version   string `json:"version,omitempty"`
	ErrorKind string `json:"error_kind,omitempty"`
	Error     string `json:"error,omitempty"`
}

type oauthReport struct {
	Command      string                `json:"command"`
	Provider     string                `json:"provider,omitempty"`
	ConfigPath   string                `json:"config_path,omitempty"`
	Created      bool                  `json:"created,omitempty"`
	TokenStorage bool                  `json:"token_storage"`
	Providers    []oauthProviderStatus `json:"providers,omitempty"`
	NextSteps    []string              `json:"next_steps,omitempty"`
}

func runOAuth(ctx context.Context, out io.Writer, opts options) error {
	switch strings.TrimSpace(opts.oauthCommand) {
	case "list":
		return writeOAuthReport(out, buildOAuthListReport(), opts.reportFormat)
	case "doctor":
		report, err := buildOAuthDoctorReport(ctx, opts.oauthProvider)
		if err != nil {
			return err
		}
		return writeOAuthReport(out, report, opts.reportFormat)
	case "init":
		report, err := buildOAuthInitReport(ctx, opts)
		if err != nil {
			return err
		}
		return writeOAuthReport(out, report, opts.reportFormat)
	default:
		return fmt.Errorf("oauth requires list, doctor, or init")
	}
}

func buildOAuthListReport() oauthReport {
	return oauthReport{
		Command:      "list",
		TokenStorage: false,
		Providers:    oauthProviderStatuses(oauthProviderSpecs(), nil),
	}
}

func buildOAuthDoctorReport(ctx context.Context, providerName string) (oauthReport, error) {
	specs, err := filteredOAuthProviderSpecs(providerName)
	if err != nil {
		return oauthReport{}, err
	}
	statuses := make([]oauthProviderStatus, 0, len(specs))
	for _, spec := range specs {
		statuses = append(statuses, doctorOAuthProvider(ctx, spec))
	}
	return oauthReport{
		Command:      "doctor",
		Provider:     strings.TrimSpace(providerName),
		TokenStorage: false,
		Providers:    statuses,
	}, nil
}

func buildOAuthInitReport(ctx context.Context, opts options) (oauthReport, error) {
	providerName := strings.TrimSpace(opts.oauthProvider)
	if providerName == "" {
		return oauthReport{}, fmt.Errorf("oauth init requires a provider name")
	}
	spec, ok := oauthProviderSpecByName(providerName)
	if !ok {
		return oauthReport{}, unknownOAuthProviderError(providerName)
	}
	if !spec.InitReady || strings.TrimSpace(spec.ModelScript) == "" {
		return oauthReport{}, fmt.Errorf("oauth provider %q does not have a built-in model-command wrapper yet", providerName)
	}
	scriptPath, err := findRepoFile(filepath.Join("scripts", spec.ModelScript), spec.DisplayName+" model command script")
	if err != nil {
		return oauthReport{}, err
	}
	cfg := config.Config{
		Providers: map[string]config.Provider{
			"main": {
				ModelCommand: []string{"sh", scriptPath},
			},
		},
		CEOProvider: "main",
		ProviderPolicy: config.ProviderPolicy{
			DefaultProvider:  "main",
			FallbackProvider: "main",
		},
		ModelCommandTimeoutMS: opts.modelCommandTimeoutMS,
	}
	path, err := config.CreateWorkspace(ctx, opts.workspaceDir, cfg)
	if err != nil {
		return oauthReport{}, err
	}
	return oauthReport{
		Command:      "init",
		Provider:     spec.Name,
		ConfigPath:   path,
		Created:      true,
		TokenStorage: false,
		Providers:    oauthProviderStatuses([]oauthProviderSpec{spec}, map[string]string{spec.Name: "configured"}),
		NextSteps: []string{
			spec.LoginCommand,
			"ceo-packet oauth doctor " + spec.Name + " --format text",
			"ceo-packet --workspace " + shellQuote(opts.workspaceDir) + " --doctor-provider main --format text",
		},
	}, nil
}

func oauthProviderSpecs() []oauthProviderSpec {
	return []oauthProviderSpec{
		{
			Name:           "kimi",
			DisplayName:    "Kimi CLI",
			CLI:            "kimi",
			AuthType:       "cli_oauth",
			TokenStorage:   false,
			LoginCommand:   "kimi login",
			VersionCommand: "kimi --version",
			InitReady:      true,
			ModelScript:    "kimi-model-command.sh",
			Notes:          "Uses the local Kimi CLI login; the harness stores no OAuth token.",
		},
		{
			Name:           "codex",
			DisplayName:    "Codex CLI",
			CLI:            "codex",
			AuthType:       "cli_oauth",
			TokenStorage:   false,
			LoginCommand:   "codex login",
			VersionCommand: "codex --version",
			InitReady:      true,
			ModelScript:    "codex-model-command.sh",
			Notes:          "Uses the local Codex CLI login; the harness stores no OAuth token.",
		},
		{
			Name:           "claude",
			DisplayName:    "Claude Code",
			CLI:            "claude",
			AuthType:       "cli_oauth",
			TokenStorage:   false,
			LoginCommand:   "claude",
			VersionCommand: "claude --version",
			InitReady:      true,
			ModelScript:    "claude-model-command.sh",
			Notes:          "Uses the local Claude Code login; the harness stores no OAuth token.",
		},
		{
			Name:           "opencode",
			DisplayName:    "OpenCode",
			CLI:            "opencode",
			AuthType:       "cli_oauth",
			TokenStorage:   false,
			LoginCommand:   "opencode providers",
			VersionCommand: "opencode --version",
			InitReady:      true,
			ModelScript:    "opencode-model-command.sh",
			Notes:          "Uses the local OpenCode provider login; the harness stores no OAuth token.",
		},
		{
			Name:           "goose",
			DisplayName:    "Goose",
			CLI:            "goose",
			AuthType:       "cli_oauth",
			TokenStorage:   false,
			LoginCommand:   "goose configure",
			VersionCommand: "goose --version",
			InitReady:      true,
			ModelScript:    "goose-model-command.sh",
			Notes:          "Uses the local Goose provider login; the harness stores no OAuth token.",
		},
	}
}

func oauthProviderStatuses(specs []oauthProviderSpec, statuses map[string]string) []oauthProviderStatus {
	rows := make([]oauthProviderStatus, 0, len(specs))
	for _, spec := range specs {
		status := statuses[spec.Name]
		if status == "" {
			if spec.InitReady {
				status = "init_ready"
			} else {
				status = "custom_adapter_required"
			}
		}
		rows = append(rows, oauthProviderStatus{
			oauthProviderSpec: spec,
			Status:            status,
		})
	}
	return rows
}

func filteredOAuthProviderSpecs(providerName string) ([]oauthProviderSpec, error) {
	name := strings.TrimSpace(providerName)
	if name == "" {
		return oauthProviderSpecs(), nil
	}
	spec, ok := oauthProviderSpecByName(name)
	if !ok {
		return nil, unknownOAuthProviderError(name)
	}
	return []oauthProviderSpec{spec}, nil
}

func oauthProviderSpecByName(providerName string) (oauthProviderSpec, bool) {
	name := strings.ToLower(strings.TrimSpace(providerName))
	for _, spec := range oauthProviderSpecs() {
		if spec.Name == name {
			return spec, true
		}
	}
	return oauthProviderSpec{}, false
}

func unknownOAuthProviderError(providerName string) error {
	return fmt.Errorf("unknown oauth provider %q", providerName)
}

func doctorOAuthProvider(ctx context.Context, spec oauthProviderSpec) oauthProviderStatus {
	status := oauthProviderStatus{oauthProviderSpec: spec}
	path, err := exec.LookPath(spec.CLI)
	if err != nil {
		status.Status = "missing_cli"
		status.ErrorKind = "missing_cli"
		status.Error = spec.CLI + " not found on PATH"
		return status
	}
	version, err := runOAuthVersion(ctx, path)
	if err != nil {
		status.Status = "version_failed"
		status.ErrorKind = "command_failed"
		status.Error = err.Error()
		return status
	}
	status.Status = "ready"
	status.Version = strings.TrimSpace(version)
	return status
}

func runOAuthVersion(ctx context.Context, path string) (string, error) {
	runCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, path, "--version")
	output, err := cmd.CombinedOutput()
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("version command timed out")
	}
	if err != nil {
		return "", fmt.Errorf("version command failed: %w", err)
	}
	return string(output), nil
}

func writeOAuthReport(out io.Writer, report oauthReport, format reportFormat) error {
	switch format {
	case "", reportFormatJSON:
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case reportFormatText:
		_, err := io.WriteString(out, renderOAuthText(report))
		return err
	case reportFormatEvents:
		return fmt.Errorf("--format events is only available for run reports")
	default:
		return fmt.Errorf(reportFormatGuidance)
	}
}

func renderOAuthText(report oauthReport) string {
	var builder strings.Builder
	builder.WriteString("OAuth: ")
	builder.WriteString(report.Command)
	builder.WriteString("\nToken storage: none\n")
	if report.ConfigPath != "" {
		builder.WriteString("Config: ")
		builder.WriteString(report.ConfigPath)
		builder.WriteString("\n")
	}
	for _, provider := range report.Providers {
		builder.WriteString("- ")
		builder.WriteString(provider.Name)
		builder.WriteString(": ")
		builder.WriteString(provider.Status)
		if provider.Version != "" {
			builder.WriteString(" (")
			builder.WriteString(provider.Version)
			builder.WriteString(")")
		}
		if provider.ModelScript != "" {
			builder.WriteString(" via ")
			builder.WriteString(provider.ModelScript)
		}
		builder.WriteString("\n")
	}
	if len(report.NextSteps) > 0 {
		builder.WriteString("Next:\n")
		for _, step := range report.NextSteps {
			builder.WriteString("- ")
			builder.WriteString(step)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
