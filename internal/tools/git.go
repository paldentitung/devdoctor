package tools

import "devdoctor/internal/checks"

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

    return checks.Passed("Git", "Git is installed: "+version)
}