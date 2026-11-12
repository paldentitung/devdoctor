package database

import (
	"fmt"
	"strings"

	"devdoctor/internal/checks"
)

const mongoPort = 27017

type MongoDB struct{}

func (MongoDB) Run() checks.Result {
	shellPath, hasShell := checks.FindInPath("mongosh")
	serverPath, hasServer := checks.FindInPath("mongod")

	if !hasShell && !hasServer {
		return checks.Failed(
			"MongoDB",
			"Neither mongosh nor mongod was found in PATH.",
			"Install MongoDB from https://www.mongodb.com/try/download/community",
		)
	}

	var parts []string
	if hasServer {
		parts = append(parts, "mongod "+firstLine(serverPath))
	}
	if hasShell {
		parts = append(parts, "mongosh "+firstLine(shellPath))
	}
	installed := strings.Join(parts, ", ")

	if !checks.CanConnect("localhost", mongoPort) {
		return checks.Warning(
			"MongoDB",
			fmt.Sprintf("%s installed, but nothing is listening on port %d.", installed, mongoPort),
			"Start MongoDB (Linux: sudo systemctl start mongod, macOS: brew services start mongodb-community, Windows: start the MongoDB service), or ignore this if you use a remote or Docker instance.",
		)
	}

	return checks.Passed("MongoDB", fmt.Sprintf("%s, server reachable on port %d", installed, mongoPort))
}

// firstLine runs `<path> --version` and returns the first non-empty line,
// or "(version unknown)" if it fails.
func firstLine(path string) string {
	out, err := checks.GetVersion(path)
	if err != nil {
		return "(version unknown)"
	}
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return "(version unknown)"
}
