package model

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HTTPErrorKind string

const (
	HTTPErrorKindRequestFailed HTTPErrorKind = "request_failed"
	HTTPErrorKindUnauthorized  HTTPErrorKind = "unauthorized"
	HTTPErrorKindRateLimited   HTTPErrorKind = "rate_limited"
	HTTPErrorKindUnavailable   HTTPErrorKind = "unavailable"
)

type HTTPStatusError struct {
	StatusCode   int
	Kind         HTTPErrorKind
	RetryAfterMS int64
	Body         string
}

func (e *HTTPStatusError) Error() string {
	body := strings.TrimSpace(e.Body)
	if body != "" {
		return fmt.Sprintf("status %d: %s: %s", e.StatusCode, e.Unwrap(), body)
	}
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Unwrap())
}

func (e *HTTPStatusError) Unwrap() error {
	return sentinelForHTTPErrorKind(e.Kind)
}

func newHTTPStatusError(status int, body string, retryAfter string) *HTTPStatusError {
	return &HTTPStatusError{
		StatusCode:   status,
		Kind:         classifyHTTPErrorKind(status),
		RetryAfterMS: parseRetryAfterMS(retryAfter, time.Now()),
		Body:         strings.TrimSpace(body),
	}
}

func parseRetryAfterMS(value string, now time.Time) int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0
	}
	seconds, err := strconv.Atoi(trimmed)
	if err == nil {
		if seconds <= 0 {
			return 0
		}
		return int64(seconds) * int64(time.Second/time.Millisecond)
	}
	retryAt, err := http.ParseTime(trimmed)
	if err != nil {
		return 0
	}
	wait := retryAt.Sub(now)
	if wait <= 0 {
		return 0
	}
	return wait.Milliseconds()
}

func classifyHTTPStatus(status int) error {
	return sentinelForHTTPErrorKind(classifyHTTPErrorKind(status))
}

func classifyHTTPErrorKind(status int) HTTPErrorKind {
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return HTTPErrorKindUnauthorized
	case status == http.StatusTooManyRequests:
		return HTTPErrorKindRateLimited
	case status >= http.StatusInternalServerError:
		return HTTPErrorKindUnavailable
	default:
		return HTTPErrorKindRequestFailed
	}
}

func sentinelForHTTPErrorKind(kind HTTPErrorKind) error {
	switch kind {
	case HTTPErrorKindUnauthorized:
		return ErrHTTPUnauthorized
	case HTTPErrorKindRateLimited:
		return ErrHTTPRateLimited
	case HTTPErrorKindUnavailable:
		return ErrHTTPUnavailable
	default:
		return ErrHTTPRequestFailed
	}
}
