//go:build windows

package adapter

import (
	"os/exec"

	"ceoharness/internal/processcancel"
)

func isolateProbeProcess(cmd *exec.Cmd) {
	processcancel.ConfigureProcessTreeCancellation(cmd)
}
