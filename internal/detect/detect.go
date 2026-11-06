package detect

import (
	"fmt"
	"os"
	"strings"
)

type Type string

const (
	Node       Type = "node"
	TypeScript Type = "typescript"
	Python     Type = "python"
	Go         Type = "go"
	Docker     Type = "docker"
)

// Ordered slice so output is deterministic.
var markers = []struct {
	Type  Type
	Files []string
}{
	{Node, []string{"package.json"}},
	{TypeScript, []string{"tsconfig.json"}},
	{Python, []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"}},
	{Go, []string{"go.mod"}},
	{Docker, []string{"Dockerfile", "docker-compose.yml", "docker-compose.yaml", "compose.yaml"}},
}

// Detect returns every project type whose marker file exists in root.
func Detect(dir string) []Type {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Real on-disk names, so "Go.mod" never equals "go.mod",
	// even on case-insensitive filesystems.
	present := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			present[e.Name()] = true
		}
	}

	var found []Type
	for _, m := range markers {
		for _, name := range m.Files {
			if present[name] {
				found = append(found, m.Type)
				break
			}
		}
	}
	return found
}
// Parse turns "node,python" (from --type) into types, rejecting unknown names.
func Parse(s string) ([]Type, error) {
	var out []Type
	for _, part := range strings.Split(s, ",") {
		name := Type(strings.ToLower(strings.TrimSpace(part)))
		if name == "" {
			continue
		}
		ok := false
		for _, m := range markers {
			if m.Type == name {
				ok = true
				break
			}
		}
		if !ok {
			return nil, fmt.Errorf("unknown project type %q", name)
		}
		out = append(out, name)
	}
	return out, nil
}

func Strings(types []Type) []string {
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = string(t)
	}
	return out
}
