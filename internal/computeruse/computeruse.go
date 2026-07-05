package computeruse

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	PolicyDeny  = "deny"
	PolicyAsk   = "ask"
	PolicyAllow = "allow"
)

const defaultMaxBytes = 12000

type Request struct {
	App       string
	Command   []string
	Policy    string
	MaxBytes  int
	TimeoutMS int
}

type Result struct {
	Status        string
	App           string
	Permission    string
	Output        string
	Error         string
	Bytes         int
	Truncated     bool
	ExitCode      int
	ReceiptSHA256 string
}

func NormalizePolicy(policy string) string {
	switch strings.TrimSpace(policy) {
	case "":
		return PolicyAsk
	default:
		return strings.TrimSpace(policy)
	}
}

func ValidatePolicy(policy string) error {
	switch NormalizePolicy(policy) {
	case PolicyDeny, PolicyAsk, PolicyAllow:
		return nil
	default:
		return fmt.Errorf("computer_policy must be deny, ask, or allow")
	}
}

func Snapshot(ctx context.Context, req Request) Result {
	policy := NormalizePolicy(req.Policy)
	app := strings.TrimSpace(req.App)
	result := Result{App: app, Permission: policy}
	if err := ValidatePolicy(policy); err != nil {
		result.Status = "invalid"
		result.Error = err.Error()
		result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
		return result
	}
	if app == "" {
		result.Status = "invalid"
		result.Error = "app is required"
		result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
		return result
	}
	if len(req.Command) == 0 {
		result.Status = "skipped"
		result.Error = "computer command is required"
		result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
		return result
	}
	if policy != PolicyAllow {
		result.Status = "denied"
		if policy == PolicyDeny {
			result.Error = "computer policy denies desktop use"
		} else {
			result.Error = "computer policy requires explicit operator approval"
		}
		result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
		return result
	}
	argv := commandWithApp(req.Command, app)
	runCtx, cancel := context.WithTimeout(ctx, timeout(req.TimeoutMS))
	defer cancel()
	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out, outTruncated := boundedText(stdout.String(), maxBytes(req.MaxBytes))
	errText, errTruncated := boundedText(stderr.String(), maxBytes(req.MaxBytes))
	result.Output = strings.TrimSpace(out)
	result.Error = strings.TrimSpace(errText)
	result.Bytes = len(stdout.String()) + len(stderr.String())
	result.Truncated = outTruncated || errTruncated
	if err == nil {
		result.Status = "pass"
		result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
		return result
	}
	result.Status = "fail"
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
	} else {
		result.Error = strings.TrimSpace(result.Error + "\n" + err.Error())
	}
	result.ReceiptSHA256 = receiptDigest("computer_snapshot", app, policy, result.Status)
	return result
}

func commandWithApp(command []string, app string) []string {
	argv := append([]string(nil), command...)
	replaced := false
	for index, arg := range argv {
		if strings.Contains(arg, "{app}") {
			argv[index] = strings.ReplaceAll(arg, "{app}", app)
			replaced = true
		}
	}
	if !replaced {
		argv = append(argv, app)
	}
	return argv
}

func timeout(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		return 10 * time.Second
	}
	return time.Duration(timeoutMS) * time.Millisecond
}

func maxBytes(max int) int {
	if max <= 0 {
		return defaultMaxBytes
	}
	return max
}

func boundedText(text string, limit int) (string, bool) {
	if len(text) <= limit {
		return text, false
	}
	end := 0
	for index := range text {
		if index > limit {
			break
		}
		end = index
	}
	if end == 0 {
		return "", true
	}
	return text[:end], true
}

func receiptDigest(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
