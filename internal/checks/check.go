package checks

import "github.com/paldentitung/devdoctor/internal/detect"

type Check interface {
	Run() Result
}

type Group struct {
	Name   string
	Checks []Check
}

// Scoped wraps a check so it only runs for the given project types.
// Embedding Check means Scoped satisfies Check without extra code.
type Scoped struct {
	Check
	Types []detect.Type
}

// For limits c to the given project types. Unwrapped checks always run.
func For(c Check, types ...detect.Type) Check {
	return Scoped{Check: c, Types: types}
}
