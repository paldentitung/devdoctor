package checks

import (
    "fmt"
    "net"
    "os"
    "os/exec"
    "strings"
)

// Is the command available in PATH? Returns its full path.
func FindInPath(command string) (string, bool) {
    path, err := exec.LookPath(command)
    return path, err == nil
}

// Runs "<command> --version" (or custom args) and returns the trimmed output.
func GetVersion(path string, args ...string) (string, error) {
    if len(args) == 0 {
        args = []string{"--version"}
    }
    out, err := exec.Command(path, args...).Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(out)), nil
}

// Does a file or folder exist?
func FileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

// Is a local TCP port free?
func PortIsFree(port int) bool {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    ln.Close()
    return true
}