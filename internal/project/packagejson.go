package project

import "devdoctor/internal/checks"

// PackageJSON checks that package.json exists in the current folder.
type PackageJSON struct{}

func (PackageJSON) Run() checks.Result {
	if !checks.FileExists("package.json") {
		return checks.Warning(
			"package.json",
			"package.json was not found in the current folder.",
			"Run devdoctor from your project root, or run npm init.",
		)
	}

	return checks.Passed("package.json", "package.json found.")
}