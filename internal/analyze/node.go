package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/paldentitung/devdoctor/internal/detect"
)

// How to recognise Node frameworks. This is also the order they are
// reported in, so the biggest and most specific come first: Next.js before
// React (a Next.js app also uses React), then plain libraries, then build
// tools, then server libraries.
var nodeRules = []rule{
	{Name: "next", Deps: []string{"next"}, Files: []string{"next.config.*"}},
	{Name: "nuxt", Deps: []string{"nuxt", "nuxt3"}, Files: []string{"nuxt.config.*"}},
	{Name: "nest", Deps: []string{"@nestjs/core"}, Files: []string{"nest-cli.json"}},
	{Name: "angular", Deps: []string{"@angular/core"}, Files: []string{"angular.json"}},
	{Name: "sveltekit", Deps: []string{"@sveltejs/kit"}},
	{Name: "remix", Deps: []string{"@remix-run/react", "@remix-run/node"}},
	{Name: "astro", Deps: []string{"astro"}, Files: []string{"astro.config.*"}},
	{Name: "react", Deps: []string{"react"}},
	{Name: "vue", Deps: []string{"vue"}},
	{Name: "svelte", Deps: []string{"svelte"}},
	{Name: "vite", Deps: []string{"vite"}, Files: []string{"vite.config.*"}},
	{Name: "express", Deps: []string{"express"}},
	{Name: "fastify", Deps: []string{"fastify"}},
}

// Each package manager leaves its own lockfile in the project, so seeing one
// tells us which manager installed the project's libraries. Bun has used two
// file names over time, so it appears twice. The order here doesn't matter:
// every file that exists is collected.
var nodeLockfiles = []struct{ File, Manager string }{
	{"pnpm-lock.yaml", "pnpm"},
	{"yarn.lock", "yarn"},
	{"package-lock.json", "npm"},
	{"bun.lockb", "bun"},
	{"bun.lock", "bun"},
}

// packageJSON holds only the parts of package.json that we read.
// Everything else in the file is ignored.
type packageJSON struct {
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Scripts          map[string]string `json:"scripts"`
	PackageManager   string            `json:"packageManager"`

	// Some old packages write "engines" as a list instead of an object.
	// Reading that straight into a map would make the WHOLE file fail to
	// load, so it is kept raw here and decoded separately, forgivingly.
	Engines json.RawMessage `json:"engines"`
}

// nodeAnalyzer reads package.json, and the lockfiles that sit next to it.
type nodeAnalyzer struct{}

func (nodeAnalyzer) Type() detect.Type { return detect.Node }

// Analyze reads package.json in dir. If the file is missing or broken, that
// becomes a warning, and we still look for config files next to it (such as
// vite.config.ts), which can show the framework even without package.json.
func (nodeAnalyzer) Analyze(dir string) Analysis {
	a := Analysis{Type: detect.Node}

	var p packageJSON
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		a.Warnings = append(a.Warnings, "package.json not found")
	case err != nil:
		a.Warnings = append(a.Warnings, "package.json could not be read: "+err.Error())
	default:
		if err := json.Unmarshal(data, &p); err != nil {
			a.Warnings = append(a.Warnings, "package.json is not valid: "+err.Error())
		}
	}

	// Collect every library the project declares. Libraries needed to run the
	// app count for more than devDependencies (tools used only while
	// building), so they go in first. If a name appears twice, the first
	// one is kept.
	deps := depSet{}
	for name, v := range p.Dependencies {
		deps.add(name, v, "package.json (dependencies)")
	}
	for name, v := range p.PeerDependencies {
		deps.add(name, v, "package.json (peerDependencies)")
	}
	for name, v := range p.DevDependencies {
		deps.add(name, v, "package.json (devDependencies)")
	}
	a.Frameworks = match(dir, nodeRules, deps)

	// The Node version the project asks for, e.g. "engines": {"node": ">=18"}.
	var engines map[string]string
	_ = json.Unmarshal(p.Engines, &engines) // if "engines" isn't an object, engines stays empty: no requirement found
	if raw := strings.TrimSpace(engines["node"]); raw != "" {
		a.Requires = &Constraint{Tool: "node", Raw: raw, Source: "package.json → engines.node"}
	}

	// The names of the scripts (dev, build, test...), sorted so the output
	// is the same every run.
	for name := range p.Scripts {
		a.Scripts = append(a.Scripts, name)
	}
	sort.Strings(a.Scripts)

	var warns []string
	a.PackageManager, warns = nodePackageManager(dir, p.PackageManager)
	a.Warnings = append(a.Warnings, warns...)

	return a
}

// nodePackageManager works out which package manager the project uses.
//
// The best source is the "packageManager" field in package.json: for example
// "pnpm@9.1.0" means pnpm. Without it, we look at which lockfile exists.
//
// It also warns when the two disagree, for example package.json says pnpm
// but a package-lock.json (from npm) is sitting in the folder.
func nodePackageManager(dir, declared string) (string, []string) {
	var locks []string
	for _, l := range nodeLockfiles {
		if fileExists(dir, l.File) {
			locks = appendUnique(locks, l.Manager)
		}
	}

	if declared != "" {
		name, _, _ := strings.Cut(declared, "@") // "pnpm@9.1.0" -> "pnpm"
		var warns []string
		for _, l := range locks {
			if l != name {
				warns = append(warns, fmt.Sprintf(
					"package.json says %s but a %s lockfile exists", name, l))
			}
		}
		return name, warns
	}

	pm, warn := pickManager(locks)
	if warn != "" {
		return pm, []string{warn}
	}
	return pm, nil
}
