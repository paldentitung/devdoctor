package registry

import (
    "devdoctor/internal/checks"
    "devdoctor/internal/packagemanager"
    "devdoctor/internal/runtimes"
    "devdoctor/internal/tools"
)

func Default() []checks.Group {
    return []checks.Group{
        {Name: "Runtimes", Checks: []checks.Check{
            runtimes.Node{},
        }},
        {Name: "Package managers", Checks: []checks.Check{
            packagemanager.Npm{},
        }},
        {Name: "Tools", Checks: []checks.Check{
            tools.Git{},
        }},
    }
}