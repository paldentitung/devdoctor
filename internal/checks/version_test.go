package checks

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		in   string
		want Version
		ok   bool
	}{
		{"v18.19.0", Version{18, 19, 0}, true},
		{"Python 3.12.1", Version{3, 12, 1}, true},
		{"git version 2.43.0", Version{2, 43, 0}, true},
		{"go version go1.22.1 linux/amd64", Version{1, 22, 1}, true},
		{"no digits here", Version{}, false},
	}
	for _, tt := range tests {
		got, ok := ParseVersion(tt.in)
		if ok != tt.ok || got != tt.want {
			t.Errorf("ParseVersion(%q) = %v, %v; want %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestAtLeast(t *testing.T) {
	if !(Version{18, 10, 0}).AtLeast(Version{18, 9, 0}) {
		t.Error("18.10 should be at least 18.9")
	}
	if (Version{16, 20, 0}).AtLeast(Version{18, 0, 0}) {
		t.Error("16.20 should not be at least 18.0")
	}
	if !(Version{18, 0, 0}).AtLeast(Version{18, 0, 0}) {
		t.Error("equal versions should pass")
	}
}
