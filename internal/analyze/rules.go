package analyze

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A dep is one dependency (a library the project uses) and where it was
// declared.
//
// Example: Version "^14.1.0", Where "package.json (dependencies)".
type dep struct {
	Version string
	Where   string // which file, and which part of it
}

// evidence builds the proof text shown to the user, for example
// "package.json (dependencies): next ^14.1.0".
func (d dep) evidence(name string) string {
	s := d.Where + ": " + name
	if d.Version != "" {
		s += " " + d.Version
	}
	return s
}

// A depSet is the project's whole dependency list: each library name
// (like "next") mapped to how it was declared.
type depSet map[string]dep

// add records a dependency, unless one with that name is already recorded.
//
// Callers add the most trustworthy source first (real dependencies before
// devDependencies), so when a name appears twice, the first one is kept.
func (d depSet) add(name, version, where string) {
	if _, ok := d[name]; !ok {
		d[name] = dep{Version: version, Where: where}
	}
}

// lookup finds a dependency by name and returns the name it found, how it
// was declared, and whether it found anything.
//
// Besides an exact match it also accepts the name followed by a path, so
// looking up "github.com/labstack/echo" finds "github.com/labstack/echo/v4".
// (Go adds /v4, /v5... to a module's name when it has major versions.)
func (d depSet) lookup(name string) (string, dep, bool) {
	if v, ok := d[name]; ok {
		return name, v, true
	}

	prefix := name + "/"
	var keys []string
	for k := range d {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return "", dep{}, false
	}
	sort.Strings(keys) // Go maps have no fixed order; sorting keeps the output the same every run
	return keys[0], d[keys[0]], true
}

// A rule describes how to recognise ONE framework.
//
// Example: {Name: "next", Deps: ["next"], Files: ["next.config.*"]} means
// "report Next.js if the project depends on next, or if a file named
// next.config.js, next.config.mjs, ... exists".
//
// Dependencies are checked before files, because a dependency is stronger
// proof than a config file.
type rule struct {
	Name  string   // what to call the framework in the output
	Deps  []string // library names that prove it
	Files []string // file-name patterns that prove it ("*" matches any text)
}

// match tests every rule against the project and returns a Finding for each
// rule that applies, in the order the rules are listed.
//
// That order matters for the reader: put the more specific frameworks first
// (next before react). A Next.js app also uses React, but "next" is the more
// useful thing to see first.
func match(dir string, rules []rule, deps depSet) []Finding {
	files := listFiles(dir)

	var out []Finding
	for _, r := range rules {
		if ev := evidence(r, files, deps); ev != "" {
			out = append(out, Finding{Name: r.Name, Evidence: ev})
		}
	}
	return out
}

// evidence returns the proof that a rule applies, or "" if it doesn't.
// It tries the dependencies first, then the config files.
func evidence(r rule, files []string, deps depSet) string {
	for _, name := range r.Deps {
		if key, d, ok := deps.lookup(name); ok {
			return d.evidence(key)
		}
	}
	for _, pattern := range r.Files {
		for _, f := range files {
			if ok, _ := filepath.Match(pattern, f); ok {
				return f + " exists"
			}
		}
	}
	return ""
}

// listFiles returns the names of the files (not folders) in dir, in
// alphabetical order. If the folder can't be read, it returns nothing.
func listFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}
