package checks

import (
	"regexp"
	"strings"
	"unicode"
)

// opSpaceRe glues an operator to its version, so ">= 18" becomes ">=18".
// That way each rule is a single word and can be split on spaces safely.
var opSpaceRe = regexp.MustCompile(`(>=|<=|==|~=|!=|\^|~|>|<|=)\s+`)

var lowerBoundOps = []string{">=", "==", "~=", "^", "~", "=", ">"}

// ParseMinConstraint extracts the lowest allowed version from a project
// version constraint.
//
// It supports common version formats used by npm, Go and Python, including
// comparison operators (>=18), caret (^18.2), tilde (~3.11), wildcards
// (==3.9.*), plain versions (18, v20.1), and ranges (1.2 - 1.5).
//
// Combined constraints are supported:
//   - Comma or space means AND: the lowest lower bound is used.
//   - || means OR: the lowest available alternative is used.
//
// Only the minimum version matters. Upper bounds, exclusions and strict
// comparison operators are not enforced.
//
// It returns false when no usable lower bound can be determined, allowing
// the caller to skip the check rather than guess.

func ParseMinConstraint(s string) (Version, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, false
	}

	var lowest Version
	have := false

	for _, alt := range strings.Split(s, "||") {
		v, ok := alternativeMin(alt)
		if !ok {
			return Version{}, false
		}
		if !have || !v.AtLeast(lowest) { // v is older than what we have so far
			lowest, have = v, true
		}
	}

	return lowest, have
}

// alternativeMin handles one side of an "||" (one OR choice). Inside it,
// several rules joined by a comma or space must ALL hold, so the answer is
// the highest of their lower limits: ">=3.8 >=3.10" means 3.10 or newer.
func alternativeMin(alt string) (Version, bool) {
	// A range like "1.2 - 1.5": keep only the lower end.
	if i := strings.Index(alt, " - "); i >= 0 {
		alt = alt[:i]
	}

	alt = opSpaceRe.ReplaceAllString(strings.TrimSpace(alt), "$1")
	tokens := strings.FieldsFunc(alt, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})

	var best Version
	have := false

	for _, tok := range tokens {
		if strings.HasPrefix(tok, "<") || strings.HasPrefix(tok, "!=") {
			continue
		}

		for _, op := range lowerBoundOps {
			if strings.HasPrefix(tok, op) {
				tok = strings.TrimPrefix(tok, op)
				break
			}
		}
		tok = strings.TrimPrefix(tok, "v")

		if tok == "" || tok[0] < '0' || tok[0] > '9' {
			continue
		}

		v, ok := ParseVersion(tok)
		if !ok {
			continue
		}
		if !have || v.AtLeast(best) {
			best, have = v, true
		}
	}

	return best, have
}
