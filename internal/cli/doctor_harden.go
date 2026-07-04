package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ceoharness/internal/config"
)

const (
	doctorStatusPass    = "pass"
	doctorStatusFail    = "fail"
	doctorStatusBlocked = "blocked"
	doctorStatusSkipped = "skipped"

	doctorRequired = "required"
	doctorOptional = "optional"
)

type doctorLookupFunc func(string) (string, error)

func appendDoctorCheck(report *doctorReport, check doctorCheck) {
	report.Checks = append(report.Checks, check)
	if doctorCheckFails(check) {
		report.Status = doctorStatusFail
		report.Summary = "harness doctor failed"
	}
}

func doctorCheckFails(check doctorCheck) bool {
	return check.Status == doctorStatusFail || check.Status == doctorStatusBlocked
}

func doctorToolChecks(lookup doctorLookupFunc) []doctorCheck {
	requiredTools := []string{"go"}
	optionalStrictTools := []string{"gofumpt", "golangci-lint", "nilaway"}
	optionalProviderCLIs := []string{"codex", "kimi", "opencode", "pi"}
	checks := make([]doctorCheck, 0, len(requiredTools)+len(optionalStrictTools)+len(optionalProviderCLIs))
	for _, name := range requiredTools {
		checks = append(checks, doctorToolCheck("tool."+name, name, doctorRequired, lookup))
	}
	for _, name := range optionalStrictTools {
		checks = append(checks, doctorToolCheck("tool."+name, name, doctorOptional, lookup))
	}
	for _, name := range optionalProviderCLIs {
		checks = append(checks, doctorToolCheck("provider_cli."+name, name, doctorOptional, lookup))
	}
	return checks
}

func doctorToolCheck(checkName string, binName string, requirement string, lookup doctorLookupFunc) doctorCheck {
	path, err := lookup(binName)
	if err == nil && strings.TrimSpace(path) != "" {
		return doctorCheck{Name: checkName, Status: doctorStatusPass, Requirement: requirement, Source: "path", Path: path}
	}
	status := doctorStatusBlocked
	if requirement == doctorOptional {
		status = doctorStatusSkipped
	}
	return doctorCheck{
		Name:        checkName,
		Status:      status,
		Requirement: requirement,
		Source:      "path",
		Error:       fmt.Sprintf("%s not found on PATH", binName),
		Guidance:    fmt.Sprintf("install %s only if this workflow uses it", binName),
	}
}

func runLocalDoctorChecks(ctx context.Context, opts options) []doctorCheck {
	checks := append([]doctorCheck{}, doctorToolChecks(exec.LookPath)...)
	checks = append(checks, installScriptDoctorChecks()...)
	checks = append(checks, workspaceDoctorChecks(ctx, opts)...)
	return checks
}

func installScriptDoctorChecks() []doctorCheck {
	return []doctorCheck{
		scriptDoctorCheck("install_script", filepath.Join("scripts", "install-local.sh")),
		scriptDoctorCheck("release_script", filepath.Join("scripts", "release-local.sh")),
	}
}

func scriptDoctorCheck(name string, relativePath string) doctorCheck {
	path, err := findRepoFile(relativePath, name)
	if err != nil {
		if _, repoErr := findRepoFile("go.mod", "repo root"); repoErr != nil {
			return doctorCheck{
				Name:        name,
				Status:      doctorStatusSkipped,
				Requirement: doctorOptional,
				Error:       err.Error(),
				Guidance:    "source checkout script check skipped for binary-only install",
			}
		}
		return doctorCheck{Name: name, Status: doctorStatusBlocked, Requirement: doctorRequired, Error: err.Error()}
	}
	info, err := os.Stat(path)
	if err != nil {
		return doctorCheck{Name: name, Status: doctorStatusBlocked, Requirement: doctorRequired, Path: path, Error: err.Error()}
	}
	if info.Mode()&0o111 == 0 {
		return doctorCheck{Name: name, Status: doctorStatusBlocked, Requirement: doctorRequired, Path: path, Error: "script is not executable"}
	}
	return doctorCheck{Name: name, Status: doctorStatusPass, Requirement: doctorRequired, Path: path}
}

func workspaceDoctorChecks(ctx context.Context, opts options) []doctorCheck {
	root := strings.TrimSpace(opts.workspaceDir)
	if root == "" {
		return nil
	}
	checks := []doctorCheck{}
	info, err := os.Stat(root)
	if err != nil {
		return []doctorCheck{{Name: "workspace", Status: doctorStatusBlocked, Requirement: doctorRequired, Path: root, Error: err.Error()}}
	}
	if !info.IsDir() {
		return []doctorCheck{{Name: "workspace", Status: doctorStatusBlocked, Requirement: doctorRequired, Path: root, Error: "workspace is not a directory"}}
	}
	checks = append(checks, doctorCheck{Name: "workspace", Status: doctorStatusPass, Requirement: doctorRequired, Path: root})
	configPath := filepath.Join(root, config.WorkspaceConfigName)
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		checks = append(checks, doctorCheck{Name: "workspace_config", Status: doctorStatusSkipped, Requirement: doctorOptional, Path: configPath, Guidance: "run config init when this workspace needs saved provider/check policy"})
		return checks
	}
	if err != nil {
		checks = append(checks, doctorCheck{Name: "workspace_config", Status: doctorStatusBlocked, Requirement: doctorRequired, Path: configPath, Error: err.Error()})
		return checks
	}
	if _, err := config.LoadWorkspace(ctx, root); err != nil {
		checks = append(checks, doctorCheck{Name: "workspace_config", Status: doctorStatusBlocked, Requirement: doctorRequired, Path: configPath, Error: err.Error()})
		return checks
	}
	checks = append(checks, doctorCheck{Name: "workspace_config", Status: doctorStatusPass, Requirement: doctorRequired, Path: configPath})
	checks = append(checks, writePolicyDoctorCheck(ctx, opts))
	return checks
}

func writePolicyDoctorCheck(ctx context.Context, opts options) doctorCheck {
	resolved, err := optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return doctorCheck{Name: "write_policy", Status: doctorStatusBlocked, Requirement: doctorRequired, Error: err.Error()}
	}
	switch strings.TrimSpace(resolved.writePolicy) {
	case "", cliWritePolicyObserve, cliWritePolicyDryRun, cliWritePolicyPreview, cliWritePolicyApprovedWrite:
		return doctorCheck{Name: "write_policy", Status: doctorStatusPass, Requirement: doctorRequired, Guidance: "writes remain previewed or approval-gated"}
	case cliWritePolicyTrustedLocal:
		return doctorCheck{Name: "write_policy", Status: doctorStatusBlocked, Requirement: doctorRequired, Error: "trusted-local allows direct writes", Guidance: "use preview, dry-run, observe, or approved-write for safer operator checks"}
	default:
		return doctorCheck{Name: "write_policy", Status: doctorStatusBlocked, Requirement: doctorRequired, Error: "invalid write policy"}
	}
}
