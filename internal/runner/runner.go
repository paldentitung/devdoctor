// Package runner runs checks and returns their results.
// It only produces data; printing is left to the caller (main.go).
package runner

import "github.com/paldentitung/devdoctor/internal/checks"
import "github.com/paldentitung/devdoctor/internal/detect"

// GroupResult holds the results of one group of checks.
type GroupResult struct {
	Name    string          `json:"name"`
	Results []checks.Result `json:"results"`
}

// Summary counts results by status.
type Summary struct {
	Passed   int
	Warnings int
	Failed   int
}

// Run executes every check in every group, in order.
// It does not stop when a check fails.
func Run(groups []checks.Group) []GroupResult {
	out := make([]GroupResult, 0, len(groups))

	for _, g := range groups {
		gr := GroupResult{Name: g.Name}

		for _, c := range g.Checks {
			gr.Results = append(gr.Results, c.Run())
		}

		out = append(out, gr)
	}

	return out
}

// Summarize counts how many checks passed, warned or failed.
func Summarize(results []GroupResult) Summary {
	var s Summary

	for _, g := range results {
		for _, r := range g.Results {
			switch r.Status {
			case checks.StatusPassed:
				s.Passed++
			case checks.StatusWarning:
				s.Warnings++
			case checks.StatusFailed:
				s.Failed++
			}
		}
	}

	return s
}

// Select drops checks that don't apply to the detected types, and drops
// groups left empty. Checks that aren't Scoped always stay.
func Select(groups []checks.Group, detected []detect.Type) []checks.Group {
	have := make(map[detect.Type]bool, len(detected))
	for _, t := range detected {
		have[t] = true
	}

	var out []checks.Group
	for _, g := range groups {
		ng := checks.Group{Name: g.Name}
		for _, c := range g.Checks {
			if s, ok := c.(checks.Scoped); ok && !matches(s.Types, have) {
				continue
			}
			ng.Checks = append(ng.Checks, c)
		}
		if len(ng.Checks) > 0 {
			out = append(out, ng)
		}
	}
	return out
}

func matches(types []detect.Type, have map[detect.Type]bool) bool {
	for _, t := range types {
		if have[t] {
			return true
		}
	}
	return false
}
