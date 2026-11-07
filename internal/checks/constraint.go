package checks

import (
	"regexp"
	"strings"
	"unicode"
)

// opSpaceRe glues an operator to its version, so ">= 18" becomes ">=18".
// That way each rule is a single word and can be split on spaces safely.
var opSpaceRe = regexp.MustCompile(`(>=|<=|==|~=|!=|\^|~|>|<|=)\s+`)

// lowerBoundOps are the operators that name the LOWEST allowed version
// (">=18" means "18 or newer"). Longer operators come first so that ">="
// is not mistaken for ">".
var lowerBoundOps = []string{">=", "==", "~=", "^", "~", "=", ">"}

// ParseMinConstraint answers one question about a project's version rule:
// "what is the lowest version it allows?"
//
// Projects write these rules in different styles, and this understands the
// common ones from npm, Go and Python:
//
//	>=18          18 or newer                          -> 18.0.0
//	^18.2         18.2 up to (not including) 19        -> 18.2.0
//	~3.11         3.11 and its patch releases          -> 3.11.0
//	~=3.9         Python's version of the same         -> 3.9.0
//	==3.9.*       any 3.9                              -> 3.9.0
//	18  18.x  v20.1                                    -> 18.0.0, 18.0.0, 20.1.0
//
// Rules can be combined:
//
//	>=3.9,<4      comma or space means AND: take the lower bound  -> 3.9.0
//	^18 || >=20   || means OR: the lowest choice wins             -> 18.0.0
//	1.2 - 1.5     a range: take the lower end                     -> 1.2.0
//
// Only the lowest allowed version is read. Upper limits ("<4"), exclusions
// ("!=3.10") and strictness (">3.8" is treated as 3.8) are ignored.
//
// If any part of the rule has no usable lower limit (such as "*", "latest"
// or "<20"), it returns false. The caller should then skip the check rather
// than guess.
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
