package system

import "testing"

func TestGetSystemInfo(t *testing.T) {
    system := GetSystemInfo()

    if system.OS == "" {
        t.Error("OS should not be empty")
    }

    if system.Architecture == "" {
        t.Error("Architecture should not be empty")
    }

    if system.CPUCores <= 0 {
        t.Error("CPU cores should be greater than zero")
    }

    if system.GoVersion == "" {
        t.Error("Go version should not be empty")
    }

	if system.HomeDirectory == "" {
    t.Error("Home directory should not be empty")
	}

	if system.Username == "" {
		t.Error("Username should not be empty")
	}

}