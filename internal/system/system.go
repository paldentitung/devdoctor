package system

import (
	"os"
	"os/user"
	"runtime"
)

type Info struct {
	OS            string
	Architecture  string
	CPUCores      int
	GoVersion     string
	HomeDirectory string
	Username      string
	Shell         string
}

func GetInfo() Info {
	currentUser, _ := user.Current()
	homeDir, _ := os.UserHomeDir()

	username := ""
	if currentUser != nil {
		username = currentUser.Username
	}

	return Info{
		OS:            runtime.GOOS,
		Architecture:  runtime.GOARCH,
		CPUCores:      runtime.NumCPU(),
		GoVersion:     runtime.Version(),
		HomeDirectory: homeDir,
		Username:      username,
		Shell:         os.Getenv("SHELL"),
	}
}
