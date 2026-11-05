package runtimes

import "devdoctor/internal/checks"

// Golang checks that Go is installed and reports its version.
type Golang struct{}

func (Golang) Run() checks.Result {
	path, ok := checks.FindInPath("go")
	if !ok {
		return checks.Failed(
			"Go",
			"Go is not installed or not available in PATH.",
			"Install Go from https://go.dev/dl and ensure it is available in PATH.",
		)
	}

	// Go uses "go version", not "go --version".
	version, err := checks.GetVersion(path, "version")
	if err != nil {
		return checks.Warning(
			"Go",
			"Go was found but its version could not be retrieved.",
			"Check your Go installation.",
		)
	}

	return checks.Passed("Go", "Go is installed: "+version)
}