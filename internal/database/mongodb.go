package database

import (
	"fmt"
	"strings"

	"devdoctor/internal/checks"
)

const mongoPort = 27017

// MongoDB checks that a MongoDB tool is installed and that a server
// is accepting connections on the default port.
type MongoDB struct{}

func (MongoDB) Run() checks.Result {
	// mongosh is the modern client. mongod is the server itself.
	path, ok := checks.FindInPath("mongosh")
	if !ok {
		path, ok = checks.FindInPath("mongod")
	}
	if !ok {
		return checks.Failed(
			"MongoDB",
			"MongoDB (mongosh or mongod) is not installed or not available in PATH.",
			"Install MongoDB from https://www.mongodb.com/try/download/community",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"MongoDB",
			"MongoDB was found but its version could not be retrieved.",
			"Check your MongoDB installation.",
		)
	}
	// mongod prints several lines. Keep the first one.
	version = strings.TrimSpace(strings.SplitN(version, "\n", 2)[0])

	if !checks.CanConnect("localhost", mongoPort) {
		return checks.Warning(
			"MongoDB",
			fmt.Sprintf("MongoDB %s is installed, but no server is listening on port %d.", version, mongoPort),
			"Start MongoDB (macOS: brew services start mongodb-community, Linux: sudo systemctl start mongod, Windows: start the MongoDB service).",
		)
	}

	return checks.Passed("MongoDB", fmt.Sprintf("MongoDB %s, server reachable on port %d", version, mongoPort))
}