package checks

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// FindInPath reports whether the command is available in PATH and returns its full path.
func FindInPath(command string) (string, bool) {
	path, err := exec.LookPath(command)
	return path, err == nil
}

// GetVersion runs "<command> --version" (or custom args) and returns the trimmed output.
// It gives up after 5 seconds.
func GetVersion(path string, args ...string) (string, error) {
	if len(args) == 0 {
		args = []string{"--version"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.WaitDelay = time.Second

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("%s timed out after 5s", filepath.Base(path))
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// FileExists reports whether a file or folder exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// PortIsFree reports whether a local TCP port is free.
func PortIsFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
