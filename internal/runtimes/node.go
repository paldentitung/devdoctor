package runtimes

import "devdoctor/internal/checks"

type Node struct{}

func (Node) Run() checks.Result {
    path, ok := checks.FindInPath("node")
    if !ok {
        return checks.Failed("Node.js", "Node.js is not installed or not available in PATH.", "Install Node.js and ensure it is available in PATH.")
    }

    version, err := checks.GetVersion(path)
    if err != nil {
        return checks.Warning("Node.js", "Node.js was found but its version could not be retrieved.", "Check your Node.js installation.")
    }

    return checks.Passed("Node.js", "Node.js is installed: "+version)
}