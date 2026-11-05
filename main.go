package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"devdoctor/internal/registry"
	"devdoctor/internal/runner"
	"devdoctor/internal/system"
)

// version is overridden at build time:
//
//	go build -ldflags "-X main.version=0.1.0"
var version = "dev"

func main() {
	jsonOutput := flag.Bool("json", false, "print results as JSON")
	noColor := flag.Bool("no-color", false, "disable colored output")
	showVersion := flag.Bool("version", false, "print devdoctor's version and exit")
	only := flag.String("only", "", "run only the group with this name (e.g. runtimes)")
	flag.Parse()

	if *showVersion {
		fmt.Println("devdoctor", version)
		return
	}

	useColor = colorEnabled(*noColor)

	results := runner.Run(registry.Default())

	if *only != "" {
		results = filterGroups(results, *only)
		if len(results) == 0 {
			fmt.Fprintf(os.Stderr, "no group matches %q\n", *only)
			os.Exit(2)
		}
	}

	summary := runner.Summarize(results)

	if *jsonOutput {
		printJSON(results)
	} else {
		printSystemInfo()
		printResults(results)
		printSummary(summary.Passed, summary.Warnings, summary.Failed)
	}

	if summary.Failed > 0 {
		os.Exit(1)
	}
}

// filterGroups keeps only groups whose name matches (case-insensitive).
func filterGroups(groups []runner.GroupResult, name string) []runner.GroupResult {
	var out []runner.GroupResult
	for _, g := range groups {
		if strings.EqualFold(g.Name, name) {
			out = append(out, g)
		}
	}
	return out
}

func printJSON(results []runner.GroupResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		fmt.Fprintln(os.Stderr, "failed to encode JSON:", err)
		os.Exit(1)
	}
}

func printSystemInfo() {
	info := system.GetInfo()

	fmt.Println("DevDoctor - System Information")
	fmt.Println("------------------------------")
	fmt.Println("Operating System:", info.OS)
	fmt.Println("Architecture:", info.Architecture)
	fmt.Println("CPU Cores:", info.CPUCores)
	fmt.Println("Go Version:", info.GoVersion)
	fmt.Println("Home Directory:", info.HomeDirectory)
	fmt.Println("Username:", info.Username)
	fmt.Println("Shell:", info.Shell)
	fmt.Println()
}