package packagemanager

import "devdoctor/internal/checks"

type Npm struct{}

func (Npm) Run() checks.Result {
    path, ok := checks.FindInPath("npm")
    if !ok {
        return checks.Failed("npm", "npm is not installed or not available in PATH.", "Install Node.js, which includes npm, and ensure it is available in PATH.")
    }

    version, err := checks.GetVersion(path)
    if err != nil {
        return checks.Warning("npm", "npm was found but its version could not be retrieved.", "Check your npm installation.")
    }

    return checks.Passed("npm", "npm is installed: "+version)
}