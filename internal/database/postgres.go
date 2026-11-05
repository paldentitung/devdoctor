package database

import "devdoctor/internal/checks"

// Postgres checks that the PostgreSQL client (psql) is installed.
type Postgres struct{}

func (Postgres) Run() checks.Result {
	path, ok := checks.FindInPath("psql")
	if !ok {
		return checks.Failed(
			"PostgreSQL",
			"psql is not installed or not available in PATH.",
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

	return checks.Passed("PostgreSQL", "PostgreSQL client is installed: "+version)
}