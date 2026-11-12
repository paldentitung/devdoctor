package tools

import "github.com/paldentitung/devdoctor/internal/checks"

// Docker checks that Docker is installed and that its daemon is running.
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

	// "docker --version" only talks to the client. "docker info" needs the
	// daemon, so it fails if Docker isn't actually running.
	serverVersion, err := checks.GetVersion(path, "info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return checks.Warning(
			"Docker",
			"Docker is installed ("+version+") but the daemon is not responding.",
			"Start Docker Desktop, or on Linux run: sudo systemctl start docker",
		)
	}

	return checks.Passed("Docker", "Docker is running (server "+serverVersion+"): "+version)
}
