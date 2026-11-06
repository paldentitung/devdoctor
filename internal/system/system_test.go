package system

import "testing"

func TestGetInfo(t *testing.T) {
	info := GetInfo()

	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Architecture == "" {
		t.Error("Architecture should not be empty")
	}

	if info.CPUCores <= 0 {
		t.Error("CPU cores should be greater than zero")
	}

	if info.GoVersion == "" {
		t.Error("Go version should not be empty")
	}

	if info.HomeDirectory == "" {
		t.Error("Home directory should not be empty")
	}

	if info.Username == "" {
		t.Error("Username should not be empty")
	}
}
