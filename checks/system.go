package checks

import (
    "os"
    "os/user"
    "runtime"
)

type SystemInfo struct {
    OS            string
    Architecture  string
    CPUCores      int
    GoVersion     string
    HomeDirectory string
    Username      string
    Shell         string
}

func GetSystemInfo() SystemInfo {
    currentUser, _ := user.Current()
    homeDir, _ := os.UserHomeDir()

    username := ""
    if currentUser != nil {
        username = currentUser.Username
    }

    return SystemInfo{
        OS:            runtime.GOOS,
        Architecture:  runtime.GOARCH,
        CPUCores:      runtime.NumCPU(),
        GoVersion:     runtime.Version(),
        HomeDirectory: homeDir,
        Username:      username,
        Shell:         os.Getenv("SHELL"),
    }
}