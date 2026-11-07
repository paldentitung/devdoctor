package checks

import "testing"

func TestParseMinConstraint(t *testing.T) {
	tests := []struct {
		in   string
		want Version
		ok   bool
	}{
		// single bounds
		{">=18", Version{18, 0, 0}, true},
		{">= 18.2.1", Version{18, 2, 1}, true},
		{"^18.19.0", Version{18, 19, 0}, true},
		{"~3.11", Version{3, 11, 0}, true},
		{"~=3.9", Version{3, 9, 0}, true},
		{"==3.9.*", Version{3, 9, 0}, true},
		{"18", Version{18, 0, 0}, true},
		{"18.x", Version{18, 0, 0}, true},
		{"v20.1", Version{20, 1, 0}, true},
		{">3.8", Version{3, 8, 0}, true},
		{">=1.2.3-rc.1", Version{1, 2, 3}, true},
		{"1.22.0", Version{1, 22, 0}, true},

		// AND: comma or space, upper bounds ignored
		{">=3.9,<4", Version{3, 9, 0}, true},
		{">=18 <21", Version{18, 0, 0}, true},
		{">=3.9,!=3.10.*", Version{3, 9, 0}, true},
		{">=3.8 >=3.10", Version{3, 10, 0}, true}, // highest lower bound

		// OR: lowest alternative wins
		{"^18 || >=20", Version{18, 0, 0}, true},
		{">=20 || ^18", Version{18, 0, 0}, true},

		// hyphen range
		{"1.2 - 1.5", Version{1, 2, 0}, true},

		// no usable lower bound -> false, so callers pass rather than guess
		{"", Version{}, false},
		{"   ", Version{}, false},
		{"<20", Version{}, false},
		{"*", Version{}, false},
		{"x", Version{}, false},
		{"latest", Version{}, false},
		{"^18 || *", Version{}, false},
		{"workspace:*", Version{}, false},
	}

	for _, tt := range tests {
		got, ok := ParseMinConstraint(tt.in)
		if ok != tt.ok || got != tt.want {
			t.Errorf("ParseMinConstraint(%q) = %v, %v; want %v, %v",
				tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
