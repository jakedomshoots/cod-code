package ceo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"ceoharness/internal/workspace"
)

type PatchApproval struct {
	Status         string `json:"status"`
	PreviewDigest  string `json:"preview_digest"`
	ApprovedDigest string `json:"approved_digest,omitempty"`
	PreviewCount   int    `json:"preview_count"`
}

type patchApprovalDigestEntry struct {
	Path string `json:"path"`
	Diff string `json:"diff"`
}

func newPatchPreviewApproval(previews []workspace.ReplaceTextResult) *PatchApproval {
	if len(previews) == 0 {
		return nil
	}
	return &PatchApproval{
		Status:        "previewed",
		PreviewDigest: patchPreviewDigest(previews),
		PreviewCount:  len(previews),
	}
}

func newApprovedPatchApproval(previews []workspace.ReplaceTextResult, approvedDigest string) *PatchApproval {
	approval := newPatchPreviewApproval(previews)
	if approval == nil {
		return nil
	}
	approval.Status = "approved"
	approval.ApprovedDigest = approvedDigest
	return approval
}

func patchPreviewDigest(previews []workspace.ReplaceTextResult) string {
	entries := make([]patchApprovalDigestEntry, 0, len(previews))
	for _, preview := range previews {
		entries = append(entries, patchApprovalDigestEntry{
			Path: preview.Path,
			Diff: preview.Diff,
		})
	}
	payload, err := json.Marshal(entries)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
