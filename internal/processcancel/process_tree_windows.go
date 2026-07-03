//go:build windows

package processcancel

import (
	"os/exec"
	"time"
)

func ConfigureProcessTreeCancellation(cmd *exec.Cmd) {
	cmd.WaitDelay = 100 * time.Millisecond
}
