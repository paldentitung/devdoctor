package analyze

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"devdoctor/internal/detect"
)

// How to recognise Python frameworks and libraries. Names are written the way
// normalizePyName produces them: lowercase, with dashes.
var pythonRules = []rule{
	{Name: "django", Deps: []string{"django"}, Files: []string{"manage.py"}},
	{Name: "flask", Deps: []string{"flask"}},
	{Name: "fastapi", Deps: []string{"fastapi"}},
	{Name: "starlette", Deps: []string{"starlette"}},
	{Name: "celery", Deps: []string{"celery"}},
	{Name: "sqlalchemy", Deps: []string{"sqlalchemy"}},
}

// Python tools each leave their own file in the project, which tells us which
// tool manages it. Pipfile counts even without its lock file. The order here
// doesn't matter: every file that exists is collected.
var pythonLockfiles = []struct{ File, Manager string }{
	{"uv.lock", "uv"},
	{"poetry.lock", "poetry"},
	{"pdm.lock", "pdm"},
	{"Pipfile.lock", "pipenv"},
	{"Pipfile", "pipenv"},
}

// pythonAnalyzer reads pyproject.toml and requirements.txt, the two common
// places where a Python project lists its libraries.
type pythonAnalyzer struct{}

func (pythonAnalyzer) Type() detect.Type { return detect.Python }

// Analyze reads the Python files in dir. Both files are optional: a project
// may have one, both, or neither.
func (pythonAnalyzer) Analyze(dir string) Analysis {
	a := Analysis{Type: detect.Python}
	deps := depSet{}

	// pyproject.toml is the modern settings file. It can list libraries and
	// the Python version the project asks for.
	var py pyproject
	if data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml")); err == nil {
		py = parsePyproject(string(data))

		for _, spec := range py.Deps {
			if name, version, ok := parseRequirement(spec); ok {
				deps.add(name, version, "pyproject.toml (dependencies)")
			}
		}
		for name, version := range py.PoetryDeps {
			deps.add(name, version, "pyproject.toml (poetry)")
		}

		switch {
		case py.RequiresPython != "":
			a.Requires = &Constraint{Tool: "python", Raw: py.RequiresPython,
				Source: "pyproject.toml → requires-python"}
		case py.PoetryPython != "":
			a.Requires = &Constraint{Tool: "python", Raw: py.PoetryPython,
				Source: "pyproject.toml → tool.poetry.dependencies.python"}
		}
	}

	// requirements.txt is the older, simpler way: one library per line.
	hasRequirements := false
	if data, err := os.ReadFile(filepath.Join(dir, "requirements.txt")); err == nil {
		hasRequirements = true
		for _, line := range strings.Split(string(data), "\n") {
			if name, version, ok := parseRequirement(line); ok {
				deps.add(name, version, "requirements.txt")
			}
		}
	}

	a.Frameworks = match(dir, pythonRules, deps)

	// Which tool manages the project? A tool's own file is the best clue.
	// If there is none, fall back on hints from the other files.
	var locks []string
	for _, l := range pythonLockfiles {
		if fileExists(dir, l.File) {
			locks = appendUnique(locks, l.Manager)
		}
	}
	pm, warn := pickManager(locks)
	switch {
	case warn != "":
		a.Warnings = append(a.Warnings, warn)
	case pm != "":
		a.PackageManager = pm
	case py.UsesPoetry:
		a.PackageManager = "poetry"
	case hasRequirements:
		a.PackageManager = "pip"
	}

	return a
}

// --- requirements ----------------------------------------------------------

// Splits a requirement line into three parts: the library name, optional
// extras in [brackets], and the rest (the version rule).
//
//	"Django[argon2]>=4.2"  ->  "Django", "[argon2]", ">=4.2"
var requirementRe = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)\s*(?:\[[^\]]*\])?\s*(.*)$`)

// parseRequirement reads one line of a requirements list and returns the
// library's name and its version rule.
//
//	"Django[argon2]>=4.2 ; python_version >= '3.9'"  ->  "django", ">=4.2"
//	"requests"                                       ->  "requests", ""
//
// ok is false for lines that are not libraries: blank lines, comments,
// options such as "-r other.txt", and links to repositories.
func parseRequirement(line string) (name, version string, ok bool) {
	if i := strings.Index(line, " #"); i >= 0 { // cut off a trailing comment: "flask  # web"
		line = line[:i]
	}
	if i := strings.Index(line, ";"); i >= 0 { // cut off a condition: ; python_version >= "3.9"
		line = line[:i]
	}
	line = strings.TrimSpace(line)
	if line == "" || line[0] == '#' || line[0] == '-' {
		return "", "", false
	}

	m := requirementRe.FindStringSubmatch(line)
	if m == nil {
		return "", "", false
	}

	// What follows the name must look like a version rule (>=1.0, ==2, ~=3).
	// This rejects lines such as "git+https://github.com/x/y.git", which
	// merely START with something that looks like a name ("git").
	rest := strings.TrimSpace(m[2])
	if rest != "" && !strings.ContainsRune("<>=!~@(", rune(rest[0])) {
		return "", "", false
	}

	return normalizePyName(m[1]), rest, true
}

// normalizePyName puts a library name in one standard form.
//
// Python treats upper and lower case, "_", "." and "-" as the same thing in
// library names (this is the PEP 503 rule), so "Flask_SQLAlchemy" and
// "flask-sqlalchemy" are the same library. We always use lowercase with dashes.
func normalizePyName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return strings.ReplaceAll(s, ".", "-")
}

// --- pyproject.toml --------------------------------------------------------

// pyproject holds the few things we read from pyproject.toml.
type pyproject struct {
	RequiresPython string            // the Python version the project asks for: [project] requires-python
	Deps           []string          // the library list in [project] dependencies, one requirement per entry
	PoetryDeps     map[string]string // libraries in [tool.poetry.dependencies], name -> version
	PoetryPython   string            // the Python version in the poetry section: python = "..."
	UsesPoetry     bool              // true if the file has any [tool.poetry...] section
}

var (
	// Finds text inside "double" or 'single' quotes.
	quotedRe = regexp.MustCompile(`"([^"]*)"|'([^']*)'`)

	// Finds version = "..." inside a poetry inline table such as
	// {version = "^3.0", extras = ["async"]}.
	inlineVersionRe = regexp.MustCompile(`version\s*=\s*["']([^"']*)["']`)
)

// parsePyproject reads the parts of pyproject.toml that we need.
//
// pyproject.toml is written in TOML, a settings format made of [sections]
// and key = value lines, for example:
//
//	[project]
//	requires-python = ">=3.10"
//	dependencies = ["fastapi>=0.100", "uvicorn[standard]"]
//
// This is NOT a full TOML parser. It follows sections, key = value lines,
// quoted text, multi-line lists and comments, and ignores everything else.
func parsePyproject(text string) pyproject {
	p := pyproject{PoetryDeps: map[string]string{}}
	lines := strings.Split(text, "\n")
	table := "" // the [section] we are currently inside

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(stripTomlComment(lines[i]))
		if line == "" {
			continue
		}

		// A section heading such as [project] or [tool.poetry.dependencies].
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			table = strings.TrimSpace(strings.Trim(line, "[]"))
			if table == "tool.poetry" || strings.HasPrefix(table, "tool.poetry.") {
				p.UsesPoetry = true
			}
			continue
		}

		// A "key = value" line. A list value may continue over several lines,
		// in which case readValue joins them and moves i forward.
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.Trim(strings.TrimSpace(key), `"'`)
		val, i = readValue(lines, i, strings.TrimSpace(val))

		switch table {
		case "project":
			switch key {
			case "requires-python":
				p.RequiresPython = unquote(val)
			case "dependencies":
				p.Deps = append(p.Deps, arrayStrings(val)...)
			}
		case "tool.poetry.dependencies":
			if key == "python" {
				p.PoetryPython = unquote(val)
			} else {
				p.PoetryDeps[normalizePyName(key)] = poetryVersion(val)
			}
		}
	}

	return p
}

// readValue returns the full value that starts on lines[i].
//
// Most values fit on one line. A list can run over several:
//
//	dependencies = [
//	    "fastapi>=0.100",
//	    "uvicorn",
//	]
//
// In that case the lines are joined into one string, and the returned index
// is the last line that was used, so the caller can skip past it.
func readValue(lines []string, i int, first string) (string, int) {
	if !strings.HasPrefix(first, "[") {
		return first, i
	}

	buf := first
	for arrayEnd(buf) < 0 && i+1 < len(lines) {
		i++
		buf += "\n" + stripTomlComment(lines[i])
	}
	return buf, i
}

// arrayEnd returns the position of the "]" that closes the list starting at
// the beginning of s, or -1 if the list isn't closed yet.
//
// Quoted text is skipped, so the "]" in "uvicorn[standard]" is not mistaken
// for the end of the list.
func arrayEnd(s string) int {
	var quote rune
	depth := 0
	for i, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '[':
			depth++
		case r == ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// arrayStrings returns the quoted texts inside a list value:
// `["a", 'b']` gives ["a", "b"].
func arrayStrings(val string) []string {
	if end := arrayEnd(val); end >= 0 {
		val = val[:end]
	}

	var out []string
	for _, m := range quotedRe.FindAllStringSubmatch(val, -1) {
		if m[1] != "" {
			out = append(out, m[1])
		} else if m[2] != "" {
			out = append(out, m[2])
		}
	}
	return out
}

// poetryVersion reads the version of a poetry dependency. Poetry allows two
// styles: plain text ("^4.2"), or a small table ({version = "^3.0", extras = [...]}).
func poetryVersion(val string) string {
	if strings.HasPrefix(val, "{") {
		if m := inlineVersionRe.FindStringSubmatch(val); m != nil {
			return m[1]
		}
		return ""
	}
	return unquote(val)
}

// unquote removes one pair of matching quotes around s: `">=3.9"` gives >=3.9.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

// stripTomlComment removes a trailing "# comment" from a line. A "#" inside
// quotes is part of the text, not a comment, so it is left alone.
func stripTomlComment(s string) string {
	var quote rune
	for i, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '#':
			return s[:i]
		}
	}
	return s
}
