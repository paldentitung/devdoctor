package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"devdoctor/internal/detect"
	"devdoctor/internal/registry"
	"devdoctor/internal/runner"
	"devdoctor/internal/system"
	"devdoctor/internal/analyze"
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
	typeFlag := flag.String("type", "", "comma-separated project types to check instead of auto-detecting (node,typescript,python,go,docker)")
	flag.Parse()

	if *showVersion {
		fmt.Println("devdoctor", version)
		return
	}

	useColor = colorEnabled(*noColor)

	use, detected, err := resolveTypes(*typeFlag)
	analyses := analyze.Analyze(".", use)		
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	results := runner.Run(runner.Select(registry.Default(analyses), use))

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
		printDetected(use, detected, *typeFlag != "")
		printAnalysis(analyses)
		printResults(results)
		printSummary(summary.Passed, summary.Warnings, summary.Failed)
	}

	if summary.Failed > 0 {
		os.Exit(1)
	}
}

// resolveTypes uses the --type flag if given, otherwise auto-detects
// from marker files in the current directory.
func resolveTypes(flagValue string) (use []detect.Type, detected []detect.Type, err error) {
	detected = detect.Detect(".")
	if flagValue == "" {
		return detected, detected, nil
	}
	use, err = detect.Parse(flagValue)
	return use, detected, err
}

// printAnalysis shows what was read from the project's own files.
func printAnalysis(analyses []analyze.Analysis) {
	for _, a := range analyses {
		fmt.Printf("%s project:\n", a.Type)
		for _, f := range a.Frameworks {
			fmt.Printf("  Framework:       %s  (%s)\n", f.Name, f.Evidence)
		}
		if a.PackageManager != "" {
			fmt.Printf("  Package manager: %s\n", a.PackageManager)
		}
		if a.Requires != nil {
			fmt.Printf("  Requires:        %s %s  (%s)\n", a.Requires.Tool, a.Requires.Raw, a.Requires.Source)
		}
		for _, w := range a.Warnings {
			fmt.Printf("  Warning:         %s\n", w)
		}
	}
	if len(analyses) > 0 {
		fmt.Println()
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

// printDetected tells the user which project types were used to pick checks.
func printDetected(use, detected []detect.Type, forced bool) {
	detectedStr := strings.Join(detect.Strings(detected), ", ")
	if detectedStr == "" {
		detectedStr = "none"
	}

	checksStr := strings.Join(detect.Strings(use), ", ")
	switch {
	case forced:
		checksStr += " (override via --type)"
	case checksStr == "":
		checksStr = "general checks only (use --type to choose)"
	}

	fmt.Println("Detected project:", detectedStr)
	fmt.Println("Checks for:      ", checksStr)
	fmt.Println()
}
