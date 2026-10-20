//go:build linux

package resume

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

const sandboxPlatform = "Linux without Landlock"

// System folders the compiler may read: its libraries, fonts, and config.
var landlockReadOnlyRoots = []string{"/usr", "/lib", "/lib64", "/bin", "/etc"}

func landlockABI() int {
	version, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, 0, 0, unix.LANDLOCK_CREATE_RULESET_VERSION)
	if errno != 0 {
		return 0
	}
	return int(version)
}

func sandboxAvailable() bool {
	return landlockABI() >= 1
}

// runSandboxed starts the compiler from one OS thread restricted by Landlock.
// The child inherits the restriction; the thread is never unlocked, so Go
// discards it when the goroutine exits and no other code runs restricted.
func runSandboxed(command *exec.Cmd, policy sandboxPolicy) error {
	done := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		if err := restrictCurrentThread(policy); err != nil {
			done <- fmt.Errorf("sandbox the LaTeX compiler: %w", err)
			return
		}
		done <- command.Run()
	}()
	return <-done
}

func restrictCurrentThread(policy sandboxPolicy) error {
	abi := landlockABI()
	if abi < 1 {
		return errors.New("Landlock is unavailable")
	}
	// ABI 1 handles EXECUTE through MAKE_SYM; later versions add more rights.
	handled := uint64(unix.LANDLOCK_ACCESS_FS_MAKE_SYM<<1 - 1)
	if abi >= 2 {
		handled |= unix.LANDLOCK_ACCESS_FS_REFER
	}
	if abi >= 3 {
		handled |= unix.LANDLOCK_ACCESS_FS_TRUNCATE
	}
	if abi >= 5 {
		handled |= unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
	}
	attr := unix.LandlockRulesetAttr{Access_fs: handled}
	if abi >= 4 {
		// Handled with no rule means denied: the compiler runs offline.
		attr.Access_net = unix.LANDLOCK_ACCESS_NET_BIND_TCP | unix.LANDLOCK_ACCESS_NET_CONNECT_TCP
	}
	ruleset, _, errno := unix.Syscall(unix.SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), unsafe.Sizeof(attr), 0)
	if errno != 0 {
		return fmt.Errorf("create Landlock ruleset: %w", errno)
	}
	defer unix.Close(int(ruleset))

	readFile := uint64(unix.LANDLOCK_ACCESS_FS_READ_FILE)
	readOnly := readFile | unix.LANDLOCK_ACCESS_FS_READ_DIR | unix.LANDLOCK_ACCESS_FS_EXECUTE
	writeFile := uint64(unix.LANDLOCK_ACCESS_FS_WRITE_FILE)
	if abi >= 3 {
		writeFile |= unix.LANDLOCK_ACCESS_FS_TRUNCATE
	}
	rules := []struct {
		path     string
		access   uint64
		optional bool
	}{
		{policy.workspace, handled, false},
		{policy.executable, readFile | unix.LANDLOCK_ACCESS_FS_EXECUTE, false},
		{policy.bundle, readFile, false},
		{"/dev/null", readFile | writeFile, true},
		{"/dev/urandom", readFile, true},
	}
	for _, root := range landlockReadOnlyRoots {
		rules = append(rules, struct {
			path     string
			access   uint64
			optional bool
		}{root, readOnly, true})
	}
	for _, rule := range rules {
		if err := addLandlockRule(int(ruleset), rule.path, rule.access&handled); err != nil {
			if rule.optional && errors.Is(err, unix.ENOENT) {
				continue
			}
			return fmt.Errorf("allow %s: %w", rule.path, err)
		}
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("set no_new_privs: %w", err)
	}
	if _, _, errno := unix.Syscall(unix.SYS_LANDLOCK_RESTRICT_SELF, ruleset, 0, 0); errno != 0 {
		return fmt.Errorf("apply Landlock ruleset: %w", errno)
	}
	return nil
}

func addLandlockRule(ruleset int, path string, access uint64) error {
	fd, err := unix.Open(path, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		// Landlock rejects directory-only rights on a file.
		access &= unix.LANDLOCK_ACCESS_FS_EXECUTE | unix.LANDLOCK_ACCESS_FS_READ_FILE | unix.LANDLOCK_ACCESS_FS_WRITE_FILE | unix.LANDLOCK_ACCESS_FS_TRUNCATE | unix.LANDLOCK_ACCESS_FS_IOCTL_DEV
	}
	attr := unix.LandlockPathBeneathAttr{Allowed_access: access, Parent_fd: int32(fd)}
	if _, _, errno := unix.Syscall6(unix.SYS_LANDLOCK_ADD_RULE, uintptr(ruleset), unix.LANDLOCK_RULE_PATH_BENEATH, uintptr(unsafe.Pointer(&attr)), 0, 0, 0); errno != 0 {
		return errno
	}
	return nil
}
