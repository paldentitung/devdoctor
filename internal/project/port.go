package project

import (
	"fmt"

	"devdoctor/internal/checks"
)

// Port checks that a local TCP port is free.
type Port struct{ Number int }

func (p Port) Run() checks.Result {
	name := fmt.Sprintf("Port %d", p.Number)

	if !checks.PortIsFree(p.Number) {
		return checks.Warning(
			name,
			"Port is already in use.",
			"Stop the process using it, or choose another port.",
		)
	}

	return checks.Passed(name, "Port is free.")
}