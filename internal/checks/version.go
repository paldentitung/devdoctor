package checks

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version is a parsed major.minor.patch version. Missing parts are 0.
type Version struct {
	Major, Minor, Patch int
}

var versionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// ParseVersion finds the first version number in a string, so
// "v18.19.0", "Python 3.12.1" and "git version 2.43.0" all work.
func ParseVersion(s string) (Version, bool) {
	m := versionRe.FindStringSubmatch(s)
	if m == nil {
		return Version{}, false
	}
	var v Version
	v.Major, _ = strconv.Atoi(m[1])
	v.Minor, _ = strconv.Atoi(m[2]) // empty string -> 0
	v.Patch, _ = strconv.Atoi(m[3])
	return v, true
}

// AtLeast reports whether v is the same as or newer than min.
// It compares numbers, so 18.10 is correctly newer than 18.9.
func (v Version) AtLeast(min Version) bool {
	if v.Major != min.Major {
		return v.Major > min.Major
	}
	if v.Minor != min.Minor {
		return v.Minor > min.Minor
	}
	return v.Patch >= min.Patch
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// MinVersion builds the final result for a tool whose version was read
// successfully: a Warning if it is older than min, otherwise Passed.
// If the version can't be parsed, it passes rather than guessing.
func MinVersion(name, raw string, min Version, url string) Result {
	have, ok := ParseVersion(raw)
	if ok && !have.AtLeast(min) {
		return Warning(
			name,
			fmt.Sprintf("%s %s is older than the required %s.", name, have, min),
			fmt.Sprintf("Upgrade %s to %s or newer: %s", name, min, url),
		)
	}
	return Passed(name, name+" is installed: "+raw)
}
