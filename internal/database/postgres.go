package database

import (
	"fmt"
	"strings"

	"devdoctor/internal/checks"
)

const postgresPort = 5432

// Postgres checks that the psql client is installed and that a server
// is accepting connections on the default port.
type Postgres struct{}

func (Postgres) Run() checks.Result {
	path, ok := checks.FindInPath("psql")
	if !ok {
		return checks.Failed(
			"PostgreSQL",
			"PostgreSQL client (psql) is not installed or not available in PATH.",
			"Install PostgreSQL from https://www.postgresql.org/download",
		)
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning(
			"PostgreSQL",
			"psql was found but its version could not be retrieved.",
			"Check your PostgreSQL installation.",
		)
	}
	version = strings.TrimSpace(version)

	if !checks.CanConnect("localhost", postgresPort) {
		return checks.Warning(
			"PostgreSQL",
			fmt.Sprintf("%s is installed, but no server is listening on port %d.", version, postgresPort),
			"Start PostgreSQL (macOS: brew services start postgresql, Linux: sudo systemctl start postgresql, Windows: start the PostgreSQL service).",
		)
	}

	return checks.Passed("PostgreSQL", fmt.Sprintf("%s, server reachable on port %d", version, postgresPort))
}