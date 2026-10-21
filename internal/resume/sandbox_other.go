//go:build !linux && !darwin

package resume

import (
	"errors"
	"os/exec"
)

const sandboxPlatform = "this platform"

// No OS sandbox yet on this platform; Compile falls back to the source check.
func sandboxAvailable() bool { return false }

func runSandboxed(*exec.Cmd, sandboxPolicy) error {
	return errors.New("no compiler sandbox on this platform")
}
