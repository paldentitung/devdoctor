package detect

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// touch creates empty marker files inside dir.
func touch(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  []Type
	}{
		{"empty folder", nil, nil},
		{"unrelated files only", []string{"README.md", "main.c"}, nil},
		{"node", []string{"package.json"}, []Type{Node}},
		{"typescript only", []string{"tsconfig.json"}, []Type{TypeScript}},
		{"node + typescript", []string{"package.json", "tsconfig.json"}, []Type{Node, TypeScript}},
		{"python requirements", []string{"requirements.txt"}, []Type{Python}},
		{"python pyproject", []string{"pyproject.toml"}, []Type{Python}},
		{"python setup.py", []string{"setup.py"}, []Type{Python}},
		{"python pipfile", []string{"Pipfile"}, []Type{Python}},
		{"python with several markers is reported once", []string{"requirements.txt", "pyproject.toml"}, []Type{Python}},
		{"go", []string{"go.mod"}, []Type{Go}},
		{"docker Dockerfile", []string{"Dockerfile"}, []Type{Docker}},
		{"docker compose yml", []string{"docker-compose.yml"}, []Type{Docker}},
		{"docker compose yaml", []string{"docker-compose.yaml"}, []Type{Docker}},
		{"docker compose.yaml", []string{"compose.yaml"}, []Type{Docker}},
		{"mixed, result follows marker order", []string{"Dockerfile", "requirements.txt", "package.json"}, []Type{Node, Python, Docker}},
		{"everything", []string{"package.json", "tsconfig.json", "pyproject.toml", "go.mod", "Dockerfile"},
			[]Type{Node, TypeScript, Python, Go, Docker}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			touch(t, dir, tt.files...)

			got := Detect(dir)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectMissingDirectory(t *testing.T) {
	got := Detect(filepath.Join(t.TempDir(), "does-not-exist"))
	if got != nil {
		t.Fatalf("Detect() on a missing dir = %v, want nil", got)
	}
}

func TestDetectDoesNotLookInSubfolders(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "app")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, sub, "package.json")

	if got := Detect(dir); got != nil {
		t.Fatalf("Detect() = %v, want nil (markers are only read from root)", got)
	}
}

func TestDetectMarkerMustBeNamedExactly(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json.bak", "Go.mod", "dockerfile.old")

	if got := Detect(dir); got != nil {
		t.Fatalf("Detect() = %v, want nil", got)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Type
	}{
		{"single", "node", []Type{Node}},
		{"multiple", "node,python", []Type{Node, Python}},
		{"keeps the order given", "docker,go,node", []Type{Docker, Go, Node}},
		{"uppercase", "NODE", []Type{Node}},
		{"mixed case", "TypeScript", []Type{TypeScript}},
		{"spaces around names", " node , python ", []Type{Node, Python}},
		{"empty entries are skipped", "node,,python,", []Type{Node, Python}},
		{"all types", "node,typescript,python,go,docker", []Type{Node, TypeScript, Python, Go, Docker}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unknown type", "banana"},
		{"one bad among good", "node,banana"},
		{"partial name", "pyth"},
		{"path-like value", "./node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err == nil {
				t.Fatalf("Parse(%q) = %v, want an error", tt.input, got)
			}
			if got != nil {
				t.Fatalf("Parse(%q) returned %v alongside the error, want nil", tt.input, got)
			}
		})
	}
}

func TestStrings(t *testing.T) {
	if got := Strings(nil); len(got) != 0 {
		t.Fatalf("Strings(nil) = %v, want empty", got)
	}

	got := Strings([]Type{Node, TypeScript, Go})
	want := []string{"node", "typescript", "go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Strings() = %v, want %v", got, want)
	}
}
