package ceo

import (
	"strings"

	"ceoharness/internal/checkrunner"
)

const (
	verificationStatusUnverified = "unverified"
	verificationStatusPending    = "pending"
	verificationStatusPass       = "pass"
	verificationStatusFail       = "fail"
)

type VerificationContract struct {
	Status             string   `json:"status"`
	RequiredCheckCount int      `json:"required_check_count"`
	RequiredChecks     []string `json:"required_checks,omitempty"`
	CheckAttemptCount  int      `json:"check_attempt_count"`
	PassedCheckCount   int      `json:"passed_check_count"`
	FailedCheckCount   int      `json:"failed_check_count"`
}

func NewPendingVerificationContract(commands [][]string) VerificationContract {
	return VerificationContract{
		Status:             verificationStatus(len(commands), 0, 0, 0),
		RequiredCheckCount: len(commands),
		RequiredChecks:     renderVerificationCommands(commands),
	}
}

func NewVerificationContract(commands [][]string, checks []checkrunner.Result) VerificationContract {
	passed, failed := finalCheckCounts(len(commands), checks)
	return VerificationContract{
		Status:             verificationStatus(len(commands), len(checks), passed, failed),
		RequiredCheckCount: len(commands),
		RequiredChecks:     renderVerificationCommands(commands),
		CheckAttemptCount:  len(checks),
		PassedCheckCount:   passed,
		FailedCheckCount:   failed,
	}
}

func finalCheckCounts(required int, checks []checkrunner.Result) (int, int) {
	finalByIndex := map[int]string{}
	for index, check := range checks {
		checkIndex := check.CheckIndex
		if checkIndex <= 0 {
			checkIndex = index + 1
		}
		if checkIndex > required {
			continue
		}
		finalByIndex[checkIndex] = check.Status
	}
	passed := 0
	failed := 0
	for checkIndex := 1; checkIndex <= required; checkIndex++ {
		switch finalByIndex[checkIndex] {
		case verificationStatusPass:
			passed++
		case verificationStatusFail:
			failed++
		}
	}
	return passed, failed
}

func verificationStatus(required int, attempts int, passed int, failed int) string {
	if required == 0 {
		return verificationStatusUnverified
	}
	if attempts == 0 {
		return verificationStatusPending
	}
	if failed > 0 {
		return verificationStatusFail
	}
	if passed == required {
		return verificationStatusPass
	}
	return verificationStatusPending
}

func renderVerificationCommands(commands [][]string) []string {
	if len(commands) == 0 {
		return nil
	}
	rendered := make([]string, 0, len(commands))
	for _, command := range commands {
		rendered = append(rendered, strings.Join(command, " "))
	}
	return rendered
}
