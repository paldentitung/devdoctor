package runtimes

import "devdoctor/internal/checks"

var minNode = checks.Version{Major: 22}

type Node struct {
	Min *checks.Version // the project's requirement; nil means use the default
}

func (n Node) Run() checks.Result {
	path, ok := checks.FindInPath("node")
	if !ok {
		return checks.Failed("Node.js", "Node.js is not installed or not available in PATH.", "Install Node.js and ensure it is available in PATH.")
	}

	version, err := checks.GetVersion(path)
	if err != nil {
		return checks.Warning("Node.js", "Node.js was found but its version could not be retrieved.", "Check your Node.js installation.")
	}

	min := minNode
	if n.Min != nil {
		min = *n.Min
	}

	return checks.MinVersion("Node.js", version, min, "https://nodejs.org")

}
