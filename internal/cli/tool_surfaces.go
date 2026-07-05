package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"ceoharness/internal/browseruse"
	"ceoharness/internal/computeruse"
	"ceoharness/internal/toolmanifest"
)

type surfaceDoctorReport struct {
	Surface        string   `json:"surface"`
	Status         string   `json:"status"`
	Summary        string   `json:"summary"`
	Policy         string   `json:"policy"`
	CommandArgv    []string `json:"command_argv,omitempty"`
	CommandPresent bool     `json:"command_present"`
}

func runToolManifest(out io.Writer, opts options) error {
	return writeJSONOrText(out, opts, toolmanifest.Default(), "tools manifest: browser.read, computer.snapshot, tools.manifest\n")
}

func runBrowser(ctx context.Context, out io.Writer, opts options) error {
	resolved, err := optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return err
	}
	opts = resolved
	switch strings.TrimSpace(opts.browserCommand) {
	case "doctor":
		policy := browseruse.NormalizePolicy(opts.browserPolicy)
		report := surfaceDoctorReport{Surface: "browser", Status: "pass", Summary: "browser read is available; default policy allows localhost only", Policy: policy, CommandArgv: append([]string(nil), opts.browserBackendCommand...), CommandPresent: len(opts.browserBackendCommand) > 0}
		if err := browseruse.ValidatePolicy(policy); err != nil {
			report.Status = "fail"
			report.Summary = err.Error()
		}
		return writeJSONOrText(out, opts, report, fmt.Sprintf("browser doctor: %s policy=%s command_present=%t\n", report.Status, report.Policy, report.CommandPresent))
	case "manifest":
		return runToolManifest(out, opts)
	case "read":
		result := browseruse.Read(ctx, browseruse.Request{URL: opts.browserURL, Policy: opts.browserPolicy, TimeoutMS: opts.toolCommandTimeoutMS})
		return writeJSONOrText(out, opts, result, browserReadText(result))
	default:
		return fmt.Errorf("unknown browser command %q", opts.browserCommand)
	}
}

func runComputer(ctx context.Context, out io.Writer, opts options) error {
	resolved, err := optionsWithWorkspaceDefaults(ctx, opts)
	if err != nil {
		return err
	}
	opts = resolved
	switch strings.TrimSpace(opts.computerCommand) {
	case "doctor":
		policy := computeruse.NormalizePolicy(opts.computerPolicy)
		report := surfaceDoctorReport{Surface: "computer", Status: "pass", Summary: "computer snapshot is available when an accessibility command is configured and policy=allow", Policy: policy, CommandArgv: append([]string(nil), opts.computerBackendCommand...), CommandPresent: len(opts.computerBackendCommand) > 0}
		if err := computeruse.ValidatePolicy(policy); err != nil {
			report.Status = "fail"
			report.Summary = err.Error()
		} else if len(opts.computerBackendCommand) == 0 {
			report.Status = "blocked"
			report.Summary = "computer command is required for desktop snapshots"
		}
		return writeJSONOrText(out, opts, report, fmt.Sprintf("computer doctor: %s policy=%s command_present=%t\n", report.Status, report.Policy, report.CommandPresent))
	case "manifest":
		return runToolManifest(out, opts)
	case "snapshot":
		result := computeruse.Snapshot(ctx, computeruse.Request{App: opts.computerApp, Command: opts.computerBackendCommand, Policy: opts.computerPolicy, TimeoutMS: opts.toolCommandTimeoutMS})
		return writeJSONOrText(out, opts, result, computerSnapshotText(result))
	default:
		return fmt.Errorf("unknown computer command %q", opts.computerCommand)
	}
}

func writeJSONOrText(out io.Writer, opts options, value any, text string) error {
	if opts.reportFormat == "" || opts.reportFormat == reportFormatJSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	}
	_, err := io.WriteString(out, text)
	return err
}

func browserReadText(result browseruse.Result) string {
	if result.Status != "pass" {
		return fmt.Sprintf("browser read: %s %s\n", result.Status, result.Error)
	}
	return fmt.Sprintf("browser read: pass status=%d bytes=%d truncated=%t receipt=%s\n%s\n", result.HTTPStatus, result.Bytes, result.Truncated, result.ReceiptSHA256, result.Output)
}

func computerSnapshotText(result computeruse.Result) string {
	if result.Status != "pass" {
		return fmt.Sprintf("computer snapshot: %s %s\n", result.Status, result.Error)
	}
	return fmt.Sprintf("computer snapshot: pass app=%s bytes=%d truncated=%t receipt=%s\n%s\n", result.App, result.Bytes, result.Truncated, result.ReceiptSHA256, result.Output)
}
