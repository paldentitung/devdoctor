package registry

import (
	"devdoctor/internal/checks"
	"devdoctor/internal/database"
	"devdoctor/internal/packagemanager"
	"devdoctor/internal/project"
	"devdoctor/internal/runtimes"
	"devdoctor/internal/tools"
)

func Default() []checks.Group {
	return []checks.Group{
		{Name: "Runtimes", Checks: []checks.Check{
			runtimes.Node{},
			runtimes.Golang{},
			runtimes.Python{},
		}},
		{Name: "Package managers", Checks: []checks.Check{
			packagemanager.Npm{},
		}},
		{Name: "Tools", Checks: []checks.Check{
			tools.Git{},
			tools.Docker{},
			tools.Curl{},
		}},
		{Name: "Databases", Checks: []checks.Check{
		database.Postgres{},
		database.MongoDB{},
		}},
		{Name: "Project", Checks: []checks.Check{
			project.PackageJSON{},
			project.Port{Number: 3000},
		}},
	}
}
