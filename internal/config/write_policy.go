package config

import (
	"fmt"
	"strings"
)

const (
	WritePolicyObserve       = "observe"
	WritePolicyDryRun        = "dry-run"
	WritePolicyPreview       = "preview"
	WritePolicyApprovedWrite = "approved-write"
	WritePolicyTrustedLocal  = "trusted-local"
)

func validateWritePolicy(cfg Config) error {
	switch strings.TrimSpace(cfg.WritePolicy) {
	case "", WritePolicyObserve, WritePolicyDryRun, WritePolicyPreview, WritePolicyApprovedWrite, WritePolicyTrustedLocal:
		return nil
	default:
		return fmt.Errorf("write_policy: %w", ErrInvalidConfig)
	}
}
