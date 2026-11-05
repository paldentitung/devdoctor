// internal/runner/runner_test.go
package runner

import (
	"testing"

	"devdoctor/internal/checks"
)

type fakeCheck struct{ result checks.Result }

func (f fakeCheck) Run() checks.Result { return f.result }

func TestRunAndSummarize(t *testing.T) {
	groups := []checks.Group{
		{Name: "Test", Checks: []checks.Check{
			fakeCheck{checks.Passed("a", "ok")},
			fakeCheck{checks.Warning("b", "meh", "fix")},
			fakeCheck{checks.Failed("c", "bad", "fix")},
		}},
	}

	results := Run(groups)
	if len(results) != 1 || len(results[0].Results) != 3 {
		t.Fatalf("unexpected results: %+v", results)
	}

	s := Summarize(results)
	if s.Passed != 1 || s.Warnings != 1 || s.Failed != 1 {
		t.Errorf("unexpected summary: %+v", s)
	}
}
