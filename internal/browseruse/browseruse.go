package browseruse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	PolicyDeny           = "deny"
	PolicyAsk            = "ask"
	PolicyAllowLocalhost = "allow-localhost"
	PolicyAllow          = "allow"
)

const defaultMaxBytes = 12000

type Request struct {
	URL       string
	Policy    string
	MaxBytes  int
	TimeoutMS int
}

type Result struct {
	Status        string
	URL           string
	Permission    string
	HTTPStatus    int
	ContentType   string
	Output        string
	Error         string
	Bytes         int
	Truncated     bool
	ReceiptSHA256 string
}

func NormalizePolicy(policy string) string {
	switch strings.TrimSpace(policy) {
	case "":
		return PolicyAllowLocalhost
	default:
		return strings.TrimSpace(policy)
	}
}

func ValidatePolicy(policy string) error {
	switch NormalizePolicy(policy) {
	case PolicyDeny, PolicyAsk, PolicyAllowLocalhost, PolicyAllow:
		return nil
	default:
		return fmt.Errorf("browser_policy must be deny, ask, allow-localhost, or allow")
	}
}

func Read(ctx context.Context, req Request) Result {
	policy := NormalizePolicy(req.Policy)
	cleanURL := strings.TrimSpace(req.URL)
	result := Result{URL: cleanURL, Permission: policy}
	if err := ValidatePolicy(policy); err != nil {
		result.Status = "invalid"
		result.Error = err.Error()
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	if cleanURL == "" {
		result.Status = "invalid"
		result.Error = "url is required"
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	parsed, err := url.Parse(cleanURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		result.Status = "invalid"
		result.Error = "url must be an absolute http or https URL"
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		result.Status = "invalid"
		result.Error = "url scheme must be http or https"
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	if err := allowedByPolicy(policy, parsed); err != nil {
		result.Status = "denied"
		result.Error = err.Error()
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	client := http.Client{Timeout: timeout(req.TimeoutMS)}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cleanURL, nil)
	if err != nil {
		result.Status = "invalid"
		result.Error = err.Error()
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	httpReq.Header.Set("User-Agent", "ceo-packet/browser-read")
	resp, err := client.Do(httpReq)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	defer resp.Body.Close()
	limit := maxBytes(req.MaxBytes)
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, int64(limit+1)))
	if readErr != nil {
		result.Status = "error"
		result.Error = readErr.Error()
		result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, "")
		return result
	}
	truncated := len(body) > limit
	if truncated {
		body = body[:limit]
	}
	result.Status = "pass"
	result.HTTPStatus = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.Output = string(body)
	result.Bytes = len(body)
	result.Truncated = truncated
	result.ReceiptSHA256 = receiptDigest("browser_read", cleanURL, policy, result.Status, fmt.Sprintf("%d", resp.StatusCode))
	return result
}

func allowedByPolicy(policy string, parsed *url.URL) error {
	switch policy {
	case PolicyDeny:
		return fmt.Errorf("browser policy denies browser use")
	case PolicyAsk:
		return fmt.Errorf("browser policy requires explicit operator approval")
	case PolicyAllowLocalhost:
		if !isLocalhost(parsed.Hostname()) {
			return fmt.Errorf("browser policy allows localhost URLs only")
		}
	}
	return nil
}

func isLocalhost(host string) bool {
	clean := strings.ToLower(strings.TrimSpace(host))
	if clean == "localhost" {
		return true
	}
	ip := net.ParseIP(clean)
	return ip != nil && ip.IsLoopback()
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

func receiptDigest(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
