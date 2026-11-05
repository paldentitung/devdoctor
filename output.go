package main

import (
	"fmt"
	"os"

	"devdoctor/internal/checks"
	"devdoctor/internal/runner"
)

// useColor is set once in main, after flags are parsed.
var useColor bool

// colorEnabled turns color off for --no-color, NO_COLOR, TERM=dumb,
// and when stdout is not a terminal (piped or redirected).
func colorEnabled(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// paint colors s with the given ANSI code, or returns it unchanged when color is off.
func paint(code, s string) string {
	if !useColor {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func printResults(groups []runner.GroupResult) {
	for _, g := range groups {
		fmt.Println(g.Name)
		fmt.Println("------------------------------")

		for _, r := range g.Results {
			var icon string
			switch r.Status {
			case checks.StatusPassed:
				icon = paint("32", "✔")
			case checks.StatusWarning:
				icon = paint("33", "!")
			default:
				icon = paint("31", "✘")
			}

			fmt.Printf("%s %s\n", icon, r.Message)

			// Suggestions only matter when something is wrong.
			if r.Status != checks.StatusPassed && r.Suggestion != "" {
				fmt.Printf("    → %s\n", r.Suggestion)
			}
		}
		fmt.Println()
	}
}

func printSummary(passed, warnings, failed int) {
	fmt.Printf("%s  %s  %s\n",
		paint("32", fmt.Sprintf("Passed: %d", passed)),
		paint("33", fmt.Sprintf("Warnings: %d", warnings)),
		paint("31", fmt.Sprintf("Failed: %d", failed)),
	)
}
