//go:build darwin

package resume

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	sandboxPlatform = "macOS without sandbox-exec"
	sandboxExec     = "/usr/bin/sandbox-exec"
)

func sandboxAvailable() bool {
	_, err := os.Stat(sandboxExec)
	return err == nil
}

// runSandboxed runs the compiler under sandbox-exec. The profile denies user,
// temp, and network access, then re-allows the workspace, the executable, and
// the bundle. The allow must name file-read-data too: a rule for that specific
// operation beats a file-read* wildcard regardless of order.
func runSandboxed(command *exec.Cmd, policy sandboxPolicy) error {
	command.Args = append([]string{sandboxExec, "-p", darwinSandboxProfile(policy), command.Path}, command.Args[1:]...)
	command.Path = sandboxExec
	return command.Run()
}

func darwinSandboxProfile(policy sandboxPolicy) string {
	quote := func(path string) string {
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(path) + `"`
	}
	workspace := quote(policy.workspace)
	return `(version 1)
(allow default)
(deny network*)
(deny file-write*)
(allow file-write* (subpath ` + workspace + `) (literal "/dev/null") (literal "/dev/dtracehelper"))
(deny file-read-data (subpath "/Users") (subpath "/Volumes") (subpath "/private/var/folders") (subpath "/private/tmp") (subpath "/private/var/root") (subpath "/Library/Keychains") (subpath "/Library/Application Support"))
(allow file-read* file-read-data (subpath ` + workspace + `) (literal ` + quote(policy.executable) + `) (literal ` + quote(policy.bundle) + `))
`
}
