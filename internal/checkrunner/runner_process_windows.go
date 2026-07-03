//go:build windows

package checkrunner

import (
	"os/exec"

	"ceoharness/internal/processcancel"
)

func configureCommandCancellation(cmd *exec.Cmd) {
	processcancel.ConfigureProcessTreeCancellation(cmd)
}
