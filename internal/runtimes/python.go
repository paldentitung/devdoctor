package runtimes

import "devdoctor/internal/checks"

var minPython = checks.Version{Major: 3, Minor: 11}

// Python checks that Python is installed and is a supported version.
type Python struct{}

func (Python) Run() checks.Result {
	path, ok := checks.FindInPath("python3")
	if !ok {
		path, ok = checks.FindInPath("python")
	}
	if !ok {
		return checks.Failed(
			"Python",
			"Python is not installed or not available in PATH.",
			"Install Python from https://www.python.org/downloads",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"Python",
			"Python was found but its version could not be retrieved.",
			"Check your Python installation.",
		)
	}

	have, ok := checks.ParseVersion(version)
	if ok && !have.AtLeast(minPython) {
		return checks.Warning(
			"Python",
			"Python "+have.String()+" is older than the required "+minPython.String()+".",
			"Upgrade Python to "+minPython.String()+" or newer: https://www.python.org/downloads",
		)
	}

	return checks.Passed("Python", "Python is installed: "+version)
}