package registry

import (
	"devdoctor/internal/analyze"
	"devdoctor/internal/checks"
)

const fallbackPort = 3000


var frameworkPorts = map[string]int{
	"next":      3000,
	"nuxt":      3000,
	"nest":      3000,
	"vite":      5173,
	"sveltekit": 5173,
	"angular":   4200,
	"astro":     4321,
	"django":    8000,
	"fastapi":   8000,
	"flask":     5000,
	"gin":       8080,
}

// defaultPort returns the port of the first detected framework that has
// a known default, or 3000.
func defaultPort(analyses []analyze.Analysis) int {
	for _, a := range analyses {
		for _, f := range a.Frameworks {
			if p, ok := frameworkPorts[f.Name]; ok {
				return p
			}
		}
	}
	return fallbackPort
}

// requiredMin returns the lowest version the project asks for the given
// tool ("node", "go", "python"), or nil if it says nothing usable.
func requiredMin(analyses []analyze.Analysis, tool string) *checks.Version {
	for _, a := range analyses {
		if a.Requires == nil || a.Requires.Tool != tool {
			continue
		}
		if v, ok := checks.ParseMinConstraint(a.Requires.Raw); ok {
			return &v
		}
	}
	return nil
}