package analyze

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/paldentitung/devdoctor/internal/detect"
)

// project writes the given files (name -> content) into a fresh temp dir.
func project(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func names(a Analysis) []string { return a.FrameworkNames() }

func evidenceOf(a Analysis, name string) string {
	for _, f := range a.Frameworks {
		if f.Name == name {
			return f.Evidence
		}
	}
	return ""
}

func hasWarning(a Analysis, substr string) bool {
	for _, w := range a.Warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

// --- Node ------------------------------------------------------------------

func TestNodeNextApp(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{
			"dependencies": {"next": "^14.1.0", "react": "^18.2.0"},
			"devDependencies": {"typescript": "^5.0.0"},
			"scripts": {"dev": "next dev", "build": "next build"},
			"engines": {"node": ">=18.17"},
			"packageManager": "pnpm@9.1.0"
		}`,
		"next.config.mjs": "export default {}",
		"pnpm-lock.yaml":  "",
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if got, want := names(a), []string{"next", "react"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
	if got, want := evidenceOf(a, "next"), "package.json (dependencies): next ^14.1.0"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
	if a.PackageManager != "pnpm" {
		t.Errorf("package manager = %q, want pnpm", a.PackageManager)
	}
	if a.Requires == nil || a.Requires.Raw != ">=18.17" || a.Requires.Tool != "node" {
		t.Errorf("requires = %+v", a.Requires)
	}
	if want := []string{"build", "dev"}; !reflect.DeepEqual(a.Scripts, want) {
		t.Errorf("scripts = %v, want %v", a.Scripts, want)
	}
	if len(a.Warnings) != 0 {
		t.Errorf("unexpected warnings: %v", a.Warnings)
	}
}

func TestNodeNestApi(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{"dependencies": {"@nestjs/core": "^10.0.0", "express": "^4.18.0"}}`,
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if got, want := names(a), []string{"nest", "express"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
}

func TestNodeDevDependencyIsLabelled(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{"devDependencies": {"vite": "^5.0.0"}}`,
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if got, want := evidenceOf(a, "vite"), "package.json (devDependencies): vite ^5.0.0"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
}

func TestNodeDependencyBeatsDevDependency(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{"dependencies": {"react": "^18"}, "devDependencies": {"react": "^17"}}`,
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if got, want := evidenceOf(a, "react"), "package.json (dependencies): react ^18"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
}

func TestNodeConfigFileOnlyEvidence(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json":   `{}`,
		"vite.config.ts": "",
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if got, want := evidenceOf(a, "vite"), "vite.config.ts exists"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
}

func TestNodeMalformedJSONWarnsAndKeepsGoing(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json":   `{"dependencies": {"next": `, // truncated
		"vite.config.js": "",
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if !hasWarning(a, "not valid") {
		t.Errorf("expected an invalid-file warning, got %v", a.Warnings)
	}
	if !a.HasFramework("vite") {
		t.Errorf("config-file detection should still work, got %v", names(a))
	}
}

func TestNodeMissingPackageJSON(t *testing.T) {
	a := nodeAnalyzer{}.Analyze(t.TempDir())

	if !hasWarning(a, "not found") {
		t.Errorf("expected a not-found warning, got %v", a.Warnings)
	}
}

func TestNodeEnginesAsArrayDoesNotBreakParsing(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{"engines": ["node >= 0.10"], "dependencies": {"express": "^4"}}`,
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if !a.HasFramework("express") {
		t.Errorf("express should still be found, got %v (warnings %v)", names(a), a.Warnings)
	}
	if a.Requires != nil {
		t.Errorf("no usable engines.node, but got %+v", a.Requires)
	}
}

func TestNodeSeveralLockfilesWarn(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json":      `{}`,
		"package-lock.json": "",
		"yarn.lock":         "",
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if a.PackageManager != "" {
		t.Errorf("package manager = %q, want empty when ambiguous", a.PackageManager)
	}
	if !hasWarning(a, "lockfiles") {
		t.Errorf("expected a lockfile warning, got %v", a.Warnings)
	}
}

func TestNodeLockfileFallback(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{}`,
		"yarn.lock":    "",
	})

	if got := (nodeAnalyzer{}).Analyze(dir).PackageManager; got != "yarn" {
		t.Errorf("package manager = %q, want yarn", got)
	}
}

func TestNodePackageManagerFieldVsLockfile(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json":      `{"packageManager": "pnpm@9.0.0"}`,
		"package-lock.json": "",
	})

	a := nodeAnalyzer{}.Analyze(dir)

	if a.PackageManager != "pnpm" {
		t.Errorf("package manager = %q, want pnpm (declared)", a.PackageManager)
	}
	if !hasWarning(a, "npm lockfile") {
		t.Errorf("expected a conflict warning, got %v", a.Warnings)
	}
}

// --- Go --------------------------------------------------------------------

func TestGoModule(t *testing.T) {
	dir := project(t, map[string]string{
		"go.mod": `module example.com/app

go 1.22.0

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/stretchr/testify v1.9.0 // a comment
	github.com/labstack/echo/v4 v4.11.0 // indirect
)

require github.com/spf13/cobra v1.8.0
`,
	})

	a := goAnalyzer{}.Analyze(dir)

	// echo is only an indirect requirement, so it must not count.
	if got, want := names(a), []string{"gin", "cobra"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
	if got, want := evidenceOf(a, "gin"), "go.mod (require): github.com/gin-gonic/gin v1.9.1"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
	if a.Requires == nil || a.Requires.Raw != "1.22.0" || a.Requires.Tool != "go" {
		t.Errorf("requires = %+v", a.Requires)
	}
	if a.PackageManager != "go modules" {
		t.Errorf("package manager = %q", a.PackageManager)
	}
}

func TestGoVersionedModulePathMatches(t *testing.T) {
	dir := project(t, map[string]string{
		"go.mod": "module x\n\ngo 1.21\n\nrequire github.com/labstack/echo/v4 v4.11.0\n",
	})

	if a := (goAnalyzer{}).Analyze(dir); !a.HasFramework("echo") {
		t.Errorf("echo/v4 should match echo, got %v", names(a))
	}
}

func TestGoMissingModFile(t *testing.T) {
	a := goAnalyzer{}.Analyze(t.TempDir())

	if !hasWarning(a, "not found") {
		t.Errorf("expected a not-found warning, got %v", a.Warnings)
	}
}

func TestGoModWithWindowsLineEndings(t *testing.T) {
	dir := project(t, map[string]string{
		"go.mod": "module x\r\n\r\ngo 1.22\r\n\r\nrequire (\r\n\tgithub.com/gin-gonic/gin v1.9.1\r\n)\r\n",
	})

	a := goAnalyzer{}.Analyze(dir)

	if !a.HasFramework("gin") || a.Requires == nil || a.Requires.Raw != "1.22" {
		t.Errorf("CRLF go.mod misread: %v, %+v", names(a), a.Requires)
	}
}

// --- Python ----------------------------------------------------------------

func TestPythonRequirementsTxt(t *testing.T) {
	dir := project(t, map[string]string{
		"requirements.txt": `# web
Django>=4.2   # the framework
FastAPI==0.110.0
uvicorn[standard]>=0.20 ; python_version >= "3.9"
-r other.txt
-e .
git+https://github.com/example/thing.git
`,
	})

	a := pythonAnalyzer{}.Analyze(dir)

	if got, want := names(a), []string{"django", "fastapi"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
	if got, want := evidenceOf(a, "django"), "requirements.txt: django >=4.2"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
	if a.PackageManager != "pip" {
		t.Errorf("package manager = %q, want pip", a.PackageManager)
	}
}

func TestPythonPEP621Pyproject(t *testing.T) {
	dir := project(t, map[string]string{
		"pyproject.toml": `[project]
name = "demo"
requires-python = ">=3.10"  # minimum
classifiers = [
    "Programming Language :: Python :: 3",
]
dependencies = [
    "fastapi>=0.100",
    # a comment line
    "uvicorn[standard]>=0.20",
    'flask',
]

[tool.something]
dependencies = ["django"]
`,
		"uv.lock": "",
	})

	a := pythonAnalyzer{}.Analyze(dir)

	// django sits in an unrelated table and must not count.
	if got, want := names(a), []string{"flask", "fastapi"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
	if a.Requires == nil || a.Requires.Raw != ">=3.10" ||
		a.Requires.Source != "pyproject.toml → requires-python" {
		t.Errorf("requires = %+v", a.Requires)
	}
	if a.PackageManager != "uv" {
		t.Errorf("package manager = %q, want uv", a.PackageManager)
	}
}

func TestPythonPoetryPyproject(t *testing.T) {
	dir := project(t, map[string]string{
		"pyproject.toml": `[tool.poetry]
name = "demo"

[tool.poetry.dependencies]
python = "^3.9"
Django = "^4.2"
flask = {version = "^3.0", extras = ["async"]}
`,
	})

	a := pythonAnalyzer{}.Analyze(dir)

	if got, want := names(a), []string{"django", "flask"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
	if got, want := evidenceOf(a, "flask"), "pyproject.toml (poetry): flask ^3.0"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
	if a.Requires == nil || a.Requires.Raw != "^3.9" {
		t.Errorf("requires = %+v", a.Requires)
	}
	if a.PackageManager != "poetry" {
		t.Errorf("package manager = %q, want poetry", a.PackageManager)
	}
}

func TestPythonDjangoByManagePy(t *testing.T) {
	dir := project(t, map[string]string{
		"requirements.txt": "requests\n",
		"manage.py":        "",
	})

	a := pythonAnalyzer{}.Analyze(dir)

	if got, want := evidenceOf(a, "django"), "manage.py exists"; got != want {
		t.Errorf("evidence = %q, want %q", got, want)
	}
}

func TestPythonNamesAreNormalised(t *testing.T) {
	name, _, ok := parseRequirement("Flask_SQLAlchemy==3.1.1")
	if !ok || name != "flask-sqlalchemy" {
		t.Errorf("name = %q, ok = %v", name, ok)
	}
}

// --- orchestration ---------------------------------------------------------

func TestAnalyzeRunsOnlyRequestedTypes(t *testing.T) {
	dir := project(t, map[string]string{
		"package.json": `{}`,
		"go.mod":       "module x\n\ngo 1.22\n",
	})

	got := Analyze(dir, []detect.Type{detect.Go, detect.Node, detect.Docker, detect.TypeScript})

	// Analyzer order, not input order; docker and typescript have no analyzer.
	if len(got) != 2 || got[0].Type != detect.Node || got[1].Type != detect.Go {
		t.Errorf("got %+v", got)
	}

	if none := Analyze(dir, []detect.Type{detect.Docker}); len(none) != 0 {
		t.Errorf("docker alone should produce nothing, got %+v", none)
	}
}

func TestDepSetLookupIsDeterministic(t *testing.T) {
	d := depSet{}
	d.add("github.com/labstack/echo/v5", "v5", "x")
	d.add("github.com/labstack/echo/v4", "v4", "x")

	for i := 0; i < 20; i++ {
		if key, _, _ := d.lookup("github.com/labstack/echo"); key != "github.com/labstack/echo/v4" {
			t.Fatalf("lookup returned %q", key)
		}
	}
}
