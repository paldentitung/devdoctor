package tools

import (
	"strings"

	"devdoctor/internal/checks"
)

// Curl checks that curl is installed and reports its version.
type Curl struct{}

func (Curl) Run() checks.Result {
	path, ok := checks.FindInPath("curl")
	if !ok {
		return checks.Failed(
			"curl",
			"curl is not installed or not available in PATH.",
			"Install curl using your package manager (for example: sudo apt install curl).",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"curl",
			"curl was found but its version could not be retrieved.",
			"Check your curl installation.",
		)
	}

	// curl prints several lines; keep only the first.
	version = strings.SplitN(version, "\n", 2)[0]

	return checks.Passed("curl", "curl is installed: "+version)
}