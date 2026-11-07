package analyze

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"devdoctor/internal/detect"
)

// How to recognise Go frameworks and libraries. A Go module is named by its
// import address, like "github.com/gin-gonic/gin". Version endings such as
// /v4 are handled automatically (see depSet.lookup), so they are left out here.
var goRules = []rule{
	{Name: "gin", Deps: []string{"github.com/gin-gonic/gin"}},
	{Name: "echo", Deps: []string{"github.com/labstack/echo"}},
	{Name: "fiber", Deps: []string{"github.com/gofiber/fiber"}},
	{Name: "chi", Deps: []string{"github.com/go-chi/chi"}},
	{Name: "gorilla-mux", Deps: []string{"github.com/gorilla/mux"}},
	{Name: "grpc", Deps: []string{"google.golang.org/grpc"}},
	{Name: "cobra", Deps: []string{"github.com/spf13/cobra"}},
	{Name: "gorm", Deps: []string{"gorm.io/gorm"}},
}

// goAnalyzer reads go.mod, the file that describes a Go project.
type goAnalyzer struct{}

func (goAnalyzer) Type() detect.Type { return detect.Go }

// Analyze reads go.mod in dir and takes two things from it: the Go version
// the project asks for (the "go 1.22" line) and the libraries it requires.
func (goAnalyzer) Analyze(dir string) Analysis {
	a := Analysis{Type: detect.Go, PackageManager: "go modules"}

	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			a.Warnings = append(a.Warnings, "go.mod not found")
		} else {
			a.Warnings = append(a.Warnings, "go.mod could not be read: "+err.Error())
		}
		return a
	}

	// go.mod looks like this:
	//
	//	module example.com/app
	//
	//	go 1.22
	//
	//	require (
	//		github.com/gin-gonic/gin v1.9.1
	//		github.com/some/lib v1.0.0 // indirect
	//	)
	//
	//	require github.com/spf13/cobra v1.8.0
	//
	// inRequire is true while we are inside a "require ( ... )" block.
	deps := depSet{}
	inRequire := false

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		// Remember whether the line ends in "// indirect" (a library that
		// only another library needs), then cut off any comment.
		indirect := strings.Contains(line, "// indirect")
		if i := strings.Index(line, "//"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}

		switch {
		case inRequire:
			if line == ")" {
				inRequire = false
				continue
			}
			addGoDep(deps, line, indirect)
		case line == "require (" || line == "require(":
			inRequire = true
		case strings.HasPrefix(line, "require "):
			addGoDep(deps, strings.TrimPrefix(line, "require "), indirect)
		case strings.HasPrefix(line, "go "):
			a.Requires = &Constraint{
				Tool:   "go",
				Raw:    strings.TrimSpace(strings.TrimPrefix(line, "go ")),
				Source: "go.mod → go",
			}
		}
	}

	a.Frameworks = match(dir, goRules, deps)
	return a
}

// addGoDep records one requirement line of the form "module version",
// such as "github.com/gin-gonic/gin v1.9.1".
//
// Indirect requirements are skipped: the project doesn't use those libraries
// itself, so they say little about what kind of project it is.
func addGoDep(deps depSet, line string, indirect bool) {
	if indirect {
		return
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	deps.add(fields[0], fields[1], "go.mod (require)")
}
