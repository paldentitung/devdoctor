package runtimes

import "devdoctor/internal/checks"


var minNode = checks.Version{Major: 22}

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

	have, ok := checks.ParseVersion(version)
	if ok && !have.AtLeast(minNode) {
		return checks.Warning(
			"Node.js",
			"Node.js "+version+" is older than the required "+minNode.String()+".",
			"Upgrade Node.js to version "+minNode.String()+" or newer: https://nodejs.org",
		)
	}

	return checks.Passed("Node.js", "Node.js is installed: "+version)
}