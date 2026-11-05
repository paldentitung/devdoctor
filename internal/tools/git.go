package tools

import "devdoctor/internal/checks"

var minGit = checks.Version{Major: 2, Minor: 45}

// Git checks that Git is installed and is a supported version.
type Git struct{}

func (Git) Run() checks.Result {
	path, ok := checks.FindInPath("git")
	if !ok {
		return checks.Failed(
			"Git",
			"Git is not installed or not available in PATH.",
			"Install Git from https://git-scm.com and ensure it is available in PATH.",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"Git",
			"Git was found but its version could not be retrieved.",
			"Check your Git installation.",
		)
	}

	have, ok := checks.ParseVersion(version)
	if ok && !have.AtLeast(minGit) {
		return checks.Warning(
			"Git",
			"Git "+have.String()+" is older than the required "+minGit.String()+".",
			"Upgrade Git to "+minGit.String()+" or newer: https://git-scm.com",
		)
	}

	return checks.Passed("Git", "Git is installed: "+version)
}