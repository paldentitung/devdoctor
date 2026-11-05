package tools

import "devdoctor/internal/checks"

// Docker checks that Docker is installed and reports its version.
type Docker struct{}

func (Docker) Run() checks.Result {
	path, ok := checks.FindInPath("docker")
	if !ok {
		return checks.Failed(
			"Docker",
			"Docker is not installed or not available in PATH.",
			"Install Docker from https://docs.docker.com/get-docker",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"Docker",
			"Docker was found but its version could not be retrieved.",
			"Check your Docker installation.",
		)
	}

	return checks.Passed("Docker", "Docker is installed: "+version)
}