// Package analyze looks INSIDE a project's files to find out what the
// project really is.
//
// The detect package only checks which files exist: "package.json is here,
// so this is a Node project". This package goes one step further and reads
// them: "package.json lists next and react, so it is a Next.js app, and it
// asks for Node 18.17 or newer".
//
// How the two fit together:
//
//	detect.Detect()    finds which kinds of project are here (node, go, ...)
//	analyze.Analyze()  reads the files of each kind and describes the project
//
// This package only produces data. It never prints anything and it never
// returns an error: if a file is missing or broken, that is recorded as a
// warning on the result and the rest of the analysis carries on.
//
// # Adding a new language
//
// Say you want to support Rust:
//
//  1. Create rust.go with a type that has the two methods of Analyzer.
//  2. Add it to the analyzers list below.
//
// Nothing else in the program needs to change.
package analyze

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devdoctor/internal/detect"
)

// A Finding is one thing we learned about the project, plus the proof.
//
// Example: Name "next" with Evidence "package.json (dependencies): next 14.1.0"
// means "this is a Next.js project, because package.json lists next".
// Showing the proof lets the user check that devdoctor got it right.
type Finding struct {
	Name     string `json:"name"`
	Evidence string `json:"evidence"`
}

// A Constraint is a version of a tool that the project asks for.
//
// Example: a package.json containing  "engines": {"node": ">=18.17"}  becomes
//
//	Tool: "node", Raw: ">=18.17", Source: "package.json → engines.node"
//
// Raw is kept exactly as written in the file, because version rules come in
// many styles (">=18", "^18 || >=20", ...). Pass it to
// checks.ParseMinConstraint to get the lowest version it allows.

type Constraint struct {
	Tool   string `json:"tool"`   
	Raw    string `json:"raw"`    
	Source string `json:"source"` 
}

// An Analysis is everything we learned about ONE kind of project.
//
// A folder can produce several: a Node frontend next to a Go backend gives
// two Analysis values. Anything we could not find is simply left empty.
type Analysis struct {
	Type           detect.Type `json:"type"`                      // which kind: node, go or python
	Frameworks     []Finding   `json:"frameworks,omitempty"`      // e.g. next, react, gin, django
	PackageManager string      `json:"package_manager,omitempty"` // e.g. pnpm, npm, poetry, pip
	Requires       *Constraint `json:"requires,omitempty"`        // the runtime version the project asks for
	Scripts        []string    `json:"scripts,omitempty"`         // commands defined in package.json, e.g. dev, build
	Warnings       []string    `json:"warnings,omitempty"`        // problems found while reading the files
}


func (a Analysis) HasFramework(name string) bool {
	for _, f := range a.Frameworks {
		if f.Name == name {
			return true
		}
	}
	return false
}

// FrameworkNames returns just the names, such as ["next", "react"], without
// the evidence, in the order they were found.
func (a Analysis) FrameworkNames() []string {
	names := make([]string, len(a.Frameworks))
	for i, f := range a.Frameworks {
		names[i] = f.Name
	}
	return names
}

// An Analyzer reads the files of ONE language or toolchain (package.json for
// Node, go.mod for Go, ...) and describes the project.
//
// Every language has its own Analyzer, but they all return the same Analysis
// shape. That is why the rest of the program never needs to know which
// language it is dealing with.
type Analyzer interface {
	// Type says which kind of project this Analyzer handles.
	Type() detect.Type

	// Analyze reads the project files found in dir. It must not fail:
	// problems go into Analysis.Warnings instead.
	Analyze(dir string) Analysis
}

// The Analyzers we have. The order is fixed so the output is always the
// same. To support a new language, add its Analyzer here.
var analyzers = []Analyzer{
	nodeAnalyzer{},
	goAnalyzer{},
	pythonAnalyzer{},
}

// Analyze reads the project in dir. For each of the given types it runs the
// matching Analyzer and returns one Analysis per type.
//
// Types that have no Analyzer yet (docker, typescript) are skipped. Results
// come back in the order of the analyzers list, not the order of types.
func Analyze(dir string, types []detect.Type) []Analysis {
	want := make(map[detect.Type]bool, len(types))
	for _, t := range types {
		want[t] = true
	}

	var out []Analysis
	for _, a := range analyzers {
		if want[a.Type()] {
			out = append(out, a.Analyze(dir))
		}
	}
	return out
}

// --- shared helpers --------------------------------------------------------

// fileExists reports whether a file or folder called name exists inside dir.
func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// pickManager decides which package manager (npm, pnpm, yarn...) a project
// uses, given the managers whose lockfiles were found.
//
// Each manager leaves its own lockfile in the project (pnpm-lock.yaml for
// pnpm, yarn.lock for yarn). If we found exactly one, that's the answer.
// If we found several, we can't tell which one the team really uses, so we
// return a warning instead of guessing.
func pickManager(found []string) (manager, warning string) {
	switch len(found) {
	case 0:
		return "", ""
	case 1:
		return found[0], ""
	default:
		return "", fmt.Sprintf(
			"lockfiles for several package managers found (%s); can't tell which one is intended",
			strings.Join(found, ", "))
	}
}

// appendUnique adds s to list, unless it is already there.
//
// Needed because two different lockfiles can mean the same manager
// (bun.lock and bun.lockb are both bun), and it should be listed once.
func appendUnique(list []string, s string) []string {
	for _, x := range list {
		if x == s {
			return list
		}
	}
	return append(list, s)
}
