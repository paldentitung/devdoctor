package registry

import (
	"devdoctor/internal/analyze"
	"devdoctor/internal/checks"
	"devdoctor/internal/database"
	"devdoctor/internal/detect"
	"devdoctor/internal/packagemanager"
	"devdoctor/internal/project"
	"devdoctor/internal/runtimes"
	"devdoctor/internal/tools"
)

func Default(analyses []analyze.Analysis) []checks.Group {
	nodeLike := []detect.Type{detect.Node, detect.TypeScript}

	return []checks.Group{
		{Name: "Runtimes", Checks: []checks.Check{
			checks.For(runtimes.Node{}, nodeLike...),
			checks.For(runtimes.Golang{}, detect.Go),
			checks.For(runtimes.Python{}, detect.Python),
		}},
		{Name: "Package managers", Checks: []checks.Check{
			checks.For(packagemanager.Npm{}, nodeLike...),
		}},
		{Name: "Tools", Checks: []checks.Check{
			tools.Git{},
			checks.For(tools.Docker{}, detect.Docker),
			tools.Curl{},
		}},
		{Name: "Databases", Checks: []checks.Check{
			database.Postgres{},
			database.MongoDB{},
		}},
		{Name: "Project", Checks: []checks.Check{
			checks.For(project.PackageJSON{}, nodeLike...),
			project.Port{Number: defaultPort(analyses)},
		}},
	}
}
